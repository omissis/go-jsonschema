package tests_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testEscaping "github.com/atombender/go-jsonschema/tests/data/core/propertyNameEscaping"
)

// pctName and punctName are the two schema property names worth asserting on:
// a percent sign has to survive fmt, and the punctuation set is everything
// encoding/json accepts in a tag beyond letters and digits.
const (
	pctName   = "pct%s"
	punctName = "punct!#$&()*+-./:;<=>?@[]^_{|}~"
)

// TestPropertyNameEscapingTags reads the emitted tags back with reflect.
//
// Compiling the fixture is not enough on its own: a tag value carrying an
// unescaped quote still compiles inside a raw string literal but truncates when
// reflect parses it. Names that cannot be represented at all are refused by the
// generator instead — see propertyNameUnrepresentable.FAIL.json.
func TestPropertyNameEscapingTags(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeFor[testEscaping.PropertyNameEscaping]()

	for _, tc := range []struct{ field, want string }{
		{"PctS", pctName},
		{"Punct", punctName},
		{"Ordinary", "ordinary"},
		{"ArrPctS", "arrPct%s"},
	} {
		sf, ok := rt.FieldByName(tc.field)
		require.True(t, ok, "field %s missing from the generated struct", tc.field)

		for _, tag := range []string{"json", "yaml", "mapstructure"} {
			got, present := sf.Tag.Lookup(tag)
			require.True(t, present, "%s tag absent on %s — the tag failed to parse", tag, tc.field)
			assert.Equal(t, tc.want, got, "%s tag on %s", tag, tc.field)
		}
	}
}

// TestPropertyNameEscapingJSONRoundTrip is the assertion reflect cannot make.
//
// reflect.StructTag.Lookup reads a tag back as written, but encoding/json
// applies its own rules to the tag NAME and silently ignores one it does not
// accept — decoding leaves the field unset and encoding emits the Go field name
// or a truncated key. So a tag being intact says nothing about the codec using
// it, and only a round trip through encoding/json does.
func TestPropertyNameEscapingJSONRoundTrip(t *testing.T) {
	t.Parallel()

	in := testEscaping.PropertyNameEscaping{
		PctS:     "a",
		Punct:    "b",
		Ordinary: "c",
		ArrPctS:  [][]string{{"d"}},
	}

	encoded, err := json.Marshal(in)
	require.NoError(t, err)

	// The schema's own names must be the keys on the wire, not the Go
	// field names.
	var asMap map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &asMap))
	assert.Contains(t, asMap, pctName)
	assert.Contains(t, asMap, punctName)
	assert.NotContains(t, asMap, "PctS")

	var out testEscaping.PropertyNameEscaping
	require.NoError(t, json.Unmarshal(encoded, &out))
	assert.Equal(t, in, out)
}
