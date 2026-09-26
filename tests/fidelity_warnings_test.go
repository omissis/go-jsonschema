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

	abs, err := filepath.Abs(schemaPath)
	require.NoError(t, err)

	var (
		mu       sync.Mutex
		captured []string
	)

	cfg := basicConfig
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
