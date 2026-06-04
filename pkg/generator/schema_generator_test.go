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

func TestGenerateReferencedRootSchemaUsesXGoRefMapping(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	consumerSchemaPath := filepath.Join(dir, "consumer.schema")
	statusSchemaPath := filepath.Join(dir, "status.schema")

	writeSchemaFile(t, statusSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/status",
  "title": "StatusSchema",
  "type": "string",
  "x-go-ref": {
    "path": "github.com/example/shared",
    "alias": "shared"
  }
}`)

	writeSchemaFile(t, consumerSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer/root-ref",
  "type": "object",
  "properties": {
    "status": {
      "$ref": "./status.schema"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/consumer/root-ref",
			OutputName:  "consumer.go",
			PackageName: "testpkg",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(consumerSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	consumerSource, ok := sources["consumer.go"]
	require.True(t, ok)

	generated := string(consumerSource)
	require.Contains(t, generated, `import shared "github.com/example/shared"`)
	require.Contains(t, generated, "*shared.StatusSchema")
	require.NotContains(t, generated, "type StatusSchema string")
}

func TestGenerateReferencedDefinitionXGoRefStillWorks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	consumerSchemaPath := filepath.Join(dir, "consumer.schema")
	sharedSchemaPath := filepath.Join(dir, "shared.schema")

	writeSchemaFile(t, sharedSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/defs",
  "$defs": {
    "User": {
      "type": "object",
      "x-go-ref": {
        "path": "github.com/example/shared",
        "alias": "shared"
      },
      "properties": {
        "id": {
          "type": "string"
        }
      }
    }
  }
}`)

	writeSchemaFile(t, consumerSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer/def-ref",
  "type": "object",
  "properties": {
    "user": {
      "$ref": "./shared.schema#/$defs/User"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/consumer/def-ref",
			OutputName:  "consumer.go",
			PackageName: "testpkg",
		},
	)

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(consumerSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	consumerSource, ok := sources["consumer.go"]
	require.True(t, ok)

	generated := string(consumerSource)
	require.Contains(t, generated, "*shared.User")
	require.NotContains(t, generated, "type User struct")
}

func TestGenerateRootSchemaWithXGoRefStillGeneratesLocalDeclaration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statusSchemaPath := filepath.Join(dir, "status.schema")

	writeSchemaFile(t, statusSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/status",
  "title": "StatusSchema",
  "type": "string",
  "x-go-ref": {
    "path": "github.com/example/shared",
    "alias": "shared"
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/shared/status",
			OutputName:  "shared.go",
			PackageName: "testpkg",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(statusSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	sharedSource, ok := sources["shared.go"]
	require.True(t, ok)

	generated := string(sharedSource)
	require.Contains(t, generated, "type StatusSchema string")
	require.NotContains(t, generated, "import shared ")
}

func TestGenerateReferencedRootSchemaXGoRefValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		xGoRefJSON string
		wantErr    string
	}{
		{
			name:       "missing path",
			xGoRefJSON: `"x-go-ref": { "alias": "shared" }`,
			wantErr:    "x-go-ref.path is required",
		},
		{
			name:       "missing alias",
			xGoRefJSON: `"x-go-ref": { "path": "github.com/example/shared" }`,
			wantErr:    "x-go-ref.alias is required",
		},
		{
			name:       "invalid alias",
			xGoRefJSON: `"x-go-ref": { "path": "github.com/example/shared", "alias": "1shared" }`,
			wantErr:    "must be a valid Go identifier",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			consumerSchemaPath := filepath.Join(dir, "consumer.schema")
			statusSchemaPath := filepath.Join(dir, "status.schema")

			writeSchemaFile(t, statusSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/status",
  "title": "StatusSchema",
  "type": "string",
  `+tc.xGoRefJSON+`
}`)

			writeSchemaFile(t, consumerSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer/root-ref-error",
  "type": "object",
  "properties": {
    "status": {
      "$ref": "./status.schema"
    }
  }
}`)

			cfg := testConfigWithMappings(
				SchemaMapping{
					SchemaID:    "https://example.com/consumer/root-ref-error",
					OutputName:  "consumer.go",
					PackageName: "testpkg",
				},
			)
			cfg.StructNameFromTitle = true

			gen, err := New(cfg)
			require.NoError(t, err)
			require.ErrorContains(t, gen.DoFile(consumerSchemaPath), tc.wantErr)
		})
	}
}

