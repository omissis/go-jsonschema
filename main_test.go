package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

func TestParseStrictAdditionalProperties(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		input   string
		want    generator.StrictAdditionalPropertiesMode
		wantErr error
	}{
		{name: "empty is off", input: "", want: generator.StrictAdditionalPropertiesOff},
		{name: "off explicit", input: "off", want: generator.StrictAdditionalPropertiesOff},
		{name: "off case-insensitive", input: "OFF", want: generator.StrictAdditionalPropertiesOff},
		{name: "respect-schema", input: "respect-schema", want: generator.StrictAdditionalPropertiesRespectSchema},
		{name: "strict", input: "strict", want: generator.StrictAdditionalPropertiesStrict},
		{name: "strict trims whitespace", input: "  strict  ", want: generator.StrictAdditionalPropertiesStrict},
		{name: "typo rejected", input: "rstrict", wantErr: errInvalidStrictAddlPropMode},
		{name: "garbage rejected", input: "yes-please", wantErr: errInvalidStrictAddlPropMode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseStrictAdditionalProperties(tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseValidateFormats(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		input   string
		want    generator.FormatValidationConfig
		wantErr error
	}{
		{name: "empty disables", input: "", want: generator.FormatValidationConfig{}},
		{name: "off disables", input: "off", want: generator.FormatValidationConfig{}},
		{name: "off case-insensitive", input: "OFF", want: generator.FormatValidationConfig{}},
		{name: "all enables without allow list", input: "all", want: generator.FormatValidationConfig{Enabled: true}},
		{name: "subset", input: "uuid,email", want: generator.FormatValidationConfig{Enabled: true, AllowList: []string{"uuid", "email"}}},
		{name: "subset trims and lowercases", input: " UUID , Email ", want: generator.FormatValidationConfig{Enabled: true, AllowList: []string{"uuid", "email"}}},
		{name: "unknown name rejected", input: "uuid,emial", wantErr: errUnknownFormatKeyword},
		{name: "empty list entry rejected", input: "uuid,,email", wantErr: errEmptyFormatListEntry},
		{name: "all mixed with name rejected", input: "all,uuid", wantErr: errFormatListWithKeyword},
		{name: "off mixed with name rejected", input: "off,uuid", wantErr: errFormatListWithKeyword},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseValidateFormats(tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
