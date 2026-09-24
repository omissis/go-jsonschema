package tests_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testEscaping "github.com/atombender/go-jsonschema/tests/data/core/propertyNameEscaping"
)

// TestPropertyNameEscapingTags reads the emitted struct tags back with reflect.
//
// Compiling the fixture is not enough on its own. A tag value carrying an
// unescaped quote still compiles inside a raw string literal but truncates when
// reflect parses it, and one carrying an unescaped backslash makes the whole tag
// unparseable — reflect then reports it as absent and encoding/json silently
// falls back to the Go field name. Both failures are invisible without this.
func TestPropertyNameEscapingTags(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeFor[testEscaping.PropertyNameEscaping]()

	for _, tc := range []struct{ field, want string }{
		{"AB", `a"b`},
		{"BackSlash", `back\slash`},
		{"PctS", "pct%s"},
		{"Ordinary", "ordinary"},
		{"HasTick", "has`tick"},
	} {
		sf, ok := rt.FieldByName(tc.field)
		require.True(t, ok, "field %s missing from the generated struct", tc.field)

		for _, tag := range []string{"json", "yaml", "mapstructure"} {
			got, present := sf.Tag.Lookup(tag)
			require.True(t, present, "%s tag absent on %s — the tag failed to parse", tag, tc.field)

			// json carries ",omitempty,omitzero" for optional fields; every
			// field here is required, so the value is the bare name.
			assert.Equal(t, tc.want, got, "%s tag on %s", tag, tc.field)
		}
	}
}