func TestGenerateReferencedRootSchemaWithoutXGoRefKeepsBehavior(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	consumerSchemaPath := filepath.Join(dir, "consumer.schema")
	statusSchemaPath := filepath.Join(dir, "status.schema")

	writeSchemaFile(t, statusSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/status/no-ref",
  "title": "StatusSchema",
  "type": "string"
}`)

	writeSchemaFile(t, consumerSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer/root-ref-no-x-go-ref",
  "type": "object",
  "properties": {
    "status": {
      "$ref": "./status.schema"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/consumer/root-ref-no-x-go-ref",
			OutputName:  "consumer.go",
			PackageName: "testpkg",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(consumerSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	consumerSource, ok := sources["consumer.go"]
	require.True(t, ok)

	consumerGenerated := string(consumerSource)
	require.Contains(t, consumerGenerated, "*StatusSchema")
	require.NotContains(t, consumerGenerated, "shared.StatusSchema")

	hasLocalStatusType := false
	for _, source := range sources {
		if strings.Contains(string(source), "type StatusSchema string") {
			hasLocalStatusType = true
			break
		}
	}

	require.True(t, hasLocalStatusType)
}

// TestGenerateXGoRefSamePackageRootRefNoSelfImport verifies that when schema A
// references schema B and both have x-go-ref pointing at the same Go package,
// and A is being generated into that same package, the generator does NOT
// produce a self-import and does NOT emit a qualified "alias.TypeName" reference.
// Instead it should use the local (unqualified) type name directly.
func TestGenerateXGoRefSamePackageRootRefNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")
	bSchemaPath := filepath.Join(dir, "b.schema")

	writeSchemaFile(t, bSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/b",
  "title": "BType",
  "type": "string",
  "x-go-ref": {
    "path": "github.com/example/shared",
    "alias": "shared"
  }
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/a",
  "title": "AType",
  "type": "object",
  "x-go-ref": {
    "path": "github.com/example/shared",
    "alias": "shared"
  },
  "properties": {
    "b": {
      "$ref": "./b.schema"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/shared/a",
			OutputName:  "shared.go",
			PackageName: "github.com/example/shared",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	sharedSource, ok := sources["shared.go"]
	require.True(t, ok)

	generated := string(sharedSource)
	// Must NOT self-import the package.
	require.NotContains(t, generated, `import shared "github.com/example/shared"`)
	// Must NOT use a qualified same-package type reference.
	require.NotContains(t, generated, "shared.BType")
	// Must use the unqualified local type name.
	require.Contains(t, generated, "BType")
}

// TestGenerateXGoRefSamePackageDefinitionRefNoSelfImport verifies the
// same-package guard for a definition-level $ref (shared.schema#/$defs/User).
func TestGenerateXGoRefSamePackageDefinitionRefNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")
	sharedSchemaPath := filepath.Join(dir, "shared.schema")

	writeSchemaFile(t, sharedSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/defs",
  "$defs": {
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
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/a",
  "title": "AType",
  "type": "object",
  "properties": {
    "user": {
      "$ref": "./shared.schema#/$defs/User"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/shared/a",
			OutputName:  "shared.go",
			PackageName: "github.com/example/shared",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	sharedSource, ok := sources["shared.go"]
	require.True(t, ok)

	generated := string(sharedSource)
	// Must NOT self-import the package.
	require.NotContains(t, generated, `import shared "github.com/example/shared"`)
	// Must NOT use a qualified same-package type reference.
	require.NotContains(t, generated, "shared.User")
	// Must use the unqualified local type name.
	require.Contains(t, generated, "User")
}

// TestGenerateXGoRefSamePackageRootRefObjectTypeNoSelfImport is the exact user
// repro scenario: schema A (object type) references schema B (object type), both
// declare x-go-ref pointing at the same Go package, and A is generated into that
// package.  The B field inside A must use the unqualified local type name and must
// NOT produce a self-import.
func TestGenerateXGoRefSamePackageRootRefObjectTypeNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "A.schema.json")
	bSchemaPath := filepath.Join(dir, "B.schema.json")

	writeSchemaFile(t, bSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/b",
  "title": "ObjectB",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "job_id": { "type": "string" }
  }
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/a",
  "title": "ObjectA",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "B": {
      "$ref": "./B.schema.json"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/shared/a",
			OutputName:  "openapi.go",
			PackageName: "myproject/openapi",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["openapi.go"]
	require.True(t, ok)

	generated := string(src)
	// Must NOT self-import the package.
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	// Must NOT use a qualified same-package type reference.
	require.NotContains(t, generated, "oapi.ObjectB")
	// Must use the unqualified local type name in the field type.
	require.Contains(t, generated, "B *ObjectB")
}

// TestGenerateXGoRefSamePackageRootRefObjectTypeNoIDNoSelfImport verifies the
// same-package x-go-ref behavior for the user repro where schemas omit id/$id.
func TestGenerateXGoRefSamePackageRootRefObjectTypeNoIDNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "A.schema.json")
	bSchemaPath := filepath.Join(dir, "B.schema.json")

	writeSchemaFile(t, bSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ObjectB",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "job_id": { "type": "string" }
  }
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ObjectA",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "B": {
      "$ref": "./B.schema.json"
    }
  }
}`)

	cfg := testConfigWithMappings()
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(bSchemaPath))
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["default.go"]
	require.True(t, ok)

	generated := string(src)
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	require.NotContains(t, generated, "oapi.ObjectB")
	require.Contains(t, generated, "B *ObjectB")
}

