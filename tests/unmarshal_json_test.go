package tests_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	testOneOfDiscAnimal "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/animal"
	testOneOfDiscNumeric "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/numericKind"
	testOneOfRefDisc "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/refDiscriminator"
	testOneOfInField "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/inField"
	testOneOfNumStringBool "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/numStringBool"
	testOneOfWithNull "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/withNull"
	testStrictAddlFalse "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlFalse"
	testStrictAddlFalseEmpty "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlFalseEmpty"
	testStrictAddlOmitted "github.com/atombender/go-jsonschema/tests/data/strictAdditionalProperties/addlOmitted"
	testStrictAlwaysAddlOmittedEmpty "github.com/atombender/go-jsonschema/tests/data/strictAdditionalPropertiesAlways/addlOmittedEmpty"
	testStrictAlwaysBasic "github.com/atombender/go-jsonschema/tests/data/strictAdditionalPropertiesAlways/basic"
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

func TestJsonUnmarshalStrictAdditionalProperties(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc      string
		json      string
		target    json.Unmarshaler
		expectErr bool
	}{
		// respect-schema mode + schema declares false
		{
			desc:   "addlFalse accepts only declared",
			json:   `{"name":"Alice","age":30}`,
			target: &testStrictAddlFalse.AddlFalse{},
		},
		{
			desc:      "addlFalse rejects extra",
			json:      `{"name":"Alice","age":30,"unexpected":true}`,
			target:    &testStrictAddlFalse.AddlFalse{},
			expectErr: true,
		},
		{
			desc:   "addlFalse minimal",
			json:   `{"name":"Bob"}`,
			target: &testStrictAddlFalse.AddlFalse{},
		},

		// respect-schema mode + schema is silent: must NOT enforce
		{
			desc:   "addlOmitted ignores extras",
			json:   `{"label":"x","extra":1,"more":"yes"}`,
			target: &testStrictAddlOmitted.AddlOmitted{},
		},

		// strict mode (always)
		{
			desc:   "always basic accepts known",
			json:   `{"kind":"foo"}`,
			target: &testStrictAlwaysBasic.Basic{},
		},
		{
			desc:      "always basic rejects unknown",
			json:      `{"kind":"foo","oops":42}`,
			target:    &testStrictAlwaysBasic.Basic{},
			expectErr: true,
		},

		// property-less object schemas: respect-schema mode + addlProps:false
		{
			desc:   "addlFalseEmpty accepts empty object",
			json:   `{}`,
			target: &testStrictAddlFalseEmpty.AddlFalseEmpty{},
		},
		{
			desc:      "addlFalseEmpty rejects any key",
			json:      `{"foo":1}`,
			target:    &testStrictAddlFalseEmpty.AddlFalseEmpty{},
			expectErr: true,
		},

		// property-less object schemas: strict mode applies even when schema is silent
		{
			desc:   "alwaysEmpty accepts empty object",
			json:   `{}`,
			target: &testStrictAlwaysAddlOmittedEmpty.AddlOmittedEmpty{},
		},
		{
			desc:      "alwaysEmpty rejects any key",
			json:      `{"foo":1}`,
			target:    &testStrictAlwaysAddlOmittedEmpty.AddlOmittedEmpty{},
			expectErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			err := tC.target.UnmarshalJSON([]byte(tC.json))
			if tC.expectErr {
				assert.Error(t, err, "expected error but got nil")
			} else {
				assert.NoError(t, err, "did not expect error")
			}
		})
	}
}

// TestStrictFieldsErrorIsDeterministic verifies the strict-fields validator
// reports all unknown keys (not just one) and in deterministic sorted order.
// Without sorting, Go's randomized map iteration would surface a different
// key (or set order) on each run.
func TestStrictFieldsErrorIsDeterministic(t *testing.T) {
	t.Parallel()

	jsonInput := []byte(`{"name":"Alice","zeta":1,"alpha":2,"mu":3}`)

	// Run repeatedly to make sure map iteration randomness doesn't trip us up.
	for range 50 {
		var v testStrictAddlFalse.AddlFalse

		err := v.UnmarshalJSON(jsonInput)
		require.Error(t, err, "expected error for unknown keys")
		// Sorted order should always be alpha, mu, zeta.
		assert.Contains(t, err.Error(), `["alpha" "mu" "zeta"]`,
			"unknown keys should be sorted in error message")
	}
}

