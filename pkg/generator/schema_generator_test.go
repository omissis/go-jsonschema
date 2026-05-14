package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