// TestGenerateXGoRefSamePackageExternalDefsRefObjectTypeNoSelfImport verifies the
// same-package guard for an external $defs ref where the definition is an object
// type and the caller package matches the x-go-ref path.
func TestGenerateXGoRefSamePackageExternalDefsRefObjectTypeNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")
	sharedSchemaPath := filepath.Join(dir, "shared.schema")

	writeSchemaFile(t, sharedSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/defs",
  "$defs": {
    "Order": {
      "type": "object",
      "x-go-ref": {
        "path": "myproject/openapi",
        "alias": "oapi"
      },
      "properties": {
        "id": { "type": "string" }
      }
    }
  }
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/a",
  "title": "ObjectA",
  "type": "object",
  "properties": {
    "order": {
      "$ref": "./shared.schema#/$defs/Order"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/shared/a",
			OutputName:  "openapi.go",
			PackageName: "myproject/openapi",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["openapi.go"]
	require.True(t, ok)

	generated := string(src)
	// Must NOT self-import the package.
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	// Must NOT use a qualified same-package type reference.
	require.NotContains(t, generated, "oapi.Order")
	// Must use the unqualified local type name in the field type.
	require.Contains(t, generated, "Order *Order")
}

// TestGenerateXGoRefSamePackageExternalDefsRefObjectTypeNoIDNoSelfImport verifies
// same-package detection for an external $defs ref when schemas have no id/$id.
func TestGenerateXGoRefSamePackageExternalDefsRefObjectTypeNoIDNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")
	sharedSchemaPath := filepath.Join(dir, "shared.schema")

	writeSchemaFile(t, sharedSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$defs": {
    "Order": {
      "type": "object",
      "x-go-ref": {
        "path": "myproject/openapi",
        "alias": "oapi"
      },
      "properties": {
        "id": { "type": "string" }
      }
    }
  }
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ObjectA",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "order": {
      "$ref": "./shared.schema#/$defs/Order"
    }
  }
}`)

	cfg := testConfigWithMappings()
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(sharedSchemaPath))
	require.NoError(t, gen.DoFile(aSchemaPath))
	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["default.go"]
	require.True(t, ok)

	generated := string(src)
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	require.NotContains(t, generated, "oapi.Order")
	require.Contains(t, generated, "Order *Order")
}

// TestGenerateXGoRefSameFileSamePackageDefsRefProducesLocalType verifies that a
// same-file $defs reference (i.e. "#/$defs/User") where the definition carries an
// x-go-ref pointing at the current output package resolves as a local named type
// without any self-import.
func TestGenerateXGoRefSameFileSamePackageDefsRefProducesLocalType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/a",
  "title": "ObjectA",
  "type": "object",
  "$defs": {
    "User": {
      "type": "object",
      "x-go-ref": {
        "path": "myproject/openapi",
        "alias": "oapi"
      },
      "properties": {
        "id": { "type": "string" }
      }
    }
  },
  "properties": {
    "user": {
      "$ref": "#/$defs/User"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/shared/a",
			OutputName:  "openapi.go",
			PackageName: "myproject/openapi",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["openapi.go"]
	require.True(t, ok)

	generated := string(src)
	// Must NOT self-import the package.
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	// Must NOT use a qualified reference.
	require.NotContains(t, generated, "oapi.User")
	// Must use User as a local (unqualified) type in the field.
	require.Contains(t, generated, "User *User")
}

// TestGenerateXGoRefSameFileSamePackageDefsRefNoIDProducesLocalType verifies
// same-file $defs behavior without id/$id on the root schema.
func TestGenerateXGoRefSameFileSamePackageDefsRefNoIDProducesLocalType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ObjectA",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "$defs": {
    "User": {
      "type": "object",
      "x-go-ref": {
        "path": "myproject/openapi",
        "alias": "oapi"
      },
      "properties": {
        "id": { "type": "string" }
      }
    }
  },
  "properties": {
    "user": {
      "$ref": "#/$defs/User"
    }
  }
}`)

	cfg := testConfigWithMappings()
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(aSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["default.go"]
	require.True(t, ok)

	generated := string(src)
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	require.NotContains(t, generated, "oapi.User")
	require.Contains(t, generated, "User *User")
}