func TestJsonUnmarshalOneOfPrimitive(t *testing.T) {
	t.Parallel()

	parseInField := func(t *testing.T, payload string) testOneOfInField.InField {
		t.Helper()

		var v testOneOfInField.InField
		if err := json.Unmarshal([]byte(payload), &v); err != nil {
			t.Fatalf("unmarshal: %s", err)
		}

		return v
	}

	t.Run("string variant accepted", func(t *testing.T) {
		t.Parallel()

		v := parseInField(t, `{"name":"a","value":"hello"}`)
		s, ok := v.Value.AsString()
		assert.True(t, ok)
		assert.Equal(t, "hello", s)

		_, gotNum := v.Value.AsNumber()
		assert.False(t, gotNum)
	})

	t.Run("number variant accepted", func(t *testing.T) {
		t.Parallel()

		v := parseInField(t, `{"name":"a","value":42.5}`)
		n, ok := v.Value.AsNumber()
		assert.True(t, ok)
		assert.InDelta(t, 42.5, n, 1e-9)
	})

	t.Run("bool variant accepted", func(t *testing.T) {
		t.Parallel()

		v := parseInField(t, `{"name":"a","value":true}`)
		b, ok := v.Value.AsBool()
		assert.True(t, ok)
		assert.True(t, b)
	})

	t.Run("null is rejected because schema does not allow it", func(t *testing.T) {
		t.Parallel()

		var v testOneOfInField.InField
		assert.Error(t, json.Unmarshal([]byte(`{"name":"a","value":null}`), &v))
	})

	t.Run("array is rejected for primitive variant", func(t *testing.T) {
		t.Parallel()

		var v testOneOfInField.InField
		assert.Error(t, json.Unmarshal([]byte(`{"name":"a","value":[]}`), &v))
	})

	t.Run("round trip preserves value", func(t *testing.T) {
		t.Parallel()

		v := parseInField(t, `{"name":"x","value":7}`)
		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"x","value":7}`, string(out))
	})

	t.Run("string|null variant accepts null", func(t *testing.T) {
		t.Parallel()

		var v testOneOfWithNull.WithNull
		require.NoError(t, json.Unmarshal([]byte(`{"value":null}`), &v))
		assert.True(t, v.Value.IsNull())
	})

	t.Run("string|null variant accepts string", func(t *testing.T) {
		t.Parallel()

		var v testOneOfWithNull.WithNull
		require.NoError(t, json.Unmarshal([]byte(`{"value":"x"}`), &v))

		s, ok := v.Value.AsString()
		assert.True(t, ok)
		assert.Equal(t, "x", s)
	})

	t.Run("string|null variant rejects number", func(t *testing.T) {
		t.Parallel()

		var v testOneOfWithNull.WithNull
		assert.Error(t, json.Unmarshal([]byte(`{"value":7}`), &v))
	})

	// Marshal/Unmarshal round-trip symmetry around the unset/null case.
	// Schemas without a `null` variant must refuse to marshal an unset
	// wrapper (rather than emit `null` that the matching UnmarshalJSON
	// would reject). Schemas WITH a `null` variant continue to round-trip
	// `null` ↔ unset cleanly.

	t.Run("non-nullable wrapper: marshal unset value errors", func(t *testing.T) {
		t.Parallel()

		// numStringBool's oneOf is [number, string, boolean] — no null.
		// A zero NumStringBoolValue (no variant set) must NOT marshal to
		// null because UnmarshalJSON would reject that on the way back.
		v := testOneOfNumStringBool.NumStringBool{}

		_, err := json.Marshal(&v.Value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "schema does not allow null")
	})

	t.Run("nullable wrapper: round-trip preserves explicit null", func(t *testing.T) {
		t.Parallel()

		// withNull's oneOf is [string, null] — null is a legal variant.
		input := []byte(`{"value":null}`)

		var v testOneOfWithNull.WithNull

		require.NoError(t, json.Unmarshal(input, &v))
		assert.True(t, v.Value.IsNull())

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, string(input), string(out))
	})

	t.Run("IsZero is safe on a nil receiver", func(t *testing.T) {
		t.Parallel()

		// Pre-fix this panicked; the nil-receiver guard added in this
		// commit makes IsZero return true for nil pointers.
		var v *testOneOfNumStringBool.NumStringBoolValue

		assert.True(t, v.IsZero())
	})

	// presence-tracking three-state distinction: unset vs explicit null
	// vs decoded primitive. Pre-fix the wrapper conflated unset and null
	// via `value == nil`, so IsZero returned true for explicit-null
	// values and IsNull returned true for unset values. With the
	// `present` discriminator the three states are observable.

	t.Run("unset wrapper: IsZero=true, IsNull=false", func(t *testing.T) {
		t.Parallel()

		var v testOneOfWithNull.WithNullValue

		assert.True(t, v.IsZero())
		assert.False(t, v.IsNull())
	})

	t.Run("explicit null: IsZero=false, IsNull=true", func(t *testing.T) {
		t.Parallel()

		// Pre-fix IsZero returned true here, which would cause a host
		// struct using `omitzero` on this field to drop the explicit
		// null on remarshal — silent data loss. With present-tracking
		// IsZero correctly distinguishes null from unset.
		var v testOneOfWithNull.WithNull

		require.NoError(t, json.Unmarshal([]byte(`{"value":null}`), &v))
		assert.False(t, v.Value.IsZero(), "explicit null must not be IsZero (would break omitzero)")
		assert.True(t, v.Value.IsNull())
	})

	t.Run("decoded primitive: IsZero=false, IsNull=false", func(t *testing.T) {
		t.Parallel()

		var v testOneOfWithNull.WithNull

		require.NoError(t, json.Unmarshal([]byte(`{"value":"hello"}`), &v))
		assert.False(t, v.Value.IsZero())
		assert.False(t, v.Value.IsNull())
	})

	t.Run("nullable wrapper: unset marshals as null", func(t *testing.T) {
		t.Parallel()

		// A wrapper that's never been Unmarshaled into still serializes
		// as JSON null when its schema includes `null` as a variant. The
		// receiver's surrounding struct can opt out via `omitzero`.
		var v testOneOfWithNull.WithNullValue

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.Equal(t, "null", string(out))
	})

	t.Run("non-nullable wrapper: marshal unset errors", func(t *testing.T) {
		t.Parallel()

		// A wrapper for a schema that does NOT include `null` as a
		// variant must error on marshal when never populated, since
		// emitting `null` would violate the schema. Counterpart of the
		// nullable case above where unset → "null" is allowed.
		//
		// Note: the related `present && value == nil` guard in
		// MarshalJSON cannot currently be exercised from outside the
		// generator package — `value` and `present` are unexported and
		// no public setter API exists. The guard is defensive against a
		// future setter being added; it'd need an in-package test (or
		// an `unsafe.Pointer`-based reflection set) to cover.
		v := testOneOfNumStringBool.NumStringBoolValue{}

		_, err := json.Marshal(&v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "schema does not allow null")
	})
}

// TestJsonUnmarshalOneOfDiscriminated covers the Phase-5 dispatch path: a
// holder type whose UnmarshalJSON peeks a `kind`/`version` discriminator and
// fans out into the matching variant. Goldens cover code shape;
// these cases cover runtime dispatch correctness which goldens cannot.
func TestJsonUnmarshalOneOfDiscriminated(t *testing.T) {
	t.Parallel()

	t.Run("animal: dog variant decoded into Dog field", func(t *testing.T) {
		t.Parallel()

		var v testOneOfDiscAnimal.Animal
		require.NoError(t, json.Unmarshal(
			[]byte(`{"creature":{"kind":"dog","barkAt":"the postman"}}`),
			&v,
		))
		require.NotNil(t, v.Creature.Dog)
		assert.Nil(t, v.Creature.Cat)
		assert.Equal(t, "the postman", v.Creature.Dog.BarkAt)
	})

	t.Run("animal: cat variant decoded into Cat field", func(t *testing.T) {
		t.Parallel()

		var v testOneOfDiscAnimal.Animal
		require.NoError(t, json.Unmarshal(
			[]byte(`{"creature":{"kind":"cat","purr":true}}`),
			&v,
		))
		require.NotNil(t, v.Creature.Cat)
		assert.Nil(t, v.Creature.Dog)
		assert.True(t, v.Creature.Cat.Purr)
	})

	t.Run("animal: missing discriminator rejected", func(t *testing.T) {
		t.Parallel()

		var v testOneOfDiscAnimal.Animal

		err := json.Unmarshal([]byte(`{"creature":{"barkAt":"x"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing discriminator")
	})

	t.Run("animal: unknown discriminator value rejected", func(t *testing.T) {
		t.Parallel()

		var v testOneOfDiscAnimal.Animal

		err := json.Unmarshal([]byte(`{"creature":{"kind":"fish","fins":2}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown kind value")
	})

	t.Run("animal: round-trip dog preserves variant", func(t *testing.T) {
		t.Parallel()

		input := []byte(`{"creature":{"kind":"dog","barkAt":"squirrels"}}`)

		var v testOneOfDiscAnimal.Animal

		require.NoError(t, json.Unmarshal(input, &v))

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, string(input), string(out))
	})

	t.Run("animal: marshal with no variant set errors", func(t *testing.T) {
		t.Parallel()

		v := testOneOfDiscAnimal.Animal{}
		_, err := json.Marshal(&v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one variant")
	})

	t.Run("animal: marshal with two variants set errors", func(t *testing.T) {
		t.Parallel()

		v := testOneOfDiscAnimal.Animal{
			Creature: testOneOfDiscAnimal.AnimalCreature{
				Dog: &testOneOfDiscAnimal.AnimalCreatureDog{Kind: "dog", BarkAt: "x"},
				Cat: &testOneOfDiscAnimal.AnimalCreatureCat{Kind: "cat", Purr: true},
			},
		}
		_, err := json.Marshal(&v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one variant")
	})

	t.Run("numericKind: integer discriminator dispatch", func(t *testing.T) {
		t.Parallel()

		var v1 testOneOfDiscNumeric.NumericKind
		require.NoError(t, json.Unmarshal(
			[]byte(`{"payload":{"version":1,"alpha":"a"}}`),
			&v1,
		))
		require.NotNil(t, v1.Payload.Const1)
		assert.Equal(t, "a", v1.Payload.Const1.Alpha)

		var v2 testOneOfDiscNumeric.NumericKind
		require.NoError(t, json.Unmarshal(
			[]byte(`{"payload":{"version":2,"beta":"b"}}`),
			&v2,
		))
		require.NotNil(t, v2.Payload.Const2)
		assert.Equal(t, "b", v2.Payload.Const2.Beta)
	})

	t.Run("numericKind: equivalent numeric forms (1.0, 1e0) match the same variant", func(t *testing.T) {
		t.Parallel()

		// JSON allows 1, 1.0, and 1e0 as equivalent encodings of the integer 1.
		// The discriminator dispatch must accept all three and route to
		// Const1 (which has `version: 1` as its discriminator const).
		for _, version := range []string{"1", "1.0", "1e0"} {
			var v testOneOfDiscNumeric.NumericKind
			require.NoErrorf(t,
				json.Unmarshal([]byte(`{"payload":{"version":`+version+`,"alpha":"a"}}`), &v),
				"version=%s", version,
			)
			require.NotNilf(t, v.Payload.Const1, "version=%s", version)
			assert.Equalf(t, "a", v.Payload.Const1.Alpha, "version=%s", version)
		}
	})

	t.Run("numericKind: non-numeric discriminator returns typed error", func(t *testing.T) {
		t.Parallel()

		var v testOneOfDiscNumeric.NumericKind

		err := json.Unmarshal(
			[]byte(`{"payload":{"version":"not-a-number","alpha":"a"}}`),
			&v,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version discriminator must be numeric")
	})

	// $ref/allOf flattening: variants are `{"allOf":[{"$ref":Common},{...const-bearing inline...}]}`.
	// Without flattening, primitiveConstCandidates would see only the
	// outer allOf wrapper (no Properties / Required at the top level)
	// and the discriminator detection would fail. With flattening the
	// discriminator-bearing const is recognised and dispatch works.

	t.Run("refDiscriminator: allOf+$ref variant decoded into Click", func(t *testing.T) {
		t.Parallel()

		var v testOneOfRefDisc.RefDiscriminator

		require.NoError(t, json.Unmarshal(
			[]byte(`{"event":{"kind":"click","target":"#submit","timestamp":"2026-05-08T00:00:00Z"}}`),
			&v,
		))
		require.NotNil(t, v.Event.Click)
		assert.Nil(t, v.Event.Scroll)
		assert.Equal(t, "#submit", v.Event.Click.Target)
		require.NotNil(t, v.Event.Click.Timestamp)
		assert.Equal(t, "2026-05-08T00:00:00Z", *v.Event.Click.Timestamp)
	})

	t.Run("refDiscriminator: allOf+$ref variant decoded into Scroll", func(t *testing.T) {
		t.Parallel()

		var v testOneOfRefDisc.RefDiscriminator

		require.NoError(t, json.Unmarshal(
			[]byte(`{"event":{"kind":"scroll","distance":120}}`),
			&v,
		))
		require.NotNil(t, v.Event.Scroll)
		assert.Nil(t, v.Event.Click)
		assert.Equal(t, 120, v.Event.Scroll.Distance)
	})

	t.Run("refDiscriminator: round-trip preserves variant + inherited fields", func(t *testing.T) {
		t.Parallel()

		input := []byte(`{"event":{"kind":"click","target":"#x","timestamp":"t"}}`)

		var v testOneOfRefDisc.RefDiscriminator

		require.NoError(t, json.Unmarshal(input, &v))

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, string(input), string(out))
	})
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
