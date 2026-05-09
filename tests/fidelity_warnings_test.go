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
