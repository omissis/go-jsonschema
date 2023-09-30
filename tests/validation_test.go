package tests_test

import (
	"encoding/json"
	"errors"
	"testing"

	test "github.com/atombender/go-jsonschema/tests/data/validation"
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
			desc:    "myString too long",
			data:    `{"myString": "hello"}`,
			wantErr: errors.New("field myString length: must be <= 5"),
		},
		{
			desc:    "myString not present",
			data:    `{}`,
			wantErr: errors.New("field myString in A631MaxLength: required"),
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
		tC := tC

		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := test.A631MaxLength{}

			err := json.Unmarshal([]byte(tC.data), &model)

			if tC.wantErr == nil && err != nil {
				t.Errorf("got error %v, want nil", err)
			} else if tC.wantErr != nil && err == nil {
				t.Errorf("got nil, want error %v", tC.wantErr)
			} else if tC.wantErr != nil && err != nil && err.Error() != tC.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, tC.wantErr)
			}
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
			wantErr: errors.New("field myString in A632MinLength: required"),
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
		tC := tC

		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			model := test.A632MinLength{}

			err := json.Unmarshal([]byte(tC.data), &model)

			if tC.wantErr == nil && err != nil {
				t.Errorf("got error %v, want nil", err)
			} else if tC.wantErr != nil && err == nil {
				t.Errorf("got nil, want error %v", tC.wantErr)
			} else if tC.wantErr != nil && err != nil && err.Error() != tC.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, tC.wantErr)
			}
		})
	}
}
