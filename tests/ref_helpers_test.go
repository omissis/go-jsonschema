package tests_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

// refClashWarnings generates fileNames with cfg and returns the warnings that
// report a name collision.
func refClashWarnings(t *testing.T, cfg generator.Config, fileNames ...string) []string {
	t.Helper()

	var (
		mu  sync.Mutex
		out []string
	)

	cfg.Warner = func(message string) {
		mu.Lock()
		defer mu.Unlock()

		if strings.Contains(message, "Multiple types map to the name") {
			out = append(out, message)
		}
	}

	g, err := generator.New(cfg)
	require.NoError(t, err)

	for _, f := range fileNames {
		require.NoError(t, g.DoFile(f))
	}

	_, err = g.Sources()
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), out...)
}
