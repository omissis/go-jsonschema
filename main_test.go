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

func TestSplitPackageAlias(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		input     string
		wantPkg   string
		wantAlias string
		wantErr   error
	}{
		{name: "no colon", input: "example.com/foo/v1", wantPkg: "example.com/foo/v1"},
		{name: "with valid alias", input: "example.com/foo/v1:foov1", wantPkg: "example.com/foo/v1", wantAlias: "foov1"},
		{name: "alias with underscore", input: "example.com/foo:foo_v1", wantPkg: "example.com/foo", wantAlias: "foo_v1"},
		{name: "last colon wins", input: "ssh://example.com/foo:alias", wantPkg: "ssh://example.com/foo", wantAlias: "alias"},
		{name: "empty alias rejected", input: "example.com/foo:", wantErr: errInvalidImportAlias},
		{name: "alias starting with digit rejected", input: "example.com/foo:1bad", wantErr: errInvalidImportAlias},
		{name: "alias with hyphen rejected", input: "example.com/foo:bad-alias", wantErr: errInvalidImportAlias},
		{name: "alias matching keyword rejected", input: "example.com/foo:type", wantErr: errInvalidImportAlias},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pkg, alias, err := splitPackageAlias(tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if pkg != tc.wantPkg || alias != tc.wantAlias {
				t.Fatalf("got (%q, %q), want (%q, %q)", pkg, alias, tc.wantPkg, tc.wantAlias)
			}
		})
	}
}

func TestLoadKnownSchemas(t *testing.T) {
	t.Parallel()

	yamlExts := []string{".yml", ".yaml"}

	t.Run("nil entries returns empty cache", func(t *testing.T) {
		t.Parallel()

		got, err := loadKnownSchemas(nil, yamlExts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == nil {
			t.Fatal("expected non-nil empty cache, got nil")
		}

		if len(got) != 0 {
			t.Fatalf("expected empty cache, got %v", got)
		}
	})

	t.Run("missing flag separator rejected", func(t *testing.T) {
		t.Parallel()

		_, err := loadKnownSchemas([]string{"no-equals-sign"}, yamlExts)
		if !errors.Is(err, errFlagFormat) {
			t.Fatalf("expected errFlagFormat, got %v", err)
		}
	})

	t.Run("missing file rejected with errKnownSchemaLoad", func(t *testing.T) {
		t.Parallel()

		_, err := loadKnownSchemas(
			[]string{"https://example.com/x=/no/such/path/that/exists.json"},
			yamlExts,
		)
		if !errors.Is(err, errKnownSchemaLoad) {
			t.Fatalf("expected errKnownSchemaLoad, got %v", err)
		}
	})

	t.Run("duplicate URL rejected", func(t *testing.T) {
		t.Parallel()

		_, err := loadKnownSchemas([]string{
			"https://example.com/dup=tests/data/knownSchema/canonical/canonical.json",
			"https://example.com/dup=tests/data/knownSchema/canonical/canonical.json",
		}, yamlExts)
		if !errors.Is(err, errKnownSchemaDuplicate) {
			t.Fatalf("expected errKnownSchemaDuplicate, got %v", err)
		}
	})

	t.Run("happy path loads file", func(t *testing.T) {
		t.Parallel()

		const url = "https://example.com/canonical/v1/canonical.json"

		cache, err := loadKnownSchemas(
			[]string{url + "=tests/data/knownSchema/canonical/canonical.json"},
			yamlExts,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := cache[url]; !ok {
			t.Fatalf("expected cache to contain URL %q, got %v", url, cache)
		}
	})

	t.Run("yaml file routed through FromYAMLFile", func(t *testing.T) {
		t.Parallel()

		const url = "https://example.com/canonical/v1/canonical.yaml"

		cache, err := loadKnownSchemas(
			[]string{url + "=tests/data/knownSchema/canonical/canonical.yaml"},
			yamlExts,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sc, ok := cache[url]
		if !ok {
			t.Fatalf("expected cache to contain URL %q, got %v", url, cache)
		}

		// Sanity-check the YAML actually parsed by inspecting one decoded
		// field — confirms FromYAMLFile (not FromJSONFile) was invoked, since
		// the .yaml fixture isn't valid JSON.
		if sc.ObjectAsType == nil || sc.Properties == nil {
			t.Fatalf("expected Properties to be populated from YAML, got %+v", sc)
		}

		if _, ok := sc.Properties["messageId"]; !ok {
			t.Fatalf("expected `messageId` property to be parsed from YAML, got %v", sc.Properties)
		}
	})
}
