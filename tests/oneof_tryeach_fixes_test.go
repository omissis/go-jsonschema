package tests_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	testArrayParentBounds "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/arrayParentBounds"
	testArrayParentItems "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/arrayParentItems"
	testArrayUnchecked "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/arrayUncheckedElements"
	testOneOfArrayVariants "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/arrayVariants"
	testNotAdditionalProperties "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/notAdditionalProperties"
	testParentProperties "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/parentProperties"
	testOverlappingRecursion "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/tryEachOverlappingRecursion"
	testTryEachSelfReference "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/tryEachSelfReference"
)

// TestOneOfTryEachArrayLengthYAML: the YAML body runs the same length checks as
// the JSON one, which TestJsonUnmarshalOneOfTryEachArrayVariants covers.
func TestOneOfTryEachArrayLengthYAML(t *testing.T) {
	t.Parallel()

	var v testOneOfArrayVariants.LicenseChoice

	require.NoError(t, yamlv3.Unmarshal([]byte("[]\n"), &v))
	require.NotNil(t, v.Variant0)
	require.Nil(t, v.Variant1)

	require.ErrorContains(t, yamlv3.Unmarshal([]byte("- expression: MIT\n- expression: BSD-3-Clause\n"), &v),
		"no oneOf variant matched")
}

// TestOneOfTryEachInlineMaps generates with DisableCustomTypesForMaps, which
// makes an object variant without properties an inline map. The holder needs a
// named type per variant, so such a oneOf keeps the generic type; generation
// used to fail outright.
func TestOneOfTryEachInlineMaps(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.DisableCustomTypesForMaps = true
	testExampleFile(t, cfg, "./data/oneOfInlineMaps/propertylessVariants/propertylessVariants.json")
}

// TestOneOfTryEachArrayParentItems: the array holding a try-each oneOf declares
// its own element schema, which a holder could not check. Such a oneOf keeps the
// typed array, so an element missing the required id is still refused.
func TestOneOfTryEachArrayParentItems(t *testing.T) {
	t.Parallel()

	var v testArrayParentItems.ArrayParentItems

	require.ErrorContains(t, json.Unmarshal([]byte(`{"list": [{"license": "MIT"}]}`), &v), "field id in")
	require.NoError(t, json.Unmarshal([]byte(`{"list": [{"id": "a", "license": "MIT"}]}`), &v))
}

// TestOneOfTryEachArrayParentBounds: the array holding a try-each oneOf bounds
// its own length, 2 to 3 elements, and the expression variant caps it at 2. The
// holder used to drop the outer bounds, which the field it replaces had checked.
func TestOneOfTryEachArrayParentBounds(t *testing.T) {
	t.Parallel()

	license := `{"license": "MIT"}`
	expression := `{"expression": "MIT"}`

	for _, tc := range []struct {
		name    string
		payload string
		variant int // -1 when no variant may match
	}{
		{"below the outer minItems", license, -1},
		{"within the outer bounds", license + "," + license, 0},
		{"above the outer maxItems", license + "," + license + "," + license + "," + license, -1},
		{"within the variant's own maxItems", expression + "," + expression, 1},
		{"above the variant's own maxItems", expression + "," + expression + "," + expression, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v testArrayParentBounds.ArrayParentBounds

			err := json.Unmarshal([]byte(`{"licenses": [`+tc.payload+`]}`), &v)
			if tc.variant < 0 {
				require.ErrorContains(t, err, "no oneOf variant matched")

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.variant == 0, v.Licenses.Variant0 != nil)
			require.Equal(t, tc.variant == 1, v.Licenses.Variant1 != nil)
		})
	}
}

// TestOneOfTryEachNotAdditionalProperties: only `additionalProperties: false`
// makes a variant strict. The `not` schema here allows any extra value that is
// not a string, so an extra key must not rule its variant out.
func TestOneOfTryEachNotAdditionalProperties(t *testing.T) {
	t.Parallel()

	var v testNotAdditionalProperties.NotAdditionalProperties

	require.NoError(t, json.Unmarshal([]byte(`{"value": {"name": "n", "extra": 5}}`), &v))
	require.NotNil(t, v.Value.Variant0)

	// The variant with `additionalProperties: false` still refuses an extra key.
	require.ErrorContains(t, json.Unmarshal([]byte(`{"value": {"id": 1, "extra": 5}}`), &v),
		"no oneOf variant matched")
}

