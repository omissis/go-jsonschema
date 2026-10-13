package tests_test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

// TestFidelityWarningsBehavior verifies that the schema-fidelity warning
// emitted when a schema declares enforcement-implying keywords but degrades
// to interface{} fires for the right schemas and stays silent for genuinely
// open schemas.
//
// Note on the ifThenSilentDrop fixture: its allOf deliberately carries a
// `{"type": "string"}` member alongside the if/then branches. Upstream
// (omissis/go-jsonschema#590) made isPrimitiveTypeList require at least one
// member to contribute a real primitive type, so an allOf of *only* typeless
// conditionals no longer collapses the surrounding object — it now generates a
// proper struct, and this warning correctly stays silent for it. The primitive
// member is what still drives MergeTypes down its primitive short circuit,
// dropping the parent's object shape and producing the interface{} fallback
// this test is about. Do not remove it as redundant: without it the schema
// generates cleanly and the test no longer exercises the warning at all.
func TestFidelityWarningsBehavior(t *testing.T) {
	t.Parallel()

	t.Run("if/then with additionalProperties:false + required emits a warning", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t, "./data/fidelityWarnings/ifThenSilentDrop/ifThenSilentDrop.json")
		joined := strings.Join(warnings, "\n")

		require.NotEmpty(t, warnings, "expected at least one warning for if/then schema with enforcement keywords")
		assert.Contains(t, joined, "schema fidelity:")
		assert.Contains(t, joined, "if/then/else not compiled")
		assert.Contains(t, joined, "additionalProperties: false")
		assert.Contains(t, joined, "required")
		assert.Contains(t, joined, "declared property(ies)")
	})

	t.Run("constrained oneOf variant warns and names the keyword", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t,
			"./data/fidelityWarnings/oneOfConstrainedVariant/oneOfConstrainedVariant.json")
		joined := strings.Join(warnings, "\n")

		// The primitive wrapper dispatches on JSON token kind, so a variant
		// carrying a constraint the wrapper cannot enforce disqualifies the
		// whole schema and it degrades to interface{}. Before this warning
		// existed that was entirely silent, which is how it went unnoticed in
		// production schemas.
		require.NotEmpty(t, warnings)
		assert.Contains(t, joined, "schema fidelity:")
		assert.Contains(t, joined, "StillRejected")
		assert.Contains(t, joined, "primitive wrapper cannot enforce")
		assert.Contains(t, joined, "string length constraint")
	})

	t.Run("a temporal format is a type mapping, not a constraint, and stays silent", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t,
			"./data/fidelityWarnings/oneOfConstrainedVariant/oneOfConstrainedVariant.json")

		// `format: date-time` on a string variant maps the branch to time.Time
		// rather than declaring a constraint the wrapper must enforce, so the
		// schema still compiles to a real wrapper. Guards the boundary against
		// the disqualifying case above: only one of the two degrades.
		for _, w := range filterFidelityWarnings(warnings) {
			assert.NotContains(t, w, "TimeOrNumber",
				"a temporal format compiles to a time.Time branch and must not warn")
		}
	})

	t.Run("an unconstrained oneOf compiles and stays silent", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t,
			"./data/fidelityWarnings/oneOfConstrainedVariant/oneOfConstrainedVariant.json")

		// Same file: the sibling property has no constraints, gets a real
		// wrapper, and must not be reported. One field degrades, not two.
		for _, w := range filterFidelityWarnings(warnings) {
			assert.NotContains(t, w, "Unconstrained",
				"the unconstrained variant compiles to a wrapper and must not warn")
		}
	})

	t.Run("anyOf mixing object and string branches falls back and warns", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t, "./data/core/anyOf/anyOfRefToString.json")
		joined := strings.Join(filterFidelityWarnings(warnings), "\n")

		assert.Contains(t, joined, "anyOf with both object and non-object branches")
		assert.Contains(t, joined, "declared property(ies)")
	})

	t.Run("a oneOf of required-only branches on a struct is reported", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t, "./data/fidelityWarnings/oneOfRequiredOnly/oneOfRequiredOnly.json")
		joined := strings.Join(filterFidelityWarnings(warnings), "\n")

		assert.Contains(t, joined, "does not enforce all it declares: oneOf of 2 branches not compiled")
	})

	t.Run("a conditional the allOf merge drops is reported", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t,
			"./data/fidelityWarnings/conditionalRequiredInIf/conditionalRequiredInIf.json")
		joined := strings.Join(filterFidelityWarnings(warnings), "\n")

		assert.Contains(t, joined, "does not enforce all it declares: if/then/else not compiled (in allOf branch)")
	})

	t.Run("a $ref branch dropped with its oneOf is named", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t, "./data/fidelityWarnings/oneOfRefOrNull/oneOfRefOrNull.json")
		joined := strings.Join(filterFidelityWarnings(warnings), "\n")

		assert.Contains(t, joined, "falls back to interface{} due to oneOf not compiled")
		assert.Contains(t, joined, "$ref #/definitions/pair")
	})

	t.Run("conditionals and negations no code path compiles are reported once each", func(t *testing.T) {
		t.Parallel()

		warnings := filterFidelityWarnings(generateWithWarnerCapture(t,
			"./data/fidelityWarnings/uncompiledConditionals/uncompiledConditionals.json"))

		assert.ElementsMatch(t, []string{
			"schema fidelity: type UncompiledConditionals does not enforce all it declares: " +
				"if/then/else not compiled",
			"schema fidelity: type UncompiledConditionalsChoice_0 does not enforce all it declares: " +
				`"not" keyword not compiled`,
			"schema fidelity: type UncompiledConditionalsCode falls back to interface{} due to " +
				"if/then/else not compiled (in allOf branch); declared but not enforced: if/then/else (in allOf branch)",
			"schema fidelity: type UncompiledConditionalsEither falls back to interface{} due to " +
				`oneOf not compiled; declared but not enforced: "not" (in oneOf branch)`,
			"schema fidelity: type UncompiledConditionalsLegacy does not enforce all it declares: " +
				"false schema not compiled, so any value passes",
			"schema fidelity: type UncompiledConditionalsMerged does not enforce all it declares: " +
				"if/then/else not compiled (in allOf branch)",
			"schema fidelity: type UncompiledConditionalsName does not enforce all it declares: " +
				`"not" keyword not compiled`,
			"schema fidelity: type UncompiledConditionalsStatus does not enforce all it declares: " +
				`"not" keyword not compiled`,
			"schema fidelity: type UncompiledConditionalsValue does not enforce all it declares: " +
				`"not" keyword not compiled (in oneOf branch)`,
			"schema fidelity: type NamesElem does not enforce all it declares: " +
				`"not" keyword not compiled`,
			"schema fidelity: type Shape does not enforce all it declares: " +
				"if/then/else not compiled",
		}, warnings)
	})

	t.Run("only models enforce nothing, so nothing is reported as not enforced", func(t *testing.T) {
		t.Parallel()

		cfg := basicConfig
		cfg.OnlyModels = true

		// A type that falls back to interface{} is still reported: its shape
		// is lost whether or not anything is enforced.
		var notEnforced []string

		for _, w := range filterFidelityWarnings(generateWithConfigAndWarnerCapture(t, cfg,
			"./data/fidelityWarnings/uncompiledConditionals/uncompiledConditionals.json")) {
			if strings.Contains(w, "does not enforce all it declares") {
				notEnforced = append(notEnforced, w)
			}
		}

		assert.Empty(t, notEnforced)
	})

	t.Run("conditionals that lose nothing stay silent", func(t *testing.T) {
		t.Parallel()

		warnings := filterFidelityWarnings(generateWithWarnerCapture(t,
			"./data/fidelityWarnings/vacuousConditionals/vacuousConditionals.json"))

		assert.Empty(t, warnings, "an if alone, a vacuous then and else, and a compiled discriminator lose nothing")
	})

	t.Run("schema with no enforcement-implying keywords stays silent", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t, "./data/fidelityWarnings/openSchemaNoWarn/openSchemaNoWarn.json")

		fidelityWarnings := filterFidelityWarnings(warnings)
		assert.Empty(t, fidelityWarnings, "expected no fidelity warnings for an open schema, got: %v", fidelityWarnings)
	})
}

