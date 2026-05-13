package tests_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testAdditionalProperties "github.com/atombender/go-jsonschema/tests/data/core/additionalProperties"
	testAllOf "github.com/atombender/go-jsonschema/tests/data/core/allOf"
	testAnyOf "github.com/atombender/go-jsonschema/tests/data/core/anyOf"
	testOneOfEnvelope "github.com/atombender/go-jsonschema/tests/data/core/oneOfEnvelope"
	testOneOfEnvelopeRefEnumDiscriminator "github.com/atombender/go-jsonschema/tests/data/core/oneOfEnvelopeRefEnumDiscriminator"
	testOneOfEnvelopeRefEnumDiscriminatorOptionalType "github.com/atombender/go-jsonschema/tests/data/core/oneOfEnvelopeRefEnumDiscriminatorOptionalType"
	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
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

func formatGopkgYAMLv3(v test.GopkgYAMLv3) string {
	return fmt.Sprintf(
		"GopkgYAMLv3{MyString: %s, MyNumber: %f, MyInteger: %d, MyBoolean: %t, MyNull: %v, MyEnum: %v}",
		*v.MyString, *v.MyNumber, *v.MyInteger, *v.MyBoolean, nil, *v.MyEnum,
	)
}

func ptr[T any](v T) *T {
	return &v
}