// TestOneOfTryEachOverlappingRecursion: when two variants can both pass the
// shape check and both lead back to their oneOf, try-each would decode every
// level of nesting once per variant, doubling the work with each level. Such a
// oneOf keeps the generic representation.
func TestOneOfTryEachOverlappingRecursion(t *testing.T) {
	t.Parallel()

	// Under try-each the innermost level passes both variants and is
	// ambiguous, so this document failed, after 2^12 decodes; every level
	// more doubled the time.
	const depth = 12

	doc := `{"node": ` + strings.Repeat(`{"children": [`, depth) + `{"children": []}` +
		strings.Repeat(`]}`, depth) + `, "chain": {"x": "a", "next": {"y": "b"}}}`

	var v testOverlappingRecursion.TryEachOverlappingRecursion

	require.NoError(t, json.Unmarshal([]byte(doc), &v))
	require.NotNil(t, v.Node)
	require.NotNil(t, v.Chain.Next)
}

// TestOneOfTryEachSelfReference decodes a try-each oneOf whose variant refers
// back to it. Its generated code used not to compile.
func TestOneOfTryEachSelfReference(t *testing.T) {
	t.Parallel()

	var v testTryEachSelfReference.Node

	require.NoError(t, json.Unmarshal([]byte(`{"children": [{"leaf": "a"}, {"children": []}]}`), &v))
	require.NotNil(t, v.Variant1)
	require.Len(t, v.Variant1.Children, 2)
	require.NotNil(t, v.Variant1.Children[0].Variant0)
	require.Equal(t, "a", v.Variant1.Children[0].Variant0.Leaf)
	require.NotNil(t, v.Variant1.Children[1].Variant1)
}

// TestOneOfParentProperties: an instance must satisfy the schema holding a oneOf
// as well as one of its variants, so the properties the parent declares beside
// its oneOf belong in every variant. They used to vanish from the generated
// variants, and with them their type and required checks.
func TestOneOfParentProperties(t *testing.T) {
	t.Parallel()

	t.Run("discriminated variants keep the parent's properties", func(t *testing.T) {
		t.Parallel()

		var v testParentProperties.ParentProperties

		require.NoError(t, json.Unmarshal([]byte(`{"pet": {"kind": "dog", "name": "Rex"}}`), &v))
		require.NotNil(t, v.Pet.Dog)
		require.Equal(t, "Rex", v.Pet.Dog.Name)

		require.ErrorContains(t, json.Unmarshal([]byte(`{"pet": {"kind": "cat"}}`), &v),
			"field name in ParentPropertiesPetCat: required")
		require.ErrorContains(t, json.Unmarshal([]byte(`{"pet": {"kind": "cat", "name": 5}}`), &v),
			"unmarshal ParentPropertiesPetCat")
	})

	t.Run("try-each variants keep the parent's property types", func(t *testing.T) {
		t.Parallel()

		var v testParentProperties.ParentProperties

		require.NoError(t, json.Unmarshal([]byte(`{"kernel": {"package": "linux-generic"}}`), &v))
		require.NotNil(t, v.Kernel.Variant0)
		require.Equal(t, "linux-generic", v.Kernel.Variant0.Package)

		// Each variant only requires its key; the type comes from the parent.
		require.ErrorContains(t, json.Unmarshal([]byte(`{"kernel": {"package": 5}}`), &v),
			"no oneOf variant matched")
		require.ErrorContains(t, yamlv3.Unmarshal([]byte("kernel:\n  package: [5]\n"), &v),
			"no oneOf variant matched")
	})

	t.Run("a variant redefining a parent property keeps the parent struct", func(t *testing.T) {
		t.Parallel()

		var v testParentProperties.ParentProperties

		// A variant redeclares name without the parent's maxLength 5, which
		// still holds.
		require.ErrorContains(t, json.Unmarshal([]byte(`{"redefined": {"name": "toolongname", "a": "x"}}`), &v),
			"field name length: must be <= 5")
	})

	t.Run("a parent-declared discriminator keeps the holder", func(t *testing.T) {
		t.Parallel()

		var v testParentProperties.ParentProperties

		require.NoError(t, json.Unmarshal([]byte(`{"keyed": {"kind": "dog", "name": "Rex"}}`), &v))
		require.NotNil(t, v.Keyed.Dog)
		require.ErrorContains(t, json.Unmarshal([]byte(`{"keyed": {"kind": "cat", "name": "toolongname"}}`), &v),
			"must be <= 5")
	})
}

// TestOneOfArrayUncheckedElements: an array variant whose elements need no
// per-element check compiles and dispatches.
func TestOneOfArrayUncheckedElements(t *testing.T) {
	t.Parallel()

	var v testArrayUnchecked.ArrayUncheckedElements

	require.NoError(t, json.Unmarshal([]byte(`{"licenses": [{"license": "MIT"}]}`), &v))
	require.NotNil(t, v.Licenses.Variant0)
	require.Nil(t, v.Licenses.Variant1)
}
