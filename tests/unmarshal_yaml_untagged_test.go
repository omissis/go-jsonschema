package tests_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	testYAMLUntagged "github.com/atombender/go-jsonschema/tests/data/extraImports/yamlUntaggedFields"
)

// TestExtraImportsYAMLUntaggedFields generates with Tags set to json only, so no
// field carries a yaml tag and UnmarshalYAML's pruning has to fall back to the
// name yaml.v3 binds an untagged field under.
func TestExtraImportsYAMLUntaggedFields(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.ExtraImports = true
	cfg.Tags = []string{"json"}
	testExampleFile(t, cfg, "./data/extraImports/yamlUntaggedFields/yamlUntaggedFields.json")
}

// TestYamlAdditionalPropertiesPrunesUntaggedFields pins that fallback. yaml.v3
// binds a field with no yaml tag under its lowercased Go name, so `foo` binds to
// Foo. Pruning must remove that same key, or the value lands both in the field
// and in AdditionalProperties.
func TestYamlAdditionalPropertiesPrunesUntaggedFields(t *testing.T) {
	t.Parallel()

	var v testYAMLUntagged.YamlUntaggedFields

	require.NoError(t, yamlv3.Unmarshal([]byte("foo: bound\nextra: x\n"), &v))
	require.NotNil(t, v.Foo, "untagged field Foo should bind the key `foo`")
	require.Equal(t, "bound", *v.Foo)
	require.Equal(t, map[string]string{"extra": "x"}, v.AdditionalProperties)
}
