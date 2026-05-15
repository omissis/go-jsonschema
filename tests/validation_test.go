package tests_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	testExclusiveMaximum "github.com/tuotuoxp/go-jsonschema/tests/data/validation/exclusiveMaximum"
	testExclusiveMinimum "github.com/tuotuoxp/go-jsonschema/tests/data/validation/exclusiveMinimum"
	testMaxLength "github.com/tuotuoxp/go-jsonschema/tests/data/validation/maxLength"
	testMaximum "github.com/tuotuoxp/go-jsonschema/tests/data/validation/maximum"
	testMinLength "github.com/tuotuoxp/go-jsonschema/tests/data/validation/minLength"
	testMinimum "github.com/tuotuoxp/go-jsonschema/tests/data/validation/minimum"
	testMultipleOf "github.com/tuotuoxp/go-jsonschema/tests/data/validation/multipleOf"
	testPattern "github.com/tuotuoxp/go-jsonschema/tests/data/validation/pattern"
	testPrimitiveDefs "github.com/tuotuoxp/go-jsonschema/tests/data/validation/primitive_defs"
	testReadOnlyFields "github.com/tuotuoxp/go-jsonschema/tests/data/validation/readOnly"
	testReadOnlyAndRequiredFields "github.com/tuotuoxp/go-jsonschema/tests/data/validation/readOnlyAndRequired"
	testRequiredFields "github.com/tuotuoxp/go-jsonschema/tests/data/validation/requiredFields"
	testReadOnlyValidationDisabledFields "github.com/tuotuoxp/go-jsonschema/tests/data/validationDisabled/readOnly"
	"github.com/tuotuoxp/go-jsonschema/tests/helpers"
)

