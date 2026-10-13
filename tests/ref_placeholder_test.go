package tests_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const aliasCycleSchema = "./data/refCycles/aliasCycle.schema.json"

// TestAliasDefinitionCycleTerminates covers definitions that are only $refs to
// one another. Their declarations stay placeholders until a ref resolves, and
// every placeholder leads to the next, so generation must keep one rather than
// resolve the ref again. The output names a type the schema never gives
// content; this checks only that generation finishes.
func TestAliasDefinitionCycleTerminates(t *testing.T) {
	t.Parallel()

	assert.Contains(t, generateFiles(t, basicConfig, aliasCycleSchema), "type ")
}
