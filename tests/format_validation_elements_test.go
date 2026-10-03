package tests_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/atombender/go-jsonschema/pkg/generator"
	testElements "github.com/atombender/go-jsonschema/tests/data/formatValidation/elements"
	testInlineMaps "github.com/atombender/go-jsonschema/tests/data/formatValidationInlineMaps"
)

const elementsValidUUID = "123e4567-e89b-12d3-a456-426614174000"

// TestFormatValidationInlineMaps generates with DisableCustomTypesForMaps, so a
// map stays inline in its field and the field's own validators check its values.
func TestFormatValidationInlineMaps(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.FormatValidation = generator.FormatValidationConfig{Enabled: true}
	cfg.DisableCustomTypesForMaps = true
	testExamples(t, cfg, "./data/formatValidationInlineMaps")
}

// TestFormatValidationElements decodes strings held in arrays and maps. The
// check used to reach only a string-typed field, so `items: {format: uuid}`
// accepted any string. Each case names the element the error must point at;
// a declared type does not know the property holding it, so its path starts
// at the element.
func TestFormatValidationElements(t *testing.T) {
	t.Parallel()

	u := elementsValidUUID

	t.Run("valid elements decode", func(t *testing.T) {
		t.Parallel()

		var v testElements.Elements

		require.NoError(t, json.Unmarshal([]byte(`{
			"ids": ["`+u+`"], "idGrid": [["`+u+`"]], "maybeIds": [null, "`+u+`"],
			"emails": ["a@example.com"], "byName": {"a": "`+u+`"},
			"listsByName": {"a": ["`+u+`"]}, "idList": ["`+u+`"]
		}`), &v))
	})

	for _, tc := range []struct {
		name    string
		payload string
		want    string
	}{
		{"array item", `{"ids": ["` + u + `", "nope"]}`, `field ids[1]: must be a valid uuid`},
		{"nested array item", `{"idGrid": [["` + u + `", "nope"]]}`, `field idGrid[0][1]: must be a valid uuid`},
		{"nullable item", `{"maybeIds": [null, "nope"]}`, `field maybeIds[1]: must be a valid uuid`},
		{"email item", `{"emails": ["nope"]}`, `field emails[0]: must be a valid email`},
		{"declared map value", `{"byName": {"a": "nope"}}`, `field ["a"]: must be a valid uuid`},
		{"declared map of arrays", `{"listsByName": {"a": ["nope"]}}`, `field ["a"][0]: must be a valid uuid`},
		{"declared array item", `{"idList": ["nope"]}`, `field [0]: must be a valid uuid`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v testElements.Elements

			require.ErrorContains(t, json.Unmarshal([]byte(tc.payload), &v), tc.want)
		})
	}

	t.Run("yaml shares the check", func(t *testing.T) {
		t.Parallel()

		var v testElements.Elements

		require.ErrorContains(t, yamlv3.Unmarshal([]byte("ids: [nope]\n"), &v), `field ids[0]: must be a valid uuid`)
	})

	t.Run("inline map value", func(t *testing.T) {
		t.Parallel()

		var v testInlineMaps.InlineMaps

		require.NoError(t, json.Unmarshal([]byte(`{"byName": {"a": "`+u+`"}}`), &v))
		require.ErrorContains(t, json.Unmarshal([]byte(`{"byName": {"a": "nope"}}`), &v),
			`field byName["a"]: must be a valid uuid`)
	})
}
