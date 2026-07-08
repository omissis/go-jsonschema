package tests_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	testAdditionalProperties "github.com/atombender/go-jsonschema/tests/data/core/additionalProperties"
	testAllOf "github.com/atombender/go-jsonschema/tests/data/core/allOf"
	testAnyOf "github.com/atombender/go-jsonschema/tests/data/core/anyOf"
	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
	testFormatAll "github.com/atombender/go-jsonschema/tests/data/formatValidation/all"
	testFormatEmail "github.com/atombender/go-jsonschema/tests/data/formatValidation/email"
	testFormatHostname "github.com/atombender/go-jsonschema/tests/data/formatValidation/hostname"
	testFormatRegex "github.com/atombender/go-jsonschema/tests/data/formatValidation/regex"
	testFormatURI "github.com/atombender/go-jsonschema/tests/data/formatValidation/uri"
	testFormatURIRef "github.com/atombender/go-jsonschema/tests/data/formatValidation/uriReference"
	testFormatUUID "github.com/atombender/go-jsonschema/tests/data/formatValidation/uuid"
	testValudationRequiredFields "github.com/atombender/go-jsonschema/tests/data/validation/requiredFields"
)

func TestJsonUnmarshalValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		json     string
		target   any
		assertFn func(target any)
	}{
		{
			desc:   "requiredFields - nullable",
			json:   `{"myNullableObject": null, "myNullableString": null, "myNullableStringArray": null}`,
			target: &testValudationRequiredFields.RequiredNullable{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testValudationRequiredFields.RequiredNullable{
						MyNullableObject:      nil,
						MyNullableString:      nil,
						MyNullableStringArray: nil,
					},
					target,
				)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			if err := json.Unmarshal([]byte(tC.json), tC.target); err != nil {
				t.Fatalf("unmarshal error: %s", err)
			}

			tC.assertFn(tC.target)
		})
	}
}

func TestJsonUmarshalAnyOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		json     string
		target   any
		assertFn func(target any)
	}{
		{
			desc: "anyOf.1 - 1",
			json: `{
				"configurations": [
					{
						"foo": "hello"
					},
					{
						"bar": 2.2
					},
					{
						"baz": true
					}
				],
				"flags": "hello"

			}`,
			target: &testAnyOf.AnyOf1{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf1{
						Configurations: []testAnyOf.AnyOf1ConfigurationsElem{
							{Foo: ptr("hello")},
							{Bar: ptr(2.2)},
							{Baz: ptr(true)},
						},
						Flags: "hello",
					},
					target,
				)
			},
		},
		{
			desc: "anyOf.1 - 2",
			json: `{
				"configurations": [
					{
						"foo": "ciao"
					},
					{
						"bar": 200
					}
				],
				"flags": true

			}`,
			target: &testAnyOf.AnyOf1{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf1{
						Configurations: []testAnyOf.AnyOf1ConfigurationsElem{
							{Foo: ptr("ciao")},
							{Bar: ptr(200.0)},
						},
						Flags: true,
					},
					target,
				)
			},
		},
		{
			desc: "anyOf.2 - 1",
			json: `{
				"configurations": [
					{
						"foo": "ciao"
					},
					{
						"bar": 2
					},
					{
						"baz": false
					}
				]
			}`,
			target: &testAnyOf.AnyOf2{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf2{
						Configurations: []testAnyOf.AnyOf2ConfigurationsElem{
							{Foo: ptr("ciao")},
							{Bar: ptr(2.0)},
							{Baz: ptr(false)},
						},
					},
					target,
				)
			},
		},
		{
			desc:   "anyOf.3 - 1",
			json:   `{"foo": "ciao"}`,
			target: &testAnyOf.AnyOf3{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf3{
						Foo: ptr("ciao"),
					},
					target,
				)
			},
		},
		{
			desc:   "anyOf.3 - 2",
			json:   `{"bar": 2.0}`,
			target: &testAnyOf.AnyOf3{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf3{
						Bar: ptr(2.0),
					},
					target,
				)
			},
		},
		{
			desc:   "anyOf.3 - 3",
			json:   `{"configurations": ["ciao"]}`,
			target: &testAnyOf.AnyOf3{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf3{
						Configurations: []any{"ciao"},
					},
					target,
				)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			if err := json.Unmarshal([]byte(tC.json), tC.target); err != nil {
				t.Fatalf("unmarshal error: %s", err)
			}

			tC.assertFn(tC.target)
		})
	}
}

func TestJsonUmarshalAllOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		json     string
		target   any
		assertFn func(target any)
	}{
		{
			desc: "allOf.1 - 1",
			json: `{
				"configurations": [
					{
						"foo": "hello",
						"bar": 2.2
					}
				]
			}`,
			target: &testAllOf.AllOf1{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAllOf.AllOf1{
						Configurations: []testAllOf.AllOf1ConfigurationsElem{
							{Foo: "hello", Bar: 2.2},
						},
					},
					target,
				)
			},
		},
		{
			desc: "allOf.2 - 1",
			json: `{
				"configurations": [
					{
						"foo": "hello",
						"bar": 2.2,
						"baz": true
					}
				]
			}`,
			target: &testAllOf.AllOf2{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAllOf.AllOf2{
						Configurations: []testAllOf.AllOf2ConfigurationsElem{
							{Foo: "hello", Bar: 2.2, Baz: ptr(true)},
						},
					},
					target,
				)
			},
		},
		{
			desc: "allOf.3 - 1",
			json: `{
				"foo": "hello",
				"bar": 2.2,
				"configurations": ["ciao"]
			}`,
			target: &testAllOf.AllOf3{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAllOf.AllOf3{Foo: "hello", Bar: 2.2, Configurations: []any{"ciao"}},
					target,
				)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			if err := json.Unmarshal([]byte(tC.json), tC.target); err != nil {
				t.Fatalf("unmarshal error: %s", err)
			}

			tC.assertFn(tC.target)
		})
	}
}

func TestJSONUnmarshalAdditionalProperties(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		json     string
		target   json.Unmarshaler
		assertFn func(target json.Unmarshaler)
	}{
		{
			desc: "array",
			json: `{
				"name": "hello world",
				"property1": ["one", "two"],
				"property2": [3, 4]
			}`,
			target: &testAdditionalProperties.ArrayAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.ArrayAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string][]any{"property1": {"one", "two"}, "property2": {3.0, 4.0}}, addProps)
			},
		},
		{
			desc: "bool",
			json: `{
				"name": "hello world",
				"property1": true,
				"property2": false
			}`,
			target: &testAdditionalProperties.BoolAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.BoolAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string]bool{"property1": true, "property2": false}, addProps)
			},
		},
		{
			desc: "int",
			json: `{
				"name": "hello world",
				"property1": 1,
				"property2": 2
			}`,
			target: &testAdditionalProperties.IntAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.IntAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string]int{"property1": 1, "property2": 2}, addProps)
			},
		},
		{
			desc: "number",
			json: `{
				"name": "hello world",
				"property1": 1.1,
				"property2": 2.3
			}`,
			target: &testAdditionalProperties.NumberAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.NumberAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string]float64{"property1": 1.1, "property2": 2.3}, addProps)
			},
		},
		{
			desc: "object",
			json: `{
				"name": "hello world",
				"surname": {
					"hello": 1.1,
					"world": "what's up?"
				}
			}`,
			target: &testAdditionalProperties.ObjectAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.ObjectAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string]any{"surname": map[string]any{"hello": 1.1, "world": "what's up?"}}, addProps)
			},
		},
		{
			desc: "object with props",
			json: `{
				"foo": "foo value",
				"bar": "bar value",
				"baz": {
					"property1": "hello",
					"property2": 123
				}
			}`,
			target: &testAdditionalProperties.ObjectWithPropsAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.ObjectWithPropsAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string]any{"baz": map[string]any{"property1": "hello", "property2": 123.0}}, addProps)
			},
		},
		{
			desc: "string",
			json: `{
				"name": "hello world",
				"property1": "hello",
				"property2": "world"
			}`,
			target: &testAdditionalProperties.StringAdditionalProperties{},
			assertFn: func(target json.Unmarshaler) {
				addProps := target.(*testAdditionalProperties.StringAdditionalProperties).AdditionalProperties

				assert.Equal(t, map[string]string{"property1": "hello", "property2": "world"}, addProps)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			if err := tC.target.UnmarshalJSON([]byte(tC.json)); err != nil {
				t.Fatalf("unmarshal error: %s", err)
			}

			tC.assertFn(tC.target)
		})
	}
}

func TestJsonUnmarshalFormatValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc      string
		json      string
		target    json.Unmarshaler
		expectErr bool
	}{
		// uuid
		{desc: "uuid valid", json: `{"id":"550e8400-e29b-41d4-a716-446655440000"}`, target: &testFormatUUID.Uuid{}},
		{desc: "uuid valid uppercase", json: `{"id":"550E8400-E29B-41D4-A716-446655440000"}`, target: &testFormatUUID.Uuid{}},
		{desc: "uuid invalid", json: `{"id":"not-a-uuid"}`, target: &testFormatUUID.Uuid{}, expectErr: true},
		{desc: "uuid invalid length", json: `{"id":"550e8400-e29b-41d4-a716-44665544000"}`, target: &testFormatUUID.Uuid{}, expectErr: true},
		{desc: "uuid empty rejected", json: `{"id":""}`, target: &testFormatUUID.Uuid{}, expectErr: true},
		{
			desc:   "uuid optional absent ok",
			json:   `{"id":"550e8400-e29b-41d4-a716-446655440000"}`,
			target: &testFormatUUID.Uuid{},
		},
		{
			desc:   "uuid optional present and valid",
			json:   `{"id":"550e8400-e29b-41d4-a716-446655440000","parentId":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`,
			target: &testFormatUUID.Uuid{},
		},
		{
			desc:      "uuid optional invalid rejected",
			json:      `{"id":"550e8400-e29b-41d4-a716-446655440000","parentId":"oops"}`,
			target:    &testFormatUUID.Uuid{},
			expectErr: true,
		},

		// email (RFC 5321 addr-spec only)
		{desc: "email valid", json: `{"primary":"user@example.com"}`, target: &testFormatEmail.Email{}},
		{desc: "email invalid", json: `{"primary":"not-an-email"}`, target: &testFormatEmail.Email{}, expectErr: true},
		{desc: "email rejects display-name form", json: `{"primary":"Alice <alice@example.com>"}`, target: &testFormatEmail.Email{}, expectErr: true},
		{desc: "email rejects bracketed addr-spec", json: `{"primary":"<bob@example.com>"}`, target: &testFormatEmail.Email{}, expectErr: true},
		{desc: "email empty rejected", json: `{"primary":""}`, target: &testFormatEmail.Email{}, expectErr: true},
		{
			desc:   "email optional nil ok",
			json:   `{"primary":"user@example.com"}`,
			target: &testFormatEmail.Email{},
		},
		{
			desc:   "email optional present and valid",
			json:   `{"primary":"user@example.com","secondary":"alt@example.com"}`,
			target: &testFormatEmail.Email{},
		},
		{
			desc:      "email optional invalid rejected",
			json:      `{"primary":"user@example.com","secondary":"oops"}`,
			target:    &testFormatEmail.Email{},
			expectErr: true,
		},
		{
			desc:      "email optional rejects display-name",
			json:      `{"primary":"user@example.com","secondary":"Alice <alice@example.com>"}`,
			target:    &testFormatEmail.Email{},
			expectErr: true,
		},

		// uri (absolute required)
		{desc: "uri valid absolute", json: `{"endpoint":"https://example.com/path"}`, target: &testFormatURI.Uri{}},
		{desc: "uri rejected when not absolute", json: `{"endpoint":"/just/a/path"}`, target: &testFormatURI.Uri{}, expectErr: true},
		{desc: "uri rejected with whitespace", json: `{"endpoint":"https://example.com/path with space"}`, target: &testFormatURI.Uri{}, expectErr: true},
		{desc: "uri rejected with control char", json: `{"endpoint":"https://example.com/\u0007"}`, target: &testFormatURI.Uri{}, expectErr: true},
		{desc: "uri rejected when empty", json: `{"endpoint":""}`, target: &testFormatURI.Uri{}, expectErr: true},
		{desc: "uri valid with pct-encoded path", json: `{"endpoint":"https://example.com/a%20b"}`, target: &testFormatURI.Uri{}},

		// uri-reference (empty is valid per RFC 3986; rejects whitespace and bad pct-encoding)
		{desc: "uri-reference valid relative", json: `{"ref":"/relative/path"}`, target: &testFormatURIRef.UriReference{}},
		{desc: "uri-reference valid absolute", json: `{"ref":"https://example.com"}`, target: &testFormatURIRef.UriReference{}},
		{desc: "uri-reference valid empty (same-document ref)", json: `{"ref":""}`, target: &testFormatURIRef.UriReference{}},
		{desc: "uri-reference valid fragment only", json: `{"ref":"#section"}`, target: &testFormatURIRef.UriReference{}},
		{desc: "uri-reference rejects whitespace", json: `{"ref":"hello world"}`, target: &testFormatURIRef.UriReference{}, expectErr: true},
		{desc: "uri-reference rejects malformed pct-encoding", json: `{"ref":"/path%ZZ"}`, target: &testFormatURIRef.UriReference{}, expectErr: true},
		{desc: "uri-reference rejects lone percent", json: `{"ref":"/path%"}`, target: &testFormatURIRef.UriReference{}, expectErr: true},

		// hostname (RFC 1123)
		{desc: "hostname valid", json: `{"host":"example.com"}`, target: &testFormatHostname.Hostname{}},
		{desc: "hostname valid single label", json: `{"host":"localhost"}`, target: &testFormatHostname.Hostname{}},
		{desc: "hostname rejected leading dot", json: `{"host":".example.com"}`, target: &testFormatHostname.Hostname{}, expectErr: true},
		{desc: "hostname rejected underscore", json: `{"host":"bad_host.com"}`, target: &testFormatHostname.Hostname{}, expectErr: true},
		{desc: "hostname rejected empty", json: `{"host":""}`, target: &testFormatHostname.Hostname{}, expectErr: true},
		{
			// 64 valid labels of "abcd" joined by dots = 64*4 + 63 = 319 chars total; each label is fine but total > 253.
			desc:      "hostname rejected when exceeds 253 octets",
			json:      `{"host":"` + strings.Repeat("abcd.", 63) + `abcd"}`,
			target:    &testFormatHostname.Hostname{},
			expectErr: true,
		},

		// regex (Go's RE2 syntax — empty regex compiles to match-anything)
		{desc: "regex valid", json: `{"pattern":"^[a-z]+$"}`, target: &testFormatRegex.Regex{}},
		{desc: "regex invalid", json: `{"pattern":"["}`, target: &testFormatRegex.Regex{}, expectErr: true},
		{desc: "regex empty accepted", json: `{"pattern":""}`, target: &testFormatRegex.Regex{}},

		// combined schema
		{
			desc:   "all formats valid",
			json:   `{"userId":"550e8400-e29b-41d4-a716-446655440000","contact":"a@b.com","homepage":"https://x","icon":"icon.png","host":"a.b","rule":"^x$"}`,
			target: &testFormatAll.All{},
		},
		{
			desc:      "all schema rejects bad email",
			json:      `{"userId":"550e8400-e29b-41d4-a716-446655440000","contact":"oops"}`,
			target:    &testFormatAll.All{},
			expectErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			err := tC.target.UnmarshalJSON([]byte(tC.json))
			if tC.expectErr {
				assert.Error(t, err, "expected validation error but got nil")
			} else {
				assert.NoError(t, err, "did not expect error")
			}
		})
	}
}

func formatGopkgYAMLv3(v test.GopkgYAMLv3) string {
	return fmt.Sprintf(
		"GopkgYAMLv3{MyString: %s, MyNumber: %f, MyInteger: %d, MyBoolean: %t, MyNull: %v, MyEnum: %v}",
		*v.MyString, *v.MyNumber, *v.MyInteger, *v.MyBoolean, nil, *v.MyEnum,
	)
}

func ptr[T any](v T) *T {
	return &v
}
