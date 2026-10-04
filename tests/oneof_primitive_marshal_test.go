package tests_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	testInField "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/inField"
	testNumString "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/numString"
	testTemporal "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/temporalVariant"
)

// TestOneOfPrimitiveMarshalByValue marshals a parent holding a required wrapper
// by value. With pointer-receiver Marshal methods, encoding/json could not find
// MarshalJSON on the unaddressable field and yaml.v3 never looks for one there,
// so both emitted the wrapper's unexported fields as `{}`.
func TestOneOfPrimitiveMarshalByValue(t *testing.T) {
	t.Parallel()

	var v testInField.InField

	require.NoError(t, json.Unmarshal([]byte(`{"name": "a", "value": "x"}`), &v))

	out, err := json.Marshal(v)
	require.NoError(t, err)
	require.JSONEq(t, `{"name": "a", "value": "x"}`, string(out))

	y, err := yamlv3.Marshal(v)
	require.NoError(t, err)

	var back testInField.InField

	require.NoError(t, yamlv3.Unmarshal(y, &back), "yaml: %s", y)

	got, ok := back.Value.AsString()
	require.True(t, ok, "yaml: %s", y)
	require.Equal(t, "x", got)
}

// TestOneOfPrimitiveYAMLUnquotedTimestamp decodes unquoted YAML dates, which
// yaml.v3 tags `!!timestamp` rather than `!!str`. They are strings in the
// schema's terms, for a temporal variant and for a plain one alike.
func TestOneOfPrimitiveYAMLUnquotedTimestamp(t *testing.T) {
	t.Parallel()

	var tv testTemporal.TemporalVariant

	require.NoError(t, yamlv3.Unmarshal([]byte("timeOrNumber: 2024-01-02T03:04:05Z\n"), &tv))

	when, ok := tv.TimeOrNumber.AsDateTime()
	require.True(t, ok)
	require.True(t, when.Equal(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)), "got %v", when)

	var ns testNumString.NumString

	require.NoError(t, yamlv3.Unmarshal([]byte("value: 2024-01-02\n"), &ns))

	s, ok := ns.Value.AsString()
	require.True(t, ok)
	require.Equal(t, "2024-01-02", s)
}
