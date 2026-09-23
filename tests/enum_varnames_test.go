package tests_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

// generateCapturingWarnings runs the generator over one schema and returns
// everything the warner emitted, so the fallback paths can be asserted on
// directly rather than inferred from the golden file.
func generateCapturingWarnings(t *testing.T, fileName string) []string {
	t.Helper()

	var (
		mu       sync.Mutex
		warnings []string
	)

	cfg := basicConfig
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

// TestEnumVarnamesGolden drives the golden files for the feature's own data
// tree, alongside the behaviour assertions below.
func TestEnumVarnamesGolden(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/enumVarnames")
}

// TestEnumVarnames covers `x-enum-varnames`, which lets a schema name its own
// enum constants so that a file shared between generators yields one Go
// identifier per enum member rather than a different one per generator.
func TestEnumVarnames(t *testing.T) {
	t.Parallel()

	t.Run("a supplied varname is the complete constant name", func(t *testing.T) {
		t.Parallel()

		warnings := generateCapturingWarnings(t, "./data/enumVarnames/enumVarnames.json")

		// The point of the extension: `Created`, not `JobStatusCreated`.
		// That is what oapi-codegen emits for the same file, and matching
		// it is the reason to honour the extension at all.
		assert.Empty(t, warnings, "the valid fixture must not warn")
	})

	t.Run("a length mismatch refuses the whole extension", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// Three names for two values cannot be resolved positionally
		// without guessing which value was meant to be skipped, and
		// guessing would silently misname a constant.
		assert.Contains(t, joined, "Mismatch declares 3 x-enum-varnames for 2 enum values")
		assert.Contains(t, joined, "ignoring the extension")
	})

	t.Run("a duplicate name falls back for the later entry", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// Emitting both would be a redeclaration that does not compile.
		assert.Contains(t, joined, `x-enum-varnames[1] "Same" collides with entry 0`)
	})

	t.Run("a name already declared in the package falls back", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// Package.AddDecl dedupes by name and keeps the first declaration,
		// so emitting a constant named after its own enum's type would drop
		// it in silence rather than fail.
		assert.Contains(t, joined, `x-enum-varnames[0] "SelfNamed" is already declared in this package`)
	})

	t.Run("a name another entry would be derived anyway falls back", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// Both constants must survive: emitting the supplied name would
		// collide with the derived one and AddDecl drops the loser silently.
		assert.Contains(t, joined, `x-enum-varnames[0] "ShadowingBeta" is the name entry 1 would be given anyway`)
	})

	t.Run("an entry with no identifier characters falls back", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// Guards the sentinel: Identifierize answers "Undefined" for input
		// it cannot use, which would otherwise become a real constant name
		// and collide with any other unusable entry.
		assert.Contains(t, joined, `x-enum-varnames[1] "!!!" yields no usable Go identifier`)
		assert.NotContains(t, joined, "Undefined")
	})
}
