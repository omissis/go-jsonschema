package tests_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	namedLikeDefinitionSchema     = "./data/core/nestedRefs/nestedRefsNamedLikeDefinition.json"
	namedLikeTwoDefinitionsSchema = "./data/core/nestedRefs/nestedRefsNamedLikeTwoDefinitions.json"
)

// TestRefNamedLikeItsDefinition covers a property whose path-derived name is
// the name of the definition it references. generateDeclaredType reserves that
// name for the property before resolving the ref, and the reservation used to
// count as a rival declaration, so the definition came out suffixed and a
// collision was reported that the schema does not contain.
func TestRefNamedLikeItsDefinition(t *testing.T) {
	t.Parallel()

	assert.Empty(t, refClashWarnings(t, basicConfig, namedLikeDefinitionSchema),
		"the property's own reservation is not a second type, so nothing collides")

	src := generateFiles(t, basicConfig, namedLikeDefinitionSchema)

	assert.Empty(t, declaredType(src, "OrderShipping_1"), "no suffixed duplicate is declared")
	assert.Contains(t, declaredType(src, "Order"), "Shipping *OrderShipping `")
	assert.Contains(t, declaredType(src, "OrderShipping"), "Carrier")
}

// TestRefNamedLikeTwoDefinitions keeps the real collision next to the
// reservation: a second definition maps to the same Go name, so one of the two
// has to be suffixed and reported, and each property must still point at its
// own definition. Releasing the reservation used to free the name the
// definition had just taken, so the next schema to want it bound it to a
// different type without a warning.
func TestRefNamedLikeTwoDefinitions(t *testing.T) {
	t.Parallel()

	warnings := refClashWarnings(t, basicConfig, namedLikeTwoDefinitionsSchema)
	require.Len(t, warnings, 1, "two definitions contest one name: one report")
	assert.Contains(t, warnings[0], `Multiple types map to the name "InvoiceShipping"`)

	src := generateFiles(t, basicConfig, namedLikeTwoDefinitionsSchema)

	assert.Contains(t, declaredType(src, "Invoice"), "Shipping *InvoiceShipping `")
	assert.Contains(t, declaredType(src, "InvoiceShipping"), "Carrier")
	assert.NotContains(t, declaredType(src, "InvoiceShipping"), "TrackingNumber")

	assert.Contains(t, src, "Other *InvoiceShipping_1 `")
	assert.Contains(t, declaredType(src, "InvoiceShipping_1"), "TrackingNumber")
}

const rememberedDefinitionSchema = "./data/collisionStrategy/rememberedDefinition.schema.json"

// TestRefToDefinitionDeclaredAsAnotherType covers a definition whose type is
// another named type: step_condition is an anyOf, so it is the type declared
// for the merged anyOf, and declares nothing itself. It used to be forgotten
// once generated, so the next reference generated it again; with the name
// taken by the merged type, that copy came out as an unused StepCondition_1,
// reported as a collision.
func TestRefToDefinitionDeclaredAsAnotherType(t *testing.T) {
	t.Parallel()

	assert.Empty(t, refClashWarnings(t, basicConfig, rememberedDefinitionSchema),
		"the definition is one type, so nothing collides")

	src := generateFiles(t, basicConfig, rememberedDefinitionSchema)

	assert.NotEmpty(t, declaredType(src, "StepCondition"))
	assert.NotContains(t, src, "type StepCondition_1 ", "the definition is not declared a second time")
}
