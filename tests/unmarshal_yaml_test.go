package tests_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"

	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
	testFormatYAMLAll "github.com/atombender/go-jsonschema/tests/data/formatValidation/all"
	testFormatYAMLEmail "github.com/atombender/go-jsonschema/tests/data/formatValidation/email"
	testFormatYAMLHostname "github.com/atombender/go-jsonschema/tests/data/formatValidation/hostname"
	testFormatYAMLRegex "github.com/atombender/go-jsonschema/tests/data/formatValidation/regex"
	testFormatYAMLURI "github.com/atombender/go-jsonschema/tests/data/formatValidation/uri"
	testFormatYAMLURIRef "github.com/atombender/go-jsonschema/tests/data/formatValidation/uriReference"
	testFormatYAMLUUID "github.com/atombender/go-jsonschema/tests/data/formatValidation/uuid"
	testYAMLStrictAddlFalse "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlFalse"
	testYAMLStrictAddlFalseEmpty "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlFalseEmpty"
	testYAMLStrictAddlOmitted "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlOmitted"
	testYAMLStrictAlwaysBasic "github.com/atombender/go-jsonschema/tests/data/strictAdditionalPropertiesAlways/basic"
)

func TestYamlV3Unmarshal(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("./data/extraImports/gopkgYAMLv3/gopkgYAMLv3.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var conf test.GopkgYAMLv3

	if err := yamlv3.Unmarshal(data, &conf); err != nil {
		t.Fatal(err)
	}

	s := "example"
	n := 123.456
	i := 123
	b := true
	e := test.GopkgYAMLv3MyEnumX

	want := test.GopkgYAMLv3{
		MyString:  &s,
		MyNumber:  &n,
		MyInteger: &i,
		MyBoolean: &b,
		MyNull:    nil,
		MyEnum:    &e,
	}

	if !reflect.DeepEqual(conf, want) {
		t.Errorf(
			"Unmarshalled data does not match expected\nWant: %s\nGot:  %s",
			formatGopkgYAMLv3(want),
			formatGopkgYAMLv3(conf),
		)
	}
}

func TestYamlV3UnmarshalInvalidEnum(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("./data/extraImports/gopkgYAMLv3invalidEnum/gopkgYAMLv3invalidEnum.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var conf test.GopkgYAMLv3

	err = yamlv3.Unmarshal(data, &conf)
	if err == nil {
		t.Fatal("Expected unmarshal error")
	}

	if !strings.Contains(err.Error(), "invalid value (expected one of") {
		t.Error("Expected unmarshal error to contain enum values")
	}
}

func TestYamlUnmarshalFormatValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc      string
		yaml      string
		target    any
		expectErr bool
	}{
		// uuid
		{desc: "uuid valid", yaml: "id: 550e8400-e29b-41d4-a716-446655440000\n", target: &testFormatYAMLUUID.Uuid{}},
		{desc: "uuid invalid", yaml: "id: not-a-uuid\n", target: &testFormatYAMLUUID.Uuid{}, expectErr: true},
		{
			desc:   "uuid optional absent ok",
			yaml:   "id: 550e8400-e29b-41d4-a716-446655440000\n",
			target: &testFormatYAMLUUID.Uuid{},
		},
		{
			desc:   "uuid optional present and valid",
			yaml:   "id: 550e8400-e29b-41d4-a716-446655440000\nparentId: 6ba7b810-9dad-11d1-80b4-00c04fd430c8\n",
			target: &testFormatYAMLUUID.Uuid{},
		},
		{
			desc:      "uuid optional invalid rejected",
			yaml:      "id: 550e8400-e29b-41d4-a716-446655440000\nparentId: oops\n",
			target:    &testFormatYAMLUUID.Uuid{},
			expectErr: true,
		},

		// email
		{desc: "email valid", yaml: "primary: user@example.com\n", target: &testFormatYAMLEmail.Email{}},
		{desc: "email invalid", yaml: "primary: not-an-email\n", target: &testFormatYAMLEmail.Email{}, expectErr: true},
		{
			desc:      "email rejects display-name form",
			yaml:      "primary: \"Alice <alice@example.com>\"\n",
			target:    &testFormatYAMLEmail.Email{},
			expectErr: true,
		},

		// uri
		{desc: "uri valid absolute", yaml: "endpoint: https://example.com/path\n", target: &testFormatYAMLURI.Uri{}},
		{
			desc:      "uri rejected when not absolute",
			yaml:      "endpoint: /just/a/path\n",
			target:    &testFormatYAMLURI.Uri{},
			expectErr: true,
		},
		{
			desc:      "uri rejected with whitespace",
			yaml:      "endpoint: \"https://example.com/path with space\"\n",
			target:    &testFormatYAMLURI.Uri{},
			expectErr: true,
		},

		// uri-reference
		{desc: "uri-reference valid relative", yaml: "ref: /relative/path\n", target: &testFormatYAMLURIRef.UriReference{}},
		{
			desc:      "uri-reference rejects malformed pct-encoding",
			yaml:      "ref: /path%ZZ\n",
			target:    &testFormatYAMLURIRef.UriReference{},
			expectErr: true,
		},

		// hostname
		{desc: "hostname valid", yaml: "host: example.com\n", target: &testFormatYAMLHostname.Hostname{}},
		{desc: "hostname rejected underscore", yaml: "host: bad_host.com\n", target: &testFormatYAMLHostname.Hostname{}, expectErr: true},

		// regex
		{desc: "regex valid", yaml: "pattern: \"^[a-z]+$\"\n", target: &testFormatYAMLRegex.Regex{}},
		{desc: "regex invalid", yaml: "pattern: \"[\"\n", target: &testFormatYAMLRegex.Regex{}, expectErr: true},

		// combined
		{
			desc:   "all formats valid",
			yaml:   "userId: 550e8400-e29b-41d4-a716-446655440000\ncontact: a@b.com\nhomepage: https://x\nicon: icon.png\nhost: a.b\nrule: \"^x$\"\n",
			target: &testFormatYAMLAll.All{},
		},
		{
			desc:      "all schema rejects bad email",
			yaml:      "userId: 550e8400-e29b-41d4-a716-446655440000\ncontact: oops\n",
			target:    &testFormatYAMLAll.All{},
			expectErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			err := yamlv3.Unmarshal([]byte(tC.yaml), tC.target)
			if tC.expectErr {
				assert.Error(t, err, "expected validation error but got nil")
			} else {
				assert.NoError(t, err, "did not expect error")
			}
		})
	}
}