// TestOneOfEnvelopeUnmarshalJSON verifies that the discriminator-based
// routing produced by x-go-oneof-envelope works correctly, and that the
// sub-type validators (pattern, minimum, required) are still executed.
func TestOneOfEnvelopeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success/type_a", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelope.OneOfEnvelope
		require.NoError(t, json.Unmarshal([]byte(`{"dummy":"x","type":"a","value":{"sub_a":"hello"}}`), &v))

		assert.Equal(t, ptr("x"), v.Dummy)
		assert.Equal(t, testOneOfEnvelope.OneOfEnvelopeTypeA, v.Type)
		assert.NotNil(t, v.Value.A)
		assert.Equal(t, "hello", v.Value.A.SubA)
		assert.Nil(t, v.Value.B)
		assert.Nil(t, v.Value.C)
	})

	t.Run("success/type_b", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelope.OneOfEnvelope
		require.NoError(t, json.Unmarshal([]byte(`{"dummy":"x","type":"b","value":{"sub_b":10}}`), &v))

		assert.Equal(t, ptr("x"), v.Dummy)
		assert.Equal(t, testOneOfEnvelope.OneOfEnvelopeTypeB, v.Type)
		assert.Nil(t, v.Value.A)
		assert.NotNil(t, v.Value.B)
		assert.Equal(t, 10, v.Value.B.SubB)
		assert.Nil(t, v.Value.C)
	})

	t.Run("success/type_c", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelope.OneOfEnvelope
		require.NoError(t, json.Unmarshal([]byte(`{"dummy":"x","type":"c","value":{"sub_c":true}}`), &v))

		assert.Equal(t, ptr("x"), v.Dummy)
		assert.Equal(t, testOneOfEnvelope.OneOfEnvelopeTypeC, v.Type)
		assert.Nil(t, v.Value.A)
		assert.Nil(t, v.Value.B)
		assert.NotNil(t, v.Value.C)
		assert.True(t, v.Value.C.SubC)
	})

	t.Run("failure/invalid_type_enum_value", func(t *testing.T) {
		t.Parallel()

		// "x" is not a valid enum value for the type field; the enum validator
		// fires before the envelope routing switch, so the error is about
		// the invalid enum value rather than an unknown discriminator.
		var v testOneOfEnvelope.OneOfEnvelope
		err := json.Unmarshal([]byte(`{"dummy":"x","type":"x","value":{}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"x"`)
	})

	t.Run("failure/type_a_pattern_violation", func(t *testing.T) {
		t.Parallel()

		// AValue.sub_a must match ^[a-z]+$ — uppercase fails.
		var v testOneOfEnvelope.OneOfEnvelope
		err := json.Unmarshal([]byte(`{"dummy":"x","type":"a","value":{"sub_a":"HELLO"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern")
	})

	t.Run("failure/type_b_minimum_violation", func(t *testing.T) {
		t.Parallel()

		// BValue.sub_b must be >= 1 — zero fails.
		var v testOneOfEnvelope.OneOfEnvelope
		err := json.Unmarshal([]byte(`{"dummy":"x","type":"b","value":{"sub_b":0}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be >=")
	})

	t.Run("failure/type_c_wrong_payload_shape", func(t *testing.T) {
		t.Parallel()

		// sub_c is required by CValue — supplying sub_b instead fails.
		var v testOneOfEnvelope.OneOfEnvelope
		err := json.Unmarshal([]byte(`{"dummy":"x","type":"c","value":{"sub_b":1}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sub_c")
	})
}

func TestOneOfEnvelopeRefEnumDiscriminatorUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success/type_a", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminator.OneOfEnvelopeRefEnumDiscriminator
		require.NoError(t, json.Unmarshal([]byte(`{"type":"a","value":{"sub_a":"hello"}}`), &v))

		assert.Equal(t, testOneOfEnvelopeRefEnumDiscriminator.EnvelopeTypeA, v.Type)
		assert.NotNil(t, v.Value.A)
		assert.Equal(t, "hello", v.Value.A.SubA)
		assert.Nil(t, v.Value.B)
	})

	t.Run("success/type_b", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminator.OneOfEnvelopeRefEnumDiscriminator
		require.NoError(t, json.Unmarshal([]byte(`{"type":"b","value":{"sub_b":10}}`), &v))

		assert.Equal(t, testOneOfEnvelopeRefEnumDiscriminator.EnvelopeTypeB, v.Type)
		assert.Nil(t, v.Value.A)
		assert.NotNil(t, v.Value.B)
		assert.Equal(t, 10, v.Value.B.SubB)
	})

	t.Run("failure/invalid_type_enum_value", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminator.OneOfEnvelopeRefEnumDiscriminator
		err := json.Unmarshal([]byte(`{"type":"x","value":{}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"x"`)
	})

	t.Run("failure/type_b_minimum_violation", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminator.OneOfEnvelopeRefEnumDiscriminator
		err := json.Unmarshal([]byte(`{"type":"b","value":{"sub_b":0}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be >=")
	})
}

func TestOneOfEnvelopeRefEnumDiscriminatorOptionalTypeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success/type_a", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminatorOptionalType.OneOfEnvelopeRefEnumDiscriminatorOptionalType
		require.NoError(t, json.Unmarshal([]byte(`{"type":"a","value":{"sub_a":"hello"}}`), &v))

		require.NotNil(t, v.Type)
		assert.Equal(t, testOneOfEnvelopeRefEnumDiscriminatorOptionalType.EnvelopeTypeA, *v.Type)
		assert.NotNil(t, v.Value.A)
		assert.Equal(t, "hello", v.Value.A.SubA)
		assert.Nil(t, v.Value.B)
	})

	t.Run("success/type_b", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminatorOptionalType.OneOfEnvelopeRefEnumDiscriminatorOptionalType
		require.NoError(t, json.Unmarshal([]byte(`{"type":"b","value":{"sub_b":10}}`), &v))

		require.NotNil(t, v.Type)
		assert.Equal(t, testOneOfEnvelopeRefEnumDiscriminatorOptionalType.EnvelopeTypeB, *v.Type)
		assert.Nil(t, v.Value.A)
		assert.NotNil(t, v.Value.B)
		assert.Equal(t, 10, v.Value.B.SubB)
	})

	t.Run("failure/invalid_type_enum_value", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminatorOptionalType.OneOfEnvelopeRefEnumDiscriminatorOptionalType
		err := json.Unmarshal([]byte(`{"type":"x","value":{}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"x"`)
	})

	t.Run("failure/missing_type_unknown_discriminator", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminatorOptionalType.OneOfEnvelopeRefEnumDiscriminatorOptionalType
		err := json.Unmarshal([]byte(`{"value":{"sub_a":"hello"}}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown discriminator value")
	})
}
