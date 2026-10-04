package tests_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	testAddlNot "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlNot"
)

// TestStrictAdditionalPropertiesRespectSchemaNot pins that respect-schema mode
// enforces only the `false` schema. `additionalProperties: {not: {type:
// string}}` allows a number, so rejecting `{"extra": 5}` would reject data the
// schema accepts — for a struct with properties and for a property-less object
// alike.
func TestStrictAdditionalPropertiesRespectSchemaNot(t *testing.T) {
	t.Parallel()

	var v testAddlNot.AddlNot

	require.NoError(t, json.Unmarshal([]byte(`{"foo": "x", "extra": 5, "inner": {"extra": 5}}`), &v))
	require.NoError(t, yamlv3.Unmarshal([]byte("foo: x\nextra: 5\ninner:\n  extra: 5\n"), &v))
}
