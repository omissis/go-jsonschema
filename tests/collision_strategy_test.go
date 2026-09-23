package tests_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

const (
	alphaSchema = "./data/collisionStrategy/alphaResults.schema.json"
	betaSchema  = "./data/collisionStrategy/betaResults.schema.json"
)

// generateFiles runs one generator over several schemas, the way the CLI does
// when it is handed a glob. The collision this exercises only arises across
// top-level inputs, so it cannot be reproduced one file at a time: within a
// single schema, `$defs` names are unique by construction.
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

// declaredType returns the body of `type <name> struct { ... }`, so a test can
// tell which schema an identifier actually ended up bound to.
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

// TestCollisionStrategyPositional pins the existing default, including the
// hazard it carries. Two schemas each declaring `$defs/metadata` compete for
// the bare `Metadata` identifier, and the winner is whichever file was
// processed first — so the same identifier means different things depending on
// the order the generator was invoked with.
func TestCollisionStrategyPositional(t *testing.T) {
	t.Parallel()

	alphaFirst := generateFiles(t, basicConfig, alphaSchema, betaSchema)
	betaFirst := generateFiles(t, basicConfig, betaSchema, alphaSchema)

	// Both orders produce the same identifiers...
	assert.Contains(t, alphaFirst, "type Metadata struct {")
	assert.Contains(t, alphaFirst, "type Metadata_1 struct {")
	assert.Contains(t, betaFirst, "type Metadata struct {")
	assert.Contains(t, betaFirst, "type Metadata_1 struct {")

	// ...bound to different schemas. This is the defect, pinned so a change
	// in behaviour is visible rather than silent.
	assert.Contains(t, declaredType(alphaFirst, "Metadata"), "AlphaOnly")
	assert.Contains(t, declaredType(betaFirst, "Metadata"), "BetaOnly")
}

// TestCollisionStrategyQualify covers the opt-in fix: each definition is named
// after the schema that declares it, so the identifier depends on that schema
// alone.
func TestCollisionStrategyQualify(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.CollisionStrategy = generator.CollisionQualify

	alphaFirst := generateFiles(t, cfg, alphaSchema, betaSchema)
	betaFirst := generateFiles(t, cfg, betaSchema, alphaSchema)

	// No numeric suffix survives, because nothing collides any more.
	assert.NotContains(t, alphaFirst, "Metadata_1")
	assert.Contains(t, alphaFirst, "AlphaResultsSchemaMetadata")
	assert.Contains(t, alphaFirst, "BetaResultsSchemaMetadata")

	// The property that matters: invocation order cannot change the output.
	assert.Equal(t, sortedLines(alphaFirst), sortedLines(betaFirst),
		"qualified names must not depend on the order schemas were generated in")

	assert.Contains(t, declaredType(alphaFirst, "AlphaResultsSchemaMetadata"), "AlphaOnly")
	assert.Contains(t, declaredType(alphaFirst, "BetaResultsSchemaMetadata"), "BetaOnly")
}

// TestCollisionStrategyQualifyWithMinimalNames guards the interaction that
// would otherwise undo the fix: --minimal-names shortens an identifier to the
// least context that is still unique, which for a qualified definition means
// dropping straight back to the bare, order-dependent name.
func TestCollisionStrategyQualifyWithMinimalNames(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.CollisionStrategy = generator.CollisionQualify
	cfg.MinimalNames = true

	alphaFirst := generateFiles(t, cfg, alphaSchema, betaSchema)
	betaFirst := generateFiles(t, cfg, betaSchema, alphaSchema)

	assert.NotContains(t, alphaFirst, "type Metadata struct {")
	assert.Equal(t, sortedLines(alphaFirst), sortedLines(betaFirst))
}

func TestParseCollisionStrategy(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want generator.CollisionStrategy
	}{
		{"", generator.CollisionPositional},
		{"positional", generator.CollisionPositional},
		{"qualify", generator.CollisionQualify},
	} {
		got, err := generator.ParseCollisionStrategy(tc.in)
		require.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}

	// Rejected at startup rather than silently falling back, because the
	// choice decides generated identifiers.
	_, err := generator.ParseCollisionStrategy("bogus")
	require.ErrorIs(t, err, generator.ErrUnknownCollisionStrategy)
}

func sortedLines(s string) []string {
	lines := strings.Split(s, "\n")
	sort.Strings(lines)

	return lines
}
