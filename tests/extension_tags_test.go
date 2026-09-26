package tests_test

import (
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/generator"
	testExtensionTags "github.com/atombender/go-jsonschema/tests/data/extensionTags"
)

// generateCapturingWarningsWithConfig runs the generator over one schema under
// cfg and returns everything the warner emitted, so fallback paths can be
// asserted on directly rather than inferred from the golden file.
func generateCapturingWarningsWithConfig(
	t *testing.T, cfg generator.Config, fileName string,
) []string {
	t.Helper()

	var (
		mu       sync.Mutex
		warnings []string
	)

	cfg.Warner = func(message string) {
		mu.Lock()
		defer mu.Unlock()

		warnings = append(warnings, message)
	}

	g, err := generator.New(cfg)
	require.NoError(t, err)
	require.NoError(t, g.DoFile(fileName))

	_, err = g.Sources()
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), warnings...)
}

// generateSources returns every generated source for one schema, concatenated,
// for assertions about what did or did not reach the output.
func generateSources(t *testing.T, cfg generator.Config, fileName string) string {
	t.Helper()

	g, err := generator.New(cfg)
	require.NoError(t, err)
	require.NoError(t, g.DoFile(fileName))

	sources, err := g.Sources()
	require.NoError(t, err)

	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}

	sort.Strings(names)

	var sb strings.Builder
	for _, name := range names {
		sb.Write(sources[name])
	}

	return sb.String()
}

// extensionTagConfig maps the extensions the fixtures declare. `x-not-configured`
// is deliberately absent: an extension only reaches the output when the caller
// asks for it by name.
func extensionTagConfig() generator.Config {
	cfg := basicConfig
	cfg.ExtensionTags = map[string]string{
		"x-measurement": "slb-measurement",
		"x-precision":   "precision",
		"x-derived":     "derived",
	}

	return cfg
}

// TestExtensionTags covers `--extension-tag`, which carries a schema's `x-`
// vendor data into the struct tag so reflection-based consumers can read it.
// Without it the declaration is dropped and the same schema produces types that
// work through one generator and not another.
func TestExtensionTags(t *testing.T) {
	t.Parallel()

	testExamples(t, extensionTagConfig(), "./data/extensionTags")
}

func TestExtensionTagsBehaviour(t *testing.T) {
	t.Parallel()

	t.Run("mapped extensions emit, unmapped ones do not", func(t *testing.T) {
		t.Parallel()

		warnings := generateCapturingWarningsWithConfig(
			t, extensionTagConfig(), "./data/extensionTags/extensionTags.json",
		)

		assert.Empty(t, warnings, "the valid fixture must not warn")
	})

	t.Run("composite values are refused rather than stringified", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(generateCapturingWarningsWithConfig(
			t, extensionTagConfig(), "./data/extensionTags/extensionTagsSkipped.json",
		), "\n")

		// A struct tag is flat text; rendering an object or array into one
		// would be this generator's invention, not something the schema said.
		assert.Contains(t, joined, `Property "composite" declares x-measurement with a map[string]interface {}`)
		assert.Contains(t, joined, `Property "listed" declares x-measurement with a []interface {}`)
		assert.Contains(t, joined, "only strings, numbers and booleans")
	})

	t.Run("a backtick is refused because it would not compile", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(generateCapturingWarningsWithConfig(
			t, extensionTagConfig(), "./data/extensionTags/extensionTagsSkipped.json",
		), "\n")

		// The tag list is emitted inside a raw string literal, which a
		// backtick terminates.
		assert.Contains(t, joined, `Property "backtick" declares x-measurement with a value containing a backtick`)
	})

	t.Run("no configured extensions leaves output untouched", func(t *testing.T) {
		t.Parallel()

		// The feature is entirely opt-in: with no mapping, a schema full of
		// extensions generates exactly what it generated before.
		sources := generateSources(t, basicConfig, "./data/extensionTags/extensionTags.json")

		assert.NotContains(t, sources, "slb-measurement")
		assert.NotContains(t, sources, "precision:")
		assert.NotContains(t, sources, "derived:")
	})
}

// TestExtensionTagsRoundTrip reads the emitted tags back with reflect, which is
// how the consumers this feature exists for will read them. Compiling the
// fixture only proves the tag is syntactically intact; this proves the value
// survives, escaping included.
func TestExtensionTagsRoundTrip(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeFor[testExtensionTags.ExtensionTags]()

	for _, tc := range []struct {
		field string
		tag   string
		want  string
	}{
		{"OilRate", "slb-measurement", "Volume_Flowrate"},
		{"Precision", "precision", "3"},
		{"Derived", "derived", "true"},
		// The escaped case: an unescaped quote would close the tag early
		// and reflection would read back something truncated.
		{"Quoted", "slb-measurement", `has "quotes" inside`},
	} {
		field, ok := rt.FieldByName(tc.field)
		require.True(t, ok, "field %s missing from the generated struct", tc.field)

		got, ok := field.Tag.Lookup(tc.tag)
		require.True(t, ok, "tag %s missing on field %s", tc.tag, tc.field)
		assert.Equal(t, tc.want, got)
	}

	// An extension the caller did not map must not reach the output at all.
	unmapped, ok := rt.FieldByName("Unmapped")
	require.True(t, ok)

	_, present := unmapped.Tag.Lookup("x-not-configured")
	assert.False(t, present, "an unmapped extension must not be emitted")
}
