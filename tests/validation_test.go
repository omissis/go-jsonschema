package tests_test

import (
	"encoding/json"
	"errors"
	"testing"

	testMaxLength "github.com/atombender/go-jsonschema/tests/data/validation/maxLength"
	testMinLength "github.com/atombender/go-jsonschema/tests/data/validation/minLength"
	testPattern "github.com/atombender/go-jsonschema/tests/data/validation/pattern"
	testNumericRange "github.com/atombender/go-jsonschema/tests/data/validation/numericRange"
	testMinimum "github.com/atombender/go-jsonschema/tests/data/validation/minimum"
	testMaximum "github.com/atombender/go-jsonschema/tests/data/validation/maximum"
	testExclusiveMinimum "github.com/atombender/go-jsonschema/tests/data/validation/exclusiveMinimum"
	testExclusiveMaximum "github.com/atombender/go-jsonschema/tests/data/validation/exclusiveMaximum"
	testUniqueItems "github.com/atombender/go-jsonschema/tests/data/validation/uniqueItems"
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

func TestStringPattern(t *testing.T) {
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
			desc:    "myString pattern not matching",
			data:    `{"myString": "hi2"}`,
			wantErr: errors.New("field myString pattern does not match: ^([a-z][a-z_]*)$"),
		},
		{
			desc:    "myString not present",
			data:    `{}`,
			wantErr: errors.New("field myString in Pattern: required"),
		},
		{
			desc:    "myNullableString pattern not matching",
			data:    `{"myString": "hello","myNullableString": "hi4"}`,
			wantErr: errors.New("field myNullableString pattern does not match: ^([a-z][a-z_]*)$"),
		},
		{
			desc:    "myString and myNullableString not matching",
			data:    `{"myString": "hi1","myNullableString": "hello2"}`,
			wantErr: errors.New("field myNullableString pattern does not match: ^([a-z][a-z_]*)$"),
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

func TestNumericRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myInteger": 4, "myNumber": 4.221}`,
			wantErr: nil,
		},
		{
			desc:    "myInteger violation low",
			data:    `{"myInteger": 0, "myNumber": 6.2}`,
			wantErr: errors.New("field myInteger: violates mimimum 2"),
		},
		{
			desc:    "myInteger violation high",
			data:    `{"myInteger": 23, "myNumber": 6.2}`,
			wantErr: errors.New("field myInteger: violates maximum 5"),
		},
		{
			desc:    "myNumber violation low",
			data:    `{"myInteger": 3, "myNumber": 1.28}`,
			wantErr: errors.New("field myNumber: violates mimimum 2.4"),
		},
		{
			desc:    "myNumber violation high",
			data:    `{"myInteger": 3, "myNumber": 7.28}`,
			wantErr: errors.New("field myNumber: violates maximum 5.6"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testNumericRange.NumericRange{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestNumericMimimum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myInteger": 6, "myNumber": 6.2}`,
			wantErr: nil,
		},
		{
			desc:    "myInteger violation",
			data:    `{"myInteger": 4, "myNumber": 6.2}`,
			wantErr: errors.New("field myInteger: violates mimimum 5"),
		},
		{
			desc:    "myNumber violation",
			data:    `{"myInteger": 6, "myNumber": 4.2}`,
			wantErr: errors.New("field myNumber: violates mimimum 5.6"),
		},
		{
			desc:    "myNullableInteger violation",
			data:    `{"myInteger": 6, "myNumber": 6.2, "myNullableInteger": 7}`,
			wantErr: errors.New("field myNullableInteger: violates mimimum 10"),
		},
		{
			desc:    "myNullableNumber violation",
			data:    `{"myInteger": 6, "myNumber": 6.2, "myNullableNumber": 8.4}`,
			wantErr: errors.New("field myNullableNumber: violates mimimum 10.3"),
		},
		{
			desc:    "myInteger and myNullableInteger violation",
			data:    `{"myInteger": 4, "myNumber": 6.2, "myNullableInteger": 7}`,
			wantErr: errors.New("field myInteger: violates mimimum 5"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testMinimum.Minimum{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestNumericMaximum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myInteger": 4, "myNumber": 4.2}`,
			wantErr: nil,
		},
		{
			desc:    "myInteger violation",
			data:    `{"myInteger": 6, "myNumber": 4.2}`,
			wantErr: errors.New("field myInteger: violates maximum 5"),
		},
		{
			desc:    "myNumber violation",
			data:    `{"myInteger": 4, "myNumber": 6.2}`,
			wantErr: errors.New("field myNumber: violates maximum 5.6"),
		},
		{
			desc:    "myNullableInteger violation",
			data:    `{"myInteger": 4, "myNumber": 4.2, "myNullableInteger": 11}`,
			wantErr: errors.New("field myNullableInteger: violates maximum 10"),
		},
		{
			desc:    "myNullableNumber violation",
			data:    `{"myInteger": 4, "myNumber": 4.2, "myNullableNumber": 12.4}`,
			wantErr: errors.New("field myNullableNumber: violates maximum 10.3"),
		},
		{
			desc:    "myInteger and myNullableInteger violation",
			data:    `{"myInteger": 6, "myNumber": 4.2, "myNullableInteger": 14}`,
			wantErr: errors.New("field myInteger: violates maximum 5"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testMaximum.Maximum{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestNumericExclusiveMimimum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myInteger": 6, "myNumber": 6.2}`,
			wantErr: nil,
		},
		{
			desc:    "myInteger violation",
			data:    `{"myInteger": 5, "myNumber": 6.2}`,
			wantErr: errors.New("field myInteger: violates exclusiveMinimum 5"),
		},
		{
			desc:    "myNumber violation",
			data:    `{"myInteger": 6, "myNumber": 5.6}`,
			wantErr: errors.New("field myNumber: violates exclusiveMinimum 5.6"),
		},
		{
			desc:    "myNullableInteger violation",
			data:    `{"myInteger": 6, "myNumber": 6.2, "myNullableInteger": 10}`,
			wantErr: errors.New("field myNullableInteger: violates exclusiveMinimum 10"),
		},
		{
			desc:    "myNullableNumber violation",
			data:    `{"myInteger": 6, "myNumber": 6.2, "myNullableNumber": 10.3}`,
			wantErr: errors.New("field myNullableNumber: violates exclusiveMinimum 10.3"),
		},
		{
			desc:    "myInteger and myNullableInteger violation",
			data:    `{"myInteger": 4, "myNumber": 6.2, "myNullableInteger": 7}`,
			wantErr: errors.New("field myInteger: violates exclusiveMinimum 5"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testExclusiveMinimum.ExclusiveMinimum{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestNumericExclusiveMaximum(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myInteger": 4, "myNumber": 4.2}`,
			wantErr: nil,
		},
		{
			desc:    "myInteger violation",
			data:    `{"myInteger": 5, "myNumber": 4.2}`,
			wantErr: errors.New("field myInteger: violates exclusiveMaximum 5"),
		},
		{
			desc:    "myNumber violation",
			data:    `{"myInteger": 4, "myNumber": 5.6}`,
			wantErr: errors.New("field myNumber: violates exclusiveMaximum 5.6"),
		},
		{
			desc:    "myNullableInteger violation",
			data:    `{"myInteger": 4, "myNumber": 4.2, "myNullableInteger": 10}`,
			wantErr: errors.New("field myNullableInteger: violates exclusiveMaximum 10"),
		},
		{
			desc:    "myNullableNumber violation",
			data:    `{"myInteger": 4, "myNumber": 4.2, "myNullableNumber": 10.3}`,
			wantErr: errors.New("field myNullableNumber: violates exclusiveMaximum 10.3"),
		},
		{
			desc:    "myInteger and myNullableInteger violation",
			data:    `{"myInteger": 6, "myNumber": 4.2, "myNullableInteger": 14}`,
			wantErr: errors.New("field myInteger: violates exclusiveMaximum 5"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testExclusiveMaximum.ExclusiveMaximum{}

			err := json.Unmarshal([]byte(tC.data), &model)

			helpers.CheckError(t, tC.wantErr, err)
		})
	}
}

func TestArrayUniqueItems(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc    string
		data    string
		wantErr error
	}{
		{
			desc:    "no violations",
			data:    `{"myStringArray": ["a", "b", "c"]}`,
			wantErr: nil,
		},
		{
			desc:    "myStringArray violation",
			data:    `{"myStringArray": ["a", "b", "a"]}`,
			wantErr: errors.New("field myStringArray: items must be unique"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := testUniqueItems.UniqueItems{}

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