func TestMaxStringLength(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myString": "hi"}`,
			wantErr: nil,
		},
		{
			desc:    "myString has the max allowed length",
			data:    `{"myString": "hello"}`,
			wantErr: nil,
		},
		{
			desc:    "myString too long",
			data:    `{"myString": "hello world"}`,
			wantErr: errors.New("field myString length: must be <= 5"),
		},
		{
			desc:    "myString not present",
			data:    `{}`,
			wantErr: errors.New("field myString in MaxLength: required"),
		},
		{
			desc:    "myNullableString too long",
			data:    `{"myString": "hi","myNullableString": "hello world"}`,
			wantErr: errors.New("field myNullableString length: must be <= 10"),
		},
		{
			desc:    "myString and myNullableString too long",
			data:    `{"myString": "hello","myNullableString": "hello world"}`,
			wantErr: errors.New("field myNullableString length: must be <= 10"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testMaxLength.MaxLength{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestMinStringLength(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myString": "hello"}`,
			wantErr: nil,
		},
		{
			desc:    "myString too short",
			data:    `{"myString": "hi"}`,
			wantErr: errors.New("field myString length: must be >= 5"),
		},
		{
			desc:    "myString not present",
			data:    `{}`,
			wantErr: errors.New("field myString in MinLength: required"),
		},
		{
			desc:    "myNullableString too short",
			data:    `{"myString": "hello","myNullableString": "hi"}`,
			wantErr: errors.New("field myNullableString length: must be >= 10"),
		},
		{
			desc:    "myString and myNullableString too short",
			data:    `{"myString": "hi","myNullableString": "hello"}`,
			wantErr: errors.New("field myNullableString length: must be >= 10"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testMinLength.MinLength{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestRequiredFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "object without required property fails validation",
			data:    `{}`,
			wantErr: errors.New("field myNullableObject in RequiredNullable: required"),
		},
		{
			desc:    "required properties may be null",
			data:    `{ "myNullableObject": null, "myNullableStringArray": null, "myNullableString": null }`,
			wantErr: nil,
		},
		{
			desc:    "required properties may have a non-null value",
			data:    `{ "myNullableObject": { "myNestedProp": "foobar" }, "myNullableStringArray": ["hello"], "myNullableString": "world" }`,
			wantErr: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testRequiredFields.RequiredNullable{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestReadOnlyFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "object without readOnly property passes validation",
			data:    `{"myString": "abc"}`,
			wantErr: nil,
		},
		{
			desc:    "object with readOnly property fails validation",
			data:    `{"myString": "abc", "myReadOnlyString": "abc"}`,
			wantErr: errors.New("field myReadOnlyString in ReadOnly: read only"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testReadOnlyFields.ReadOnly{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestReadOnlyFieldsNoValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "object with readOnly property and disabled validation",
			data:    `{"myString": "abc", "myReadOnlyString": "abc"}`,
			wantErr: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testReadOnlyValidationDisabledFields.ReadOnlyNoValidation{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestReadOnlyAndRequiredFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "object with property that is both required and readOnly fails validation if given",
			data:    `{"myReadOnlyRequiredString": "abc"}`,
			wantErr: errors.New("field myReadOnlyRequiredString in ReadOnlyAndRequired: read only"),
		},
		{
			desc:    "object with property that is both required and readOnly fails validation if not given",
			data:    `{}`,
			wantErr: errors.New("field myReadOnlyRequiredString in ReadOnlyAndRequired: required"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testReadOnlyAndRequiredFields.ReadOnlyAndRequired{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestPattern(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myString": "0x12345abcde."}`,
		},
		{
			desc:    "myString does not match pattern",
			data:    `{"myString": "0x123456"}`,
			wantErr: errors.New("field MyString pattern match: must match ^0x[0-9a-f]{10}\\.$"),
		},
		{
			desc: "no violations",
			data: `{"myEscapedString": "${{SOME_VAR}}", "myString": "0x12345abcde."}`,
		},
		{
			desc:    "myEscapedString does not match pattern",
			data:    `{"myEscapedString": "${MISSING_CURLY_BRACKET}", "myString": "0x12345abcde."}`,
			wantErr: errors.New("field MyEscapedString pattern match: must match ^\\$\\{\\{(.|[\\r\\n])*\\}\\}$"),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testPattern.Pattern{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestPrimitiveDefs(t *testing.T) {
	t.Parallel()

	basePayload := map[string]any{
		"inlinePatternString":  "abc.example",
		"refPatternString":     "def.example",
		"inlineBoundedString":  "abcd",
		"refBoundedString":     "bcde",
		"inlineConstString":    "stable",
		"refConstString":       "stable",
		"inlineBoundedInteger": 5,
		"refBoundedInteger":    6,
		"inlineBoundedNumber":  3.5,
		"refBoundedNumber":     3.5,
		"inlineConstBoolean":   true,
		"refConstBoolean":      true,
	}

	cloneBase := func() map[string]any {
		out := make(map[string]any, len(basePayload))
		for key, val := range basePayload {
			out[key] = val
		}

		return out
	}

	toJSON := func(t *testing.T, m map[string]any) []byte {
		t.Helper()

		value, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}

		return value
	}

	testCases := []struct {
		desc           string
		mutate         func(payload map[string]any)
		wantErrContain string
	}{
		{
			desc:           "failure/inline_pattern",
			mutate:         func(payload map[string]any) { payload["inlinePatternString"] = "INVALID" },
			wantErrContain: "pattern match",
		},
		{
			desc:           "failure/ref_pattern",
			mutate:         func(payload map[string]any) { payload["refPatternString"] = "INVALID" },
			wantErrContain: "pattern match",
		},
		{
			desc:           "failure/inline_string_length",
			mutate:         func(payload map[string]any) { payload["inlineBoundedString"] = "ab" },
			wantErrContain: "must be >=",
		},
		{
			desc:           "failure/ref_string_length",
			mutate:         func(payload map[string]any) { payload["refBoundedString"] = "ab" },
			wantErrContain: "must be >=",
		},
		{
			desc:           "failure/inline_string_const",
			mutate:         func(payload map[string]any) { payload["inlineConstString"] = "unstable" },
			wantErrContain: "must be equal",
		},
		{
			desc:           "failure/ref_string_const",
			mutate:         func(payload map[string]any) { payload["refConstString"] = "unstable" },
			wantErrContain: "must be equal",
		},
		{
			desc:           "failure/inline_integer_range",
			mutate:         func(payload map[string]any) { payload["inlineBoundedInteger"] = 10 },
			wantErrContain: "must be <",
		},
		{
			desc:           "failure/ref_integer_range",
			mutate:         func(payload map[string]any) { payload["refBoundedInteger"] = 10 },
			wantErrContain: "must be <",
		},
		{
			desc:           "failure/inline_number_const",
			mutate:         func(payload map[string]any) { payload["inlineBoundedNumber"] = 4.5 },
			wantErrContain: "must be equal",
		},
		{
			desc:           "failure/ref_number_const",
			mutate:         func(payload map[string]any) { payload["refBoundedNumber"] = 4.5 },
			wantErrContain: "must be equal",
		},
		{
			desc:           "failure/inline_boolean_const",
			mutate:         func(payload map[string]any) { payload["inlineConstBoolean"] = false },
			wantErrContain: "must be equal to true",
		},
		{
			desc:           "failure/ref_boolean_const",
			mutate:         func(payload map[string]any) { payload["refConstBoolean"] = false },
			wantErrContain: "must be equal to true",
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var prim testPrimitiveDefs.PrimitiveDefs
		if err := json.Unmarshal(toJSON(t, cloneBase()), &prim); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			payload := cloneBase()
			tC.mutate(payload)

			var prim testPrimitiveDefs.PrimitiveDefs
			err := json.Unmarshal(toJSON(t, payload), &prim)
			if err == nil {
				t.Fatalf("expected error containing %q", tC.wantErrContain)
			}
			if !strings.Contains(err.Error(), tC.wantErrContain) {
				t.Fatalf("expected error containing %q, got %q", tC.wantErrContain, err.Error())
			}
		})
	}
}

func TestMultipleOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myInteger": 10, "myNumber": 2.4}`,
		},
		{
			desc: "no violations bigger number",
			data: `{"myInteger": 10, "myNumber": 482.6}`,
		},
		{
			desc:    "myInt not a multiple of 2",
			data:    `{"myInteger": 11, "myNumber": 2.4}`,
			wantErr: errors.New("field myInteger: must be a multiple of 2"),
		},
		{
			desc:    "myNumber not a multiple of 1.2",
			data:    `{"myInteger": 10, "myNumber": 2.5}`,
			wantErr: errors.New("field myNumber: must be a multiple of 0.2"),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			mo := testMultipleOf.MultipleOf{}

			err := json.Unmarshal([]byte(tC.data), &mo)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestMaximum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myInteger": 1, "myNumber": 1.0}`,
		},
		{
			desc:    "myInt exceeds maximum of 2",
			data:    `{"myInteger": 3, "myNumber": 1.0}`,
			wantErr: errors.New("field myInteger: must be <= 2"),
		},
		{
			desc:    "myNumber exceeds maximum of 1.2",
			data:    `{"myInteger": 1, "myNumber": 1.3}`,
			wantErr: errors.New("field myNumber: must be <= 1.2"),
		},
		{
			desc: "boundary case - exactly at maximum",
			data: `{"myInteger": 2, "myNumber": 1.2}`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			mo := testMaximum.Maximum{}

			err := json.Unmarshal([]byte(tC.data), &mo)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestMinimum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myInteger": 3, "myNumber": 1.5}`,
		},
		{
			desc:    "myInt below minimum of 2",
			data:    `{"myInteger": 1, "myNumber": 1.5}`,
			wantErr: errors.New("field myInteger: must be >= 2"),
		},
		{
			desc:    "myNumber below minimum of 1.2",
			data:    `{"myInteger": 3, "myNumber": 1.1}`,
			wantErr: errors.New("field myNumber: must be >= 1.2"),
		},
		{
			desc: "boundary case - exactly at minimum",
			data: `{"myInteger": 2, "myNumber": 1.2}`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			mo := testMinimum.Minimum{}

			err := json.Unmarshal([]byte(tC.data), &mo)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestExclusiveMaximum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myInteger": 1, "myNumber": 1.1}`,
		},
		{
			desc:    "myInt exceeds exclusive maximum of 2",
			data:    `{"myInteger": 2, "myNumber": 1.1}`,
			wantErr: errors.New("field myInteger: must be < 2"),
		},
		{
			desc:    "myNumber exceeds exclusive maximum of 1.2",
			data:    `{"myInteger": 1, "myNumber": 1.2}`,
			wantErr: errors.New("field myNumber: must be < 1.2"),
		},
		{
			desc: "boundary case - just below exclusive maximum",
			data: `{"myInteger": 1, "myNumber": 1.19}`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			mo := testExclusiveMaximum.ExclusiveMaximum{}

			err := json.Unmarshal([]byte(tC.data), &mo)

			helpers.CheckError(t, tC.wantErr, err)

			mo2 := testExclusiveMaximum.ExclusiveMaximumOld{}

			err = json.Unmarshal([]byte(tC.data), &mo2)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestExclusiveMinimum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myInteger": 3, "myNumber": 1.3}`,
		},
		{
			desc:    "myInt below exclusive minimum of 2",
			data:    `{"myInteger": 2, "myNumber": 1.3}`,
			wantErr: errors.New("field myInteger: must be > 2"),
		},
		{
			desc:    "myNumber below exclusive minimum of 1.2",
			data:    `{"myInteger": 3, "myNumber": 1.2}`,
			wantErr: errors.New("field myNumber: must be > 1.2"),
		},
		{
			desc: "boundary case - just above exclusive minimum",
			data: `{"myInteger": 3, "myNumber": 1.21}`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			mo := testExclusiveMinimum.ExclusiveMinimum{}

			err := json.Unmarshal([]byte(tC.data), &mo)

			helpers.CheckError(t, tC.wantErr, err)

			mo2 := testExclusiveMinimum.ExclusiveMinimumOld{}

			err = json.Unmarshal([]byte(tC.data), &mo2)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}
