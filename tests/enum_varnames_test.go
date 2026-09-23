package tests_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// generateCapturingWarnings runs the generator over one schema under the
// standard config and returns everything the warner emitted.
func generateCapturingWarnings(t *testing.T, fileName string) []string {
	t.Helper()

	return generateCapturingWarningsWithConfig(t, basicConfig, fileName)
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

	t.Run("a constant a later enum would derive is not lost", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesCrossEnum.json"), "\n",
		)

		// Each enum only knows its own names, so one enum's supplied varname
		// can be the name a later enum derives. AddDecl keeps the first
		// declaration, so the later constant used to vanish outright.
		assert.Contains(t, joined, `constant name "XcSecondX" is already declared in this package`)
		assert.Contains(t, joined, `"XcSecondX_1"`)
	})

	t.Run("a varname naming an import or a predeclared identifier falls back", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// `const fmt` beside `import "fmt"` does not compile, and `const nil`
		// would shadow the nil the generated code compares errors against.
		assert.Contains(t, joined, `x-enum-varnames[0] "fmt" is the name of an imported package`)
		assert.Contains(t, joined, `x-enum-varnames[1] "nil" is a predeclared Go identifier`)
	})

	t.Run("a varname cannot take a later enum's value list", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesInvalid.json"), "\n",
		)

		// Each enum's allowed values are a variable named enumValues_<Type>.
		// A constant holding that name made AddDecl drop the variable, and
		// the enum's validator then ranged over the constant's string —
		// comparing runes, it rejected valid values and accepted others,
		// with nothing failing to compile.
		assert.Contains(t, joined,
			`x-enum-varnames[3] "enumValues_Shadowing" uses the prefix reserved for generated enum value lists`)
	})

	t.Run("a constant's name is not given to a later type", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesCrossEnum.json"), "\n",
		)

		// Type allocation only knew other types, so XcTarget was declared
		// under the constant's name and AddDecl dropped it: the output
		// referred to a type that no longer existed.
		assert.Contains(t, joined, `Multiple types map to the name "XcTarget"; declaring duplicate as "XcTarget_1" instead`)
	})

	t.Run("a varname cannot take a type still being generated", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesCrossEnum.json"), "\n",
		)

		// XcHolder is registered as a type but not yet declared in the
		// package when its own property's varname is resolved.
		assert.Contains(t, joined, `x-enum-varnames[0] "XcHolder" is already declared in this package`)
	})

	t.Run("a constant an import later claims is renamed", func(t *testing.T) {
		t.Parallel()

		joined := strings.Join(
			generateCapturingWarnings(t, "./data/enumVarnames/enumVarnamesCrossEnum.json"), "\n",
		)

		// Nothing imports time when the constant is named; a later type
		// does, so the clash is only visible once generation ends.
		assert.Contains(t, joined,
			`Constant "time" shares its name with an imported package; declaring it as "time_1" instead`)
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