// TestYamlUnmarshalStrictAdditionalProperties mirrors
// TestJsonUnmarshalStrictAdditionalProperties but exercises the YAML path,
// which goes through the same shared unmarshal-body pipeline. A divergence
// in the yaml-tag stripping logic or in unknown-field rejection would
// otherwise pass undetected with only the JSON tests.
func TestYamlUnmarshalStrictAdditionalProperties(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc      string
		yaml      string
		target    any
		expectErr bool
	}{
		// respect-schema mode + schema declares false
		{desc: "addlFalse accepts only declared", yaml: "name: Alice\nage: 30\n", target: &testYAMLStrictAddlFalse.AddlFalse{}},
		{
			desc:      "addlFalse rejects extra",
			yaml:      "name: Alice\nage: 30\nunexpected: true\n",
			target:    &testYAMLStrictAddlFalse.AddlFalse{},
			expectErr: true,
		},

		// respect-schema mode + schema is silent: must NOT enforce
		{
			desc:   "addlOmitted ignores extras",
			yaml:   "label: x\nextra: 1\nmore: yes\n",
			target: &testYAMLStrictAddlOmitted.AddlOmitted{},
		},

		// strict mode (always)
		{desc: "always basic accepts known", yaml: "kind: foo\n", target: &testYAMLStrictAlwaysBasic.Basic{}},
		{
			desc:      "always basic rejects unknown",
			yaml:      "kind: foo\noops: 42\n",
			target:    &testYAMLStrictAlwaysBasic.Basic{},
			expectErr: true,
		},

		// property-less object schemas: respect-schema mode + addlProps:false
		{desc: "addlFalseEmpty accepts empty mapping", yaml: "{}\n", target: &testYAMLStrictAddlFalseEmpty.AddlFalseEmpty{}},
		{
			desc:      "addlFalseEmpty rejects any key",
			yaml:      "foo: 1\n",
			target:    &testYAMLStrictAddlFalseEmpty.AddlFalseEmpty{},
			expectErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			err := yamlv3.Unmarshal([]byte(tC.yaml), tC.target)
			if tC.expectErr {
				assert.Error(t, err, "expected validation error but got nil")
			} else {
				assert.NoError(t, err, "did not expect error")
			}
		})
	}
}
