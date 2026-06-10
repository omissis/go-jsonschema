package tests_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConditionalDiscriminatorDeclineWarnings asserts that the
// conditional-discriminator detector emits a named warning when an
// enum-form `if` clause is structurally close to the canonical shape but
// fails one of the validation rules (non-string values, mixed types,
// empty enum, branch value not in parent enum). Pre-existing decline
// reasons (e.g., outer-if extra constraints) keep their silent-decline
// behavior — only NEW reasons added by the enum extension are surfaced
// here, to keep the existing const-form test surface stable.
func TestConditionalDiscriminatorDeclineWarnings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		desc          string
		schemaPath    string
		wantSubstring string
	}{
		{
			desc:          "numeric enum",
			schemaPath:    "./data/conditionalDiscriminator/enumIfNumeric/enumIfNumeric.json",
			wantSubstring: "is not a string",
		},
		{
			desc:          "mixed-type enum",
			schemaPath:    "./data/conditionalDiscriminator/enumIfMixedTypes/enumIfMixedTypes.json",
			wantSubstring: "is not a string",
		},
		{
			desc:          "empty enum",
			schemaPath:    "./data/conditionalDiscriminator/enumIfEmpty/enumIfEmpty.json",
			wantSubstring: "enum is empty",
		},
		{
			desc:          "branch value not in parent enum",
			schemaPath:    "./data/conditionalDiscriminator/enumIfBranchValueNotInParentEnum/enumIfBranchValueNotInParentEnum.json",
			wantSubstring: `not a member of the "kind" enum`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			warnings := generateWithWarnerCapture(t, tc.schemaPath)
			joined := strings.Join(warnings, "\n")
			assert.Contains(t, joined, tc.wantSubstring,
				"expected a warning containing %q in:\n%s", tc.wantSubstring, joined)
		})
	}
}

// TestConditionalDiscriminatorSilentDeclines asserts that pre-existing
// silent-decline reasons (extra constraints on the if clause, or no
// const/enum at all) remain silent — no warning emitted. Same fidelity
// gap p9's warner would have caught is the only signal a user gets, and
// that's intentional.
func TestConditionalDiscriminatorSilentDeclines(t *testing.T) {
	t.Parallel()

	t.Run("extra required on if clause is silently declined", func(t *testing.T) {
		t.Parallel()

		warnings := generateWithWarnerCapture(t,
			"./data/conditionalDiscriminator/enumIfWithExtraRequiredOnIf/enumIfWithExtraRequiredOnIf.json")
		joined := strings.Join(warnings, "\n")
		assert.NotContains(t, joined, "conditional-discriminator detection declined",
			"extra-constraints on if clause should fall through silently; got:\n%s", joined)
	})
}
