package tests_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testCondDiscBasic "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/basic"
	testCondDiscEnumIf "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/enumIf"
	testCondDiscConstAndEnum "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/enumIfConstAndEnum"
	testCondDiscManyValues "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/enumIfManyValues"
	testCondDiscMixed "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/enumIfMixedWithConst"
	testCondDiscOverlap "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/enumIfOverlapping"
	testCondDiscEnumElse "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/enumIfWithElse"
	testCondDiscWithElse "github.com/atombender/go-jsonschema/tests/data/conditionalDiscriminator/withElse"
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
	testOneOfNoDisc "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/noDiscriminator"
	testOneOfDiscNumeric "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/numericKind"
	testOneOfRefDisc "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/refDiscriminator"
	testOneOfStrictShape "github.com/atombender/go-jsonschema/tests/data/oneOfDiscriminated/strictShape"
	testOneOfInField "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/inField"
	testOneOfNumStringBool "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/numStringBool"
	testOneOfWithNull "github.com/atombender/go-jsonschema/tests/data/oneOfPrimitive/withNull"
	testRecursiveAllOfSelfRef "github.com/atombender/go-jsonschema/tests/data/recursiveAllOf/selfRef"
	testRootAllOf "github.com/atombender/go-jsonschema/tests/data/rootComposition/rootAllOf"
	testRootEnum "github.com/atombender/go-jsonschema/tests/data/rootComposition/rootEnum"
	testRootOneOfDisc "github.com/atombender/go-jsonschema/tests/data/rootComposition/rootOneOfDiscriminator"
	testRootOneOfTryEach "github.com/atombender/go-jsonschema/tests/data/rootComposition/rootOneOfTryEach"
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

