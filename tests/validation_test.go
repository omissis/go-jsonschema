package tests_test

import (
	"encoding/json"
	"errors"
	"testing"

	testExclusiveMaximum "github.com/atombender/go-jsonschema/tests/data/validation/exclusiveMaximum"
	testExclusiveMinimum "github.com/atombender/go-jsonschema/tests/data/validation/exclusiveMinimum"
	testMaxLength "github.com/atombender/go-jsonschema/tests/data/validation/maxLength"
	testMaximum "github.com/atombender/go-jsonschema/tests/data/validation/maximum"
	testMinLength "github.com/atombender/go-jsonschema/tests/data/validation/minLength"
	testMinimum "github.com/atombender/go-jsonschema/tests/data/validation/minimum"
	testMultipleOf "github.com/atombender/go-jsonschema/tests/data/validation/multipleOf"
	testPattern "github.com/atombender/go-jsonschema/tests/data/validation/pattern"
	testPrimitiveDefs "github.com/atombender/go-jsonschema/tests/data/validation/primitive_defs"
	testRequiredFields "github.com/atombender/go-jsonschema/tests/data/validation/requiredFields"
	"github.com/atombender/go-jsonschema/tests/helpers"
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
			data:    `{ "myNullableObject": { "myNestedProp": "世界" }, "myNullableStringArray": ["hello"], "myNullableString": "world" }`,
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

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc: "no violations",
			data: `{"myString": "hello"}`,
		},
		{
			desc:    "myString too short",
			data:    `{"myString": "hi"}`,
			wantErr: errors.New("field  length: must be >= 5"),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			prim := testPrimitiveDefs.PrimitiveDefs{}

			err := json.Unmarshal([]byte(tC.data), &prim)

			helpers.CheckError(t, tC.wantErr, err)
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
