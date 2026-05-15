package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tuotuoxp/go-jsonschema/internal/x/text"
	"github.com/tuotuoxp/go-jsonschema/pkg/schemas"
)

func TestResolveStructFieldSchemaTypeKeepsDereferencedCacheState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mainSchemaPath := filepath.Join(dir, "main.json")
	namedSchemaPath := filepath.Join(dir, "named.schema")

	require.NoError(t, os.WriteFile(namedSchemaPath, []byte(`{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/named-ref",
  "title": "NamedRef",
  "type": "string",
  "pattern": "^[a-z]{3}$"
}`), 0o600))

	require.NoError(t, os.WriteFile(mainSchemaPath, []byte(`{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/main-ref",
  "type": "object",
  "properties": {
    "name": {
      "$ref": "./named.schema"
    }
  }
}`), 0o600))

	cfg := Config{
		SchemaMappings:     []SchemaMapping{},
		ExtraImports:       true,
		DefaultPackageName: "github.com/example/test",
		DefaultOutputName:  "-",
		ResolveExtensions:  []string{".json", ".schema"},
		YAMLExtensions:     []string{".yaml", ".yml"},
		Tags:               []string{"json", "yaml", "mapstructure"},
		Warner:             func(string) {},
	}

	gen, err := New(cfg)
	require.NoError(t, err)

	schema, err := gen.loader.Load(mainSchemaPath, "")
	require.NoError(t, err)

	output, err := gen.findOutputFileForSchemaID(schema.ID)
	require.NoError(t, err)

	sg := newSchemaGenerator(gen, schema, mainSchemaPath, output)

	prop := schema.Properties["name"]
	_, semanticInline := sg.resolveStructFieldSchemaType(prop)
	require.False(t, semanticInline)

	cached := sg.schemaTypesByRef[prop.Ref]
	require.NotNil(t, cached)
	require.True(t, cached.Dereferenced)
}

func TestDoFileRegeneratesPreviouslyReferencedSchemaAsRootTarget(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	consumerSchemaPath := filepath.Join(dir, "consumer.schema")
	sharedSchemaPath := filepath.Join(dir, "shared.schema")

	require.NoError(t, os.WriteFile(sharedSchemaPath, []byte(`{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared",
  "type": "object",
  "$defs": {
    "Marker": {
      "type": "string"
    },
    "User": {
      "type": "object",
      "x-go-ref": {
        "path": "github.com/example/shared",
        "alias": "shared"
      },
      "properties": {
        "id": { "type": "string" }
      }
    }
  }
}`), 0o600))

	require.NoError(t, os.WriteFile(consumerSchemaPath, []byte(`{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer",
  "type": "object",
  "properties": {
    "shared": {
      "$ref": "./shared.schema#/$defs/Marker"
    }
  }
}`), 0o600))

	cfg := Config{
		SchemaMappings: []SchemaMapping{
			{
				SchemaID:    "https://example.com/consumer",
				OutputName:  "consumer.go",
				PackageName: "testpkg",
			},
			{
				SchemaID:    "https://example.com/shared",
				OutputName:  "shared.go",
				PackageName: "testpkg",
			},
		},
		ExtraImports:       true,
		DefaultPackageName: "testpkg",
		DefaultOutputName:  "default.go",
		ResolveExtensions:  []string{".json", ".schema"},
		YAMLExtensions:     []string{".yaml", ".yml"},
		Tags:               []string{"json", "yaml", "mapstructure"},
		Warner:             func(string) {},
	}

	gen, err := New(cfg)
	require.NoError(t, err)

	require.NoError(t, gen.DoFile(consumerSchemaPath))
	require.NoError(t, gen.DoFile(sharedSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	sharedSource, ok := sources["shared.go"]
	require.True(t, ok)
	require.True(t, strings.Contains(string(sharedSource), "type User struct"))
}

func TestResolveReferencedDefinitionTypeNameUsesFallbackByDefault(t *testing.T) {
	t.Parallel()

	sg := &schemaGenerator{
		Generator: &Generator{
			caser: text.NewCaser(nil, nil),
		},
	}
	definition := &schemas.Type{
		XGoRef: &schemas.XGoRefExtension{
			Path:  "github.com/example/shared",
			Alias: "shared",
		},
	}

	got, err := sg.resolveReferencedDefinitionTypeName(definition, "User")
	require.NoError(t, err)
	require.Equal(t, "User", got)
}

func TestResolveReferencedDefinitionTypeNameUsesTitleWhenEnabled(t *testing.T) {
	t.Parallel()

	sg := &schemaGenerator{
		Generator: &Generator{
			config: Config{
				StructNameFromTitle: true,
			},
			caser: text.NewCaser(nil, nil),
		},
	}

	definition := &schemas.Type{
		Title: "shared user title",
	}

	got, err := sg.resolveReferencedDefinitionTypeName(definition, "SharedUser")
	require.NoError(t, err)
	require.Equal(t, "SharedUserTitle", got)
}

func TestResolveReferencedXGoRefMappingUsesFallbackName(t *testing.T) {
	t.Parallel()

	sg := &schemaGenerator{
		Generator: &Generator{
			caser: text.NewCaser(nil, nil),
		},
	}

	definition := &schemas.Type{
		XGoRef: &schemas.XGoRefExtension{
			Path:  "github.com/example/shared",
			Alias: "shared",
		},
	}

	mappedType, importPath, importAlias, ok, err := sg.resolveReferencedXGoRefMapping(
		definition,
		"User",
		"#/$defs/User",
	)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "shared.User", mappedType)
	require.Equal(t, "github.com/example/shared", importPath)
	require.Equal(t, "shared", importAlias)
}

func TestResolveReferencedXGoRefMappingRequiresPath(t *testing.T) {
	t.Parallel()

	sg := &schemaGenerator{
		Generator: &Generator{
			caser: text.NewCaser(nil, nil),
		},
	}

	definition := &schemas.Type{
		XGoRef: &schemas.XGoRefExtension{
			Alias: "shared",
		},
	}

	_, _, _, ok, err := sg.resolveReferencedXGoRefMapping(definition, "User", "#/$defs/User")
	require.False(t, ok)
	require.ErrorContains(t, err, "x-go-ref.path is required")
}

func TestResolveReferencedXGoRefMappingRequiresAlias(t *testing.T) {
	t.Parallel()

	sg := &schemaGenerator{
		Generator: &Generator{
			caser: text.NewCaser(nil, nil),
		},
	}

	definition := &schemas.Type{
		XGoRef: &schemas.XGoRefExtension{
			Path: "github.com/example/shared",
		},
	}

	_, _, _, ok, err := sg.resolveReferencedXGoRefMapping(definition, "User", "#/$defs/User")
	require.False(t, ok)
	require.ErrorContains(t, err, "x-go-ref.alias is required")
}

func TestResolveReferencedXGoRefMappingRejectsInvalidAlias(t *testing.T) {
	t.Parallel()

	sg := &schemaGenerator{
		Generator: &Generator{
			caser: text.NewCaser(nil, nil),
		},
	}

	definition := &schemas.Type{
		XGoRef: &schemas.XGoRefExtension{
			Path:  "github.com/example/shared",
			Alias: "1shared",
		},
	}

	_, _, _, ok, err := sg.resolveReferencedXGoRefMapping(definition, "User", "#/$defs/User")
	require.False(t, ok)
	require.ErrorContains(t, err, "must be a valid Go identifier")
}