// TestGenerateXGoRefSamePackageInlineRefNoIDNoSelfImport verifies same-package
// detection for the inline path (array items -> $ref) without id/$id.
func TestGenerateXGoRefSamePackageInlineRefNoIDNoSelfImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	aSchemaPath := filepath.Join(dir, "a.schema")
	bSchemaPath := filepath.Join(dir, "b.schema")

	writeSchemaFile(t, bSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ObjectB",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "job_id": { "type": "string" }
  }
}`)

	writeSchemaFile(t, aSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ObjectA",
  "x-go-ref": {
    "path": "myproject/openapi",
    "alias": "oapi"
  },
  "type": "object",
  "properties": {
    "items": {
      "type": "array",
      "items": {
        "$ref": "./b.schema"
      }
    }
  }
}`)

	cfg := testConfigWithMappings()
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(bSchemaPath))
	require.NoError(t, gen.DoFile(aSchemaPath))
	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["default.go"]
	require.True(t, ok)

	generated := string(src)
	require.NotContains(t, generated, `import oapi "myproject/openapi"`)
	require.NotContains(t, generated, "[]oapi.ObjectB")
	require.Contains(t, generated, "Items []ObjectB")
}

// TestGenerateXGoRefExternalPackageRootRefObjectTypeStillImports confirms that
// when the x-go-ref path differs from the output package, the generator still
// emits a qualified alias.TypeName and the required import.  This is the
// canonical external-package use case and must continue to work.
func TestGenerateXGoRefExternalPackageRootRefObjectTypeStillImports(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	consumerSchemaPath := filepath.Join(dir, "consumer.schema")
	bSchemaPath := filepath.Join(dir, "b.schema")

	writeSchemaFile(t, bSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/b",
  "title": "ObjectB",
  "x-go-ref": {
    "path": "github.com/example/shared",
    "alias": "shared"
  },
  "type": "object",
  "properties": {
    "job_id": { "type": "string" }
  }
}`)

	writeSchemaFile(t, consumerSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer",
  "title": "Consumer",
  "type": "object",
  "properties": {
    "B": {
      "$ref": "./b.schema"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/consumer",
			OutputName:  "consumer.go",
			PackageName: "testpkg",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(consumerSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["consumer.go"]
	require.True(t, ok)

	generated := string(src)
	// Must import the external package.
	require.Contains(t, generated, `shared "github.com/example/shared"`)
	// Must use the qualified type name.
	require.Contains(t, generated, "shared.ObjectB")
	// Must NOT generate a local ObjectB struct.
	require.NotContains(t, generated, "type ObjectB struct")
}

// TestGenerateXGoRefExternalPackageDefsRefObjectTypeStillImports confirms that
// an external $defs ref to an object type with x-go-ref still emits a qualified
// alias.TypeName and import when the packages differ.
func TestGenerateXGoRefExternalPackageDefsRefObjectTypeStillImports(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	consumerSchemaPath := filepath.Join(dir, "consumer.schema")
	sharedSchemaPath := filepath.Join(dir, "shared.schema")

	writeSchemaFile(t, sharedSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/shared/defs",
  "$defs": {
    "Order": {
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
}`)

	writeSchemaFile(t, consumerSchemaPath, `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer",
  "title": "Consumer",
  "type": "object",
  "properties": {
    "order": {
      "$ref": "./shared.schema#/$defs/Order"
    }
  }
}`)

	cfg := testConfigWithMappings(
		SchemaMapping{
			SchemaID:    "https://example.com/consumer",
			OutputName:  "consumer.go",
			PackageName: "testpkg",
		},
	)
	cfg.StructNameFromTitle = true

	gen, err := New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(consumerSchemaPath))

	sources, err := gen.Sources()
	require.NoError(t, err)

	src, ok := sources["consumer.go"]
	require.True(t, ok)

	generated := string(src)
	// Must import the external package.
	require.Contains(t, generated, `shared "github.com/example/shared"`)
	// Must use the qualified type name.
	require.Contains(t, generated, "shared.Order")
	// Must NOT generate a local Order struct.
	require.NotContains(t, generated, "type Order struct")
}

func writeSchemaFile(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
}

func testConfigWithMappings(mappings ...SchemaMapping) Config {
	return Config{
		SchemaMappings:     mappings,
		ExtraImports:       true,
		DefaultPackageName: "testpkg",
		DefaultOutputName:  "default.go",
		ResolveExtensions:  []string{".json", ".schema"},
		YAMLExtensions:     []string{".yaml", ".yml"},
		Tags:               []string{"json", "yaml", "mapstructure"},
		Warner:             func(string) {},
	}
}
