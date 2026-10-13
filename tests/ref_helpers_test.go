package tests_test

import (
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

// generateFiles generates fileNames with cfg and returns every output file's
// source, concatenated in file-name order.
func generateFiles(t *testing.T, cfg generator.Config, fileNames ...string) string {
	t.Helper()

	g, err := generator.New(cfg)
	require.NoError(t, err)

	for _, fileName := range fileNames {
		require.NoError(t, g.DoFile(fileName))
	}

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

// declaredType returns the declaration of the struct named name in src, or ""
// if src declares none.
func declaredType(src, name string) string {
	marker := "type " + name + " struct {"

	start := strings.Index(src, marker)
	if start == -1 {
		return ""
	}

	end := strings.Index(src[start:], "\n}")
	if end == -1 {
		return src[start:]
	}

	return src[start : start+end]
}

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