// TestJsonUnmarshalOneOfTryEach covers Phase 6: oneOf without a natural
// discriminator falls back to per-variant try-each with optional
// shape-compatibility filtering. Goldens cover code shape; these cases
// cover runtime dispatch correctness.
func TestJsonUnmarshalOneOfTryEach(t *testing.T) {
	t.Parallel()

	t.Run("noDiscriminator: variant A selected by required field 'a'", func(t *testing.T) {
		t.Parallel()

		var v testOneOfNoDisc.NoDiscriminator

		require.NoError(t, json.Unmarshal([]byte(`{"value":{"a":"hello"}}`), &v))
		require.NotNil(t, v.Value.Variant0)
		assert.Nil(t, v.Value.Variant1)
		assert.Equal(t, "hello", v.Value.Variant0.A)
	})

	t.Run("noDiscriminator: variant B selected by required field 'b'", func(t *testing.T) {
		t.Parallel()

		var v testOneOfNoDisc.NoDiscriminator

		require.NoError(t, json.Unmarshal([]byte(`{"value":{"b":"world"}}`), &v))
		require.NotNil(t, v.Value.Variant1)
		assert.Nil(t, v.Value.Variant0)
		assert.Equal(t, "world", v.Value.Variant1.B)
	})

	t.Run("noDiscriminator: input matching no variant rejected", func(t *testing.T) {
		t.Parallel()

		var v testOneOfNoDisc.NoDiscriminator

		err := json.Unmarshal([]byte(`{"value":{"c":"orphan"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no oneOf variant matched")
	})

	t.Run("noDiscriminator: ambiguous input matching both variants rejected", func(t *testing.T) {
		t.Parallel()

		// Both `a` and `b` present → both variants unmarshal successfully
		// (each variant requires only one field, the other is treated as
		// extra and allowed since AdditionalProperties default is true).
		var v testOneOfNoDisc.NoDiscriminator

		err := json.Unmarshal([]byte(`{"value":{"a":"x","b":"y"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ambiguous")
	})

	t.Run("noDiscriminator: round-trip preserves variant", func(t *testing.T) {
		t.Parallel()

		input := []byte(`{"value":{"a":"persisted"}}`)

		var v testOneOfNoDisc.NoDiscriminator

		require.NoError(t, json.Unmarshal(input, &v))

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, string(input), string(out))
	})

	t.Run("strictShape: shape check eliminates variant B when 'age' present", func(t *testing.T) {
		t.Parallel()

		// `age` is in variant A's properties only; with additionalProperties:false
		// on both variants, variant B's shape check fails.
		var v testOneOfStrictShape.StrictShape

		require.NoError(t, json.Unmarshal(
			[]byte(`{"value":{"name":"Alice","age":30}}`),
			&v,
		))
		require.NotNil(t, v.Value.Variant0)
		assert.Nil(t, v.Value.Variant1)
		assert.Equal(t, "Alice", v.Value.Variant0.Name)
	})

	t.Run("strictShape: shape check eliminates variant A when 'label' present", func(t *testing.T) {
		t.Parallel()

		var v testOneOfStrictShape.StrictShape

		require.NoError(t, json.Unmarshal(
			[]byte(`{"value":{"name":"Bob","label":"admin"}}`),
			&v,
		))
		require.NotNil(t, v.Value.Variant1)
		assert.Nil(t, v.Value.Variant0)
		require.NotNil(t, v.Value.Variant1.Label)
		assert.Equal(t, "admin", *v.Value.Variant1.Label)
	})

	t.Run("strictShape: input matching only common 'name' is ambiguous", func(t *testing.T) {
		t.Parallel()

		// Without a discriminator and without distinguishing keys, both
		// strict variants accept {"name":"x"} → ambiguous.
		var v testOneOfStrictShape.StrictShape

		err := json.Unmarshal([]byte(`{"value":{"name":"x"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ambiguous")
	})
}

// TestJsonUnmarshalRecursiveAllOf covers the recursive-allOf path that
// previously emitted `interface{}` because generateType silently fell
// through determineTypeName when the AllOf variants had mismatched
// Type slices ($ref with empty Type vs inline {type:"object"}). The
// fix delegates AllOf/AnyOf in generateType the same way
// generateTypeInline already does, so a recursive type round-trips
// through a properly-typed struct instead of `interface{}`.
func TestJsonUnmarshalRecursiveAllOf(t *testing.T) {
	t.Parallel()

	t.Run("flat tree node decodes both inherited and inline fields", func(t *testing.T) {
		t.Parallel()

		var v testRecursiveAllOfSelfRef.SelfRef

		require.NoError(t, json.Unmarshal(
			[]byte(`{"root":{"name":"root","value":"v0"}}`),
			&v,
		))
		require.NotNil(t, v.Root.Name)
		assert.Equal(t, "root", *v.Root.Name)
		assert.Equal(t, "v0", v.Root.Value)
		assert.Empty(t, v.Root.Children)
	})

	t.Run("nested children decode recursively", func(t *testing.T) {
		t.Parallel()

		var v testRecursiveAllOfSelfRef.SelfRef

		require.NoError(t, json.Unmarshal([]byte(`{
			"root": {
				"name": "root",
				"value": "v0",
				"children": [
					{"name": "a", "value": "va"},
					{"name": "b", "value": "vb", "children": [
						{"name": "b1", "value": "vb1"}
					]}
				]
			}
		}`), &v))

		require.Len(t, v.Root.Children, 2)
		assert.Equal(t, "va", v.Root.Children[0].Value)
		require.Len(t, v.Root.Children[1].Children, 1)
		assert.Equal(t, "vb1", v.Root.Children[1].Children[0].Value)
	})

	t.Run("missing required value field is rejected", func(t *testing.T) {
		t.Parallel()

		var v testRecursiveAllOfSelfRef.SelfRef

		err := json.Unmarshal([]byte(`{"root":{"name":"x"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value")
	})

	t.Run("round-trip preserves nested structure", func(t *testing.T) {
		t.Parallel()

		input := []byte(`{"root":{"name":"r","value":"v","children":[{"value":"c1"}]}}`)

		var v testRecursiveAllOfSelfRef.SelfRef

		require.NoError(t, json.Unmarshal(input, &v))

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, string(input), string(out))
	})
}

func TestJsonUnmarshalRootComposition(t *testing.T) {
	t.Parallel()

	// rootOneOfDiscriminator: schema is `{"oneOf":[...]}` at root with no
	// surrounding `{"type":"object"}` wrapper. Pre-fix this was silently
	// dropped; now the discriminator detection runs at root and emits a
	// holder with per-variant pointer fields.

	t.Run("rootOneOfDiscriminator: ping variant decodes into Ping field", func(t *testing.T) {
		t.Parallel()

		var v testRootOneOfDisc.RootOneOfDiscriminator

		require.NoError(t, json.Unmarshal([]byte(`{"kind":"ping","ttl":30}`), &v))
		require.NotNil(t, v.Ping)
		assert.Nil(t, v.Pong)
		assert.Equal(t, "ping", v.Ping.Kind)
		assert.Equal(t, 30, v.Ping.Ttl)
	})

	t.Run("rootOneOfDiscriminator: pong variant decodes into Pong field", func(t *testing.T) {
		t.Parallel()

		var v testRootOneOfDisc.RootOneOfDiscriminator

		require.NoError(t, json.Unmarshal([]byte(`{"kind":"pong","echo":"hi"}`), &v))
		require.NotNil(t, v.Pong)
		assert.Nil(t, v.Ping)
		assert.Equal(t, "hi", v.Pong.Echo)
	})

	t.Run("rootOneOfDiscriminator: round-trip preserves variant", func(t *testing.T) {
		t.Parallel()

		input := []byte(`{"echo":"hello","kind":"pong"}`)

		var v testRootOneOfDisc.RootOneOfDiscriminator

		require.NoError(t, json.Unmarshal(input, &v))

		out, err := json.Marshal(&v)
		require.NoError(t, err)
		assert.JSONEq(t, string(input), string(out))
	})

	// rootOneOfTryEach: object oneOf at root with no natural discriminator
	// — disambiguation happens via per-variant shape check.

	t.Run("rootOneOfTryEach: variant with 'age' key selected", func(t *testing.T) {
		t.Parallel()

		var v testRootOneOfTryEach.RootOneOfTryEach

		require.NoError(t, json.Unmarshal([]byte(`{"name":"x","age":7}`), &v))
		require.NotNil(t, v.Variant0)
		assert.Nil(t, v.Variant1)
	})

	t.Run("rootOneOfTryEach: variant with 'label' key selected", func(t *testing.T) {
		t.Parallel()

		var v testRootOneOfTryEach.RootOneOfTryEach

		require.NoError(t, json.Unmarshal([]byte(`{"name":"x","label":"l"}`), &v))
		require.NotNil(t, v.Variant1)
		assert.Nil(t, v.Variant0)
	})

	t.Run("rootOneOfTryEach: input matching no variant rejected", func(t *testing.T) {
		t.Parallel()

		var v testRootOneOfTryEach.RootOneOfTryEach

		err := json.Unmarshal([]byte(`{"name":"x","other":1}`), &v)
		assert.Error(t, err)
	})

	// rootAllOf: schema is `{"allOf":[{"$ref":Base},{inline}]}` at root.
	// Pre-fix this was silently dropped. Now the merged struct gains
	// fields from both branches.

	t.Run("rootAllOf: merged struct decodes both Base and inline fields", func(t *testing.T) {
		t.Parallel()

		var v testRootAllOf.RootAllOf

		require.NoError(t, json.Unmarshal([]byte(`{"name":"alice","extra":"data"}`), &v))
		assert.Equal(t, "alice", v.Name)
		assert.Equal(t, "data", v.Extra)
	})

	t.Run("rootAllOf: missing required field rejected", func(t *testing.T) {
		t.Parallel()

		var v testRootAllOf.RootAllOf

		err := json.Unmarshal([]byte(`{"name":"alice"}`), &v)
		assert.Error(t, err)
	})

	// rootEnum: schema is `{"enum":[...]}` at root. Pre-fix this was
	// silently dropped; now generates a typed string with named constants.

	t.Run("rootEnum: declared value accepted", func(t *testing.T) {
		t.Parallel()

		var v testRootEnum.RootEnum

		require.NoError(t, json.Unmarshal([]byte(`"red"`), &v))
		assert.Equal(t, testRootEnum.RootEnumRed, v)
	})

	t.Run("rootEnum: undeclared value rejected", func(t *testing.T) {
		t.Parallel()

		var v testRootEnum.RootEnum

		err := json.Unmarshal([]byte(`"purple"`), &v)
		assert.Error(t, err)
	})
}

// TestJsonUnmarshalConditionalDiscriminator exercises the runtime behavior
// of the allOf+if[const]/then[/else] tagged-union dispatch: per-variant
// required-field enforcement, unknown-discriminator rejection, JSON
// round-trip, and the missing/wrong-type discriminator error paths.
func TestJsonUnmarshalConditionalDiscriminator(t *testing.T) {
	t.Parallel()

	// basic: tagged-union with per-discriminator required fields. Generated
	// as a single struct (all fields optional except the discriminator);
	// runtime checks enforce the per-variant required set.

	t.Run("basic: create variant requires payload", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscBasic.Basic

		err := json.Unmarshal([]byte(`{"kind":"create","payload":"x"}`), &v)
		require.NoError(t, err)
		assert.Equal(t, testCondDiscBasic.BasicKindCreate, v.Kind)
		require.NotNil(t, v.Payload)
		assert.Equal(t, "x", *v.Payload)
	})

	t.Run("basic: create without payload rejected", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscBasic.Basic

		err := json.Unmarshal([]byte(`{"kind":"create"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payload")
		assert.Contains(t, err.Error(), "kind='create'")
	})

	t.Run("basic: update requires id, payload, and version", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscBasic.Basic

		err := json.Unmarshal([]byte(`{"kind":"update","id":"1","payload":"x","version":2}`), &v)
		require.NoError(t, err)
	})

	t.Run("basic: update missing version rejected", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscBasic.Basic

		err := json.Unmarshal([]byte(`{"kind":"update","id":"1","payload":"x"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version")
	})

	t.Run("basic: delete requires id only", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscBasic.Basic

		err := json.Unmarshal([]byte(`{"kind":"delete","id":"1"}`), &v)
		require.NoError(t, err)
	})

	t.Run("basic: unknown discriminator rejected by enum check", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscBasic.Basic

		err := json.Unmarshal([]byte(`{"kind":"unknown"}`), &v)
		require.Error(t, err)
	})

	// withElse: discriminator with an else clause — different required set
	// when the discriminator does NOT match the const.

	t.Run("withElse: primary requires weight", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscWithElse.WithElse

		err := json.Unmarshal([]byte(`{"category":"primary","weight":5}`), &v)
		require.NoError(t, err)
	})

	t.Run("withElse: primary missing weight rejected", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscWithElse.WithElse

		err := json.Unmarshal([]byte(`{"category":"primary","label":"x"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "weight")
	})

	t.Run("withElse: secondary requires label (else branch)", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscWithElse.WithElse

		err := json.Unmarshal([]byte(`{"category":"secondary","label":"x"}`), &v)
		require.NoError(t, err)
	})

	t.Run("withElse: secondary missing label rejected (else branch)", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscWithElse.WithElse

		err := json.Unmarshal([]byte(`{"category":"secondary","weight":5}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "label")
	})
}

// TestJsonUnmarshalConditionalDiscriminatorEnumForm exercises the runtime
// behavior of the enum-form discriminator dispatch: per-variant required
// enforcement when discStr is in the branch's value set, runtime
// interpolation of the offending value into the error message, and the
// const-vs-enum precedence + overlapping-branch semantics.
func TestJsonUnmarshalConditionalDiscriminatorEnumForm(t *testing.T) {
	t.Parallel()

	// enumIf: 4-value enum branch requires `value`; the parent enum has a
	// fifth value `eq` that does NOT trigger the branch.
	t.Run("enumIf: gt with value succeeds", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumIf.EnumIf
		require.NoError(t, json.Unmarshal([]byte(`{"operator":"gt","value":1.5}`), &v))
	})

	t.Run("enumIf: gt without value rejected with operator='gt' in error", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumIf.EnumIf

		err := json.Unmarshal([]byte(`{"operator":"gt"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value")
		assert.Contains(t, err.Error(), "operator='gt'")
	})

	t.Run("enumIf: lte without value rejected with operator='lte' in error", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumIf.EnumIf

		err := json.Unmarshal([]byte(`{"operator":"lte"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operator='lte'")
	})

	t.Run("enumIf: eq is in parent enum but NOT in branch set, no value required", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumIf.EnumIf
		require.NoError(t, json.Unmarshal([]byte(`{"operator":"eq"}`), &v))
	})

	// enumIfWithElse: 3-value enum then-branch requires billingId; else
	// branch fires for the fourth value (free) and requires freeReason.
	t.Run("enumIfWithElse: premium requires billingId", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumElse.EnumIfWithElse
		require.NoError(t, json.Unmarshal([]byte(`{"tier":"premium","billingId":"b1"}`), &v))
	})

	t.Run("enumIfWithElse: silver missing billingId rejected", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumElse.EnumIfWithElse

		err := json.Unmarshal([]byte(`{"tier":"silver"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "billingId")
		assert.Contains(t, err.Error(), "tier='silver'")
	})

	t.Run("enumIfWithElse: free triggers else, requires freeReason", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscEnumElse.EnumIfWithElse
		require.NoError(t, json.Unmarshal([]byte(`{"tier":"free","freeReason":"trial"}`), &v))

		err := json.Unmarshal([]byte(`{"tier":"free"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "freeReason")
		assert.Contains(t, err.Error(), "tier!='free'")
	})

	t.Run("enumIfWithElse: premium does NOT trigger else", func(t *testing.T) {
		t.Parallel()

		// premium with billingId but no freeReason should succeed.
		var v testCondDiscEnumElse.EnumIfWithElse
		require.NoError(t, json.Unmarshal([]byte(`{"tier":"premium","billingId":"b1"}`), &v))
	})

	// enumIfMixedWithConst: const branch + enum branch on the same key.
	// create requires payload (const form), update/patch require id+payload
	// (enum form with runtime discStr in error).
	t.Run("enumIfMixedWithConst: const branch error uses static format", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscMixed.EnumIfMixedWithConst

		err := json.Unmarshal([]byte(`{"kind":"create"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payload")
		assert.Contains(t, err.Error(), "kind='create'")
	})

	t.Run("enumIfMixedWithConst: enum branch error interpolates discStr", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscMixed.EnumIfMixedWithConst

		err := json.Unmarshal([]byte(`{"kind":"patch","id":"x"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payload")
		// Specifically `kind='patch'` not `kind='update'` — proves the runtime
		// interpolation works and didn't pin to the first value in the set.
		assert.Contains(t, err.Error(), "kind='patch'")
	})

	t.Run("enumIfMixedWithConst: delete fires no branch, accepts minimal input", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscMixed.EnumIfMixedWithConst
		require.NoError(t, json.Unmarshal([]byte(`{"kind":"delete"}`), &v))
	})

	// enumIfOverlapping: const "dog" + enum ["dog","cat"]. dog activates
	// BOTH branches; cat only the second.
	t.Run("enumIfOverlapping: dog requires both bark and age", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscOverlap.EnumIfOverlapping
		require.NoError(t, json.Unmarshal([]byte(`{"species":"dog","bark":"woof","age":3}`), &v))

		err := json.Unmarshal([]byte(`{"species":"dog","bark":"woof"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "age")

		err = json.Unmarshal([]byte(`{"species":"dog","age":3}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bark")
	})

	t.Run("enumIfOverlapping: cat requires age only (const-dog branch silent)", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscOverlap.EnumIfOverlapping
		require.NoError(t, json.Unmarshal([]byte(`{"species":"cat","age":7}`), &v))

		err := json.Unmarshal([]byte(`{"species":"cat"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "age")
		assert.Contains(t, err.Error(), "species='cat'")
	})

	// enumIfConstAndEnum: const "a" + enum ["a","b"] on the same property.
	// Const wins (more specific) — only kind == "a" should trigger the
	// required check, not "b".
	t.Run("enumIfConstAndEnum: const precedence — only 'a' triggers", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscConstAndEnum.EnumIfConstAndEnum

		err := json.Unmarshal([]byte(`{"kind":"a"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "details")

		// `b` is in the parent enum and the if-clause enum, but const-precedence
		// means only the const value `a` activates the branch.
		require.NoError(t, json.Unmarshal([]byte(`{"kind":"b"}`), &v))
	})

	// enumIfManyValues: 20-value enum, all require hex.
	t.Run("enumIfManyValues: arbitrary value triggers requirement", func(t *testing.T) {
		t.Parallel()

		var v testCondDiscManyValues.EnumIfManyValues
		require.NoError(t, json.Unmarshal([]byte(`{"color":"magenta","hex":"#f0f"}`), &v))

		err := json.Unmarshal([]byte(`{"color":"navy"}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hex")
		assert.Contains(t, err.Error(), "color='navy'")
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
