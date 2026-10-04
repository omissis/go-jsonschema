package tests_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	testAnimal "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/animal"
	testCaseCollision "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/caseCollision"
	testNumericKind "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/numericKind"
	testNumericNull "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/numericNull"
	testSelfReference "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/selfReference"
)

// TestOnlyModelsOneOfDiscriminated: under OnlyModels no holder is built, so a
// discriminated oneOf property stays an inline interface{}.
func TestOnlyModelsOneOfDiscriminated(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.OnlyModels = true
	testExampleFile(t, cfg, "./data/onlyModels/oneOfDiscriminated/oneOfDiscriminated.json")
}

// TestOneOfDiscriminatorEdgeShapes decodes discriminated oneOf shapes whose
// generated code used not to compile, or dispatched to the wrong variant.
func TestOneOfDiscriminatorEdgeShapes(t *testing.T) {
	t.Parallel()

	t.Run("case-colliding discriminator values get distinct fields", func(t *testing.T) {
		t.Parallel()

		var v testCaseCollision.CaseCollision

		require.NoError(t, json.Unmarshal([]byte(`{"pet": {"kind": "Dog", "howl": "aw"}}`), &v))
		require.Nil(t, v.Pet.Dog)
		require.NotNil(t, v.Pet.Dog_2)

		require.NoError(t, json.Unmarshal([]byte(`{"pet": {"kind": "dog", "bark": "wf"}}`), &v))
		require.NotNil(t, v.Pet.Dog)
		require.Nil(t, v.Pet.Dog_2)
	})

	t.Run("a variant referring back to its oneOf resolves to the holder", func(t *testing.T) {
		t.Parallel()

		var v testSelfReference.Pet

		require.NoError(t, json.Unmarshal([]byte(`{"kind": "dog", "friend": {"kind": "cat"}}`), &v))
		require.NotNil(t, v.Dog)
		require.NotNil(t, v.Dog.Friend)
		require.NotNil(t, v.Dog.Friend.Cat)
	})

	t.Run("null routes to the null variant beside numeric ones", func(t *testing.T) {
		t.Parallel()

		var v testNumericNull.NumericNull

		require.NoError(t, json.Unmarshal([]byte(`{"withNull": {"version": null}}`), &v))
		require.NotNil(t, v.WithNull.Null)

		require.NoError(t, json.Unmarshal([]byte(`{"withNull": {"version": 2}}`), &v))
		require.NotNil(t, v.WithNull.Const2)

		require.NoError(t, yamlv3.Unmarshal([]byte("withNull:\n  version: null\n"), &v))
		require.NotNil(t, v.WithNull.Null)
	})

	t.Run("null is not decoded as 0", func(t *testing.T) {
		t.Parallel()

		var v testNumericNull.NumericNull

		require.ErrorContains(t, json.Unmarshal([]byte(`{"withZero": {"version": null}}`), &v),
			"must be numeric, got null")

		// yaml.v3 decodes a null scalar into a float64 as 0 without an error.
		require.ErrorContains(t, yamlv3.Unmarshal([]byte("withZero:\n  version: null\n"), &v),
			"must be numeric, got null")
	})

	t.Run("an escaped string discriminator matches", func(t *testing.T) {
		t.Parallel()

		// "dog" is "dog". Matching the raw token missed it wherever
		// json.RawMessage keeps the escape, as Go 1.25's encoding/json does.
		var v testCaseCollision.CaseCollision

		require.NoError(t, json.Unmarshal([]byte(`{"pet": {"kind": "dog", "bark": "wf"}}`), &v))
		require.NotNil(t, v.Pet.Dog)
	})
}

// TestOneOfDiscriminatorYAML runs the YAML holders, which share no code with the
// JSON path.
func TestOneOfDiscriminatorYAML(t *testing.T) {
	t.Parallel()

	t.Run("string discriminator dispatches", func(t *testing.T) {
		t.Parallel()

		var v testAnimal.Animal

		require.NoError(t, yamlv3.Unmarshal([]byte("creature:\n  kind: cat\n  purr: true\n"), &v))
		require.NotNil(t, v.Creature.Cat)
		require.Nil(t, v.Creature.Dog)
	})

	t.Run("every YAML number form reaches its numeric case", func(t *testing.T) {
		t.Parallel()

		// yaml.v3 reads each of these as 1 or 2; strconv.ParseFloat on the
		// scalar text rejected the hex and octal forms.
		for _, tc := range []struct {
			doc string
			two bool
		}{
			{"payload:\n  version: 1.0\n  alpha: a\n", false},
			{"payload:\n  version: 1e0\n  alpha: a\n", false},
			{"payload:\n  version: 0x2\n  beta: b\n", true},
			{"payload:\n  version: 0o2\n  beta: b\n", true},
		} {
			var v testNumericKind.NumericKind

			require.NoError(t, yamlv3.Unmarshal([]byte(tc.doc), &v), tc.doc)
			require.Equal(t, tc.two, v.Payload.Const2 != nil, tc.doc)
			require.Equal(t, !tc.two, v.Payload.Const1 != nil, tc.doc)
		}
	})

	t.Run("missing and unknown discriminators are errors", func(t *testing.T) {
		t.Parallel()

		var v testAnimal.Animal

		require.ErrorContains(t, yamlv3.Unmarshal([]byte("creature:\n  purr: true\n"), &v), "missing discriminator")
		require.ErrorContains(t, yamlv3.Unmarshal([]byte("creature:\n  kind: cow\n"), &v), "unknown kind value")
	})

	t.Run("round trip through MarshalYAML", func(t *testing.T) {
		t.Parallel()

		var v testAnimal.Animal

		require.NoError(t, yamlv3.Unmarshal([]byte("creature:\n  kind: dog\n  barkAt: mail\n"), &v))

		out, err := yamlv3.Marshal(v)
		require.NoError(t, err)

		var back testAnimal.Animal

		require.NoError(t, yamlv3.Unmarshal(out, &back), "yaml: %s", out)
		require.NotNil(t, back.Creature.Dog, "yaml: %s", out)
		require.Equal(t, "mail", back.Creature.Dog.BarkAt)
	})
}