// generateWithWarnerCapture runs the generator on a single schema with a
// warner that records every emitted message. Returns the captured slice.
func generateWithWarnerCapture(t *testing.T, schemaPath string) []string {
	t.Helper()

	return generateWithConfigAndWarnerCapture(t, basicConfig, schemaPath)
}

// generateWithConfigAndWarnerCapture is generateWithWarnerCapture with cfg in
// place of basicConfig.
func generateWithConfigAndWarnerCapture(t *testing.T, cfg generator.Config, schemaPath string) []string {
	t.Helper()

	abs, err := filepath.Abs(schemaPath)
	require.NoError(t, err)

	var (
		mu       sync.Mutex
		captured []string
	)

	cfg.Warner = func(msg string) {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, msg)
	}

	gen, err := generator.New(cfg)
	require.NoError(t, err)
	require.NoError(t, gen.DoFile(abs))

	mu.Lock()
	defer mu.Unlock()

	out := make([]string, len(captured))
	copy(out, captured)

	return out
}

// filterFidelityWarnings returns only the messages that look like fidelity
// warnings, ignoring incidental warnings that may come from other paths.
func filterFidelityWarnings(msgs []string) []string {
	var out []string

	for _, m := range msgs {
		if strings.HasPrefix(m, "schema fidelity:") {
			out = append(out, m)
		}
	}

	return out
}
