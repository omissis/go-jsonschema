package tests_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testAdditionalProperties "github.com/tuotuoxp/go-jsonschema/tests/data/core/additionalProperties"
	testAllOf "github.com/tuotuoxp/go-jsonschema/tests/data/core/allOf"
	testAnyOf "github.com/tuotuoxp/go-jsonschema/tests/data/core/anyOf"
	testMultiOneOfEnvelope "github.com/tuotuoxp/go-jsonschema/tests/data/core/multiOneOfEnvelope"
	testOneOfEnvelope "github.com/tuotuoxp/go-jsonschema/tests/data/core/oneOfEnvelope"
	testOneOfEnvelopeRefEnumDiscriminator "github.com/tuotuoxp/go-jsonschema/tests/data/core/oneOfEnvelopeRefEnumDiscriminator"
	testOneOfEnvelopeRefEnumDiscriminatorOptionalType "github.com/tuotuoxp/go-jsonschema/tests/data/core/oneOfEnvelopeRefEnumDiscriminatorOptionalType"
	testRefSemanticInline "github.com/tuotuoxp/go-jsonschema/tests/data/core/refSemanticInline"
	testRefSemanticNamed "github.com/tuotuoxp/go-jsonschema/tests/data/core/refSemanticNamed"
	testRefToEnum "github.com/tuotuoxp/go-jsonschema/tests/data/core/refToEnum"
	testRefToPrimitiveString "github.com/tuotuoxp/go-jsonschema/tests/data/core/refToPrimitiveString"
	test "github.com/tuotuoxp/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
	testValudationRequiredFields "github.com/tuotuoxp/go-jsonschema/tests/data/validation/requiredFields"
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

func TestMultiOneOfEnvelopeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success/type_a_mode_y_nested_n1", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		require.NoError(t, json.Unmarshal([]byte(`{
			"dummy":"x",
			"type":"a",
			"value":{
				"sub_a":"hello",
				"nested_type":"n1",
				"nested_value":{"leaf_a":"leaf"}
			},
			"mode":"y",
			"payload":{"sub_y":true}
		}`), &v))

		assert.Equal(t, ptr("x"), v.Dummy)
		assert.Equal(t, testMultiOneOfEnvelope.MultiOneOfEnvelopeTypeA, v.Type)
		assert.Equal(t, testMultiOneOfEnvelope.MultiOneOfEnvelopeModeY, v.Mode)
		assert.NotNil(t, v.Value.A)
		assert.Nil(t, v.Value.B)
		assert.Nil(t, v.Value.C)
		assert.Equal(t, "hello", v.Value.A.SubA)
		assert.Equal(t, testMultiOneOfEnvelope.AValueNestedTypeN1, v.Value.A.NestedType)
		assert.NotNil(t, v.Value.A.NestedValue.N1)
		assert.Nil(t, v.Value.A.NestedValue.N2)
		assert.Equal(t, "leaf", v.Value.A.NestedValue.N1.LeafA)
		assert.Nil(t, v.Payload.X)
		assert.NotNil(t, v.Payload.Y)
		assert.Nil(t, v.Payload.Z)
		assert.True(t, v.Payload.Y.SubY)
	})

	t.Run("success/type_b_mode_x", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		require.NoError(t, json.Unmarshal([]byte(`{
			"dummy":"x",
			"type":"b",
			"value":{"sub_b":10},
			"mode":"x",
			"payload":{"sub_x":"xxx"}
		}`), &v))

		assert.Equal(t, ptr("x"), v.Dummy)
		assert.Equal(t, testMultiOneOfEnvelope.MultiOneOfEnvelopeTypeB, v.Type)
		assert.Equal(t, testMultiOneOfEnvelope.MultiOneOfEnvelopeModeX, v.Mode)
		assert.Nil(t, v.Value.A)
		assert.NotNil(t, v.Value.B)
		assert.Nil(t, v.Value.C)
		assert.Equal(t, 10, v.Value.B.SubB)
		assert.NotNil(t, v.Payload.X)
		assert.Nil(t, v.Payload.Y)
		assert.Nil(t, v.Payload.Z)
		assert.Equal(t, "xxx", v.Payload.X.SubX)
	})

	t.Run("success/type_c_mode_z", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		require.NoError(t, json.Unmarshal([]byte(`{
			"dummy":"x",
			"type":"c",
			"value":{"sub_c":true},
			"mode":"z",
			"payload":{"sub_z":11}
		}`), &v))

		assert.Equal(t, ptr("x"), v.Dummy)
		assert.Equal(t, testMultiOneOfEnvelope.MultiOneOfEnvelopeTypeC, v.Type)
		assert.Equal(t, testMultiOneOfEnvelope.MultiOneOfEnvelopeModeZ, v.Mode)
		assert.Nil(t, v.Value.A)
		assert.Nil(t, v.Value.B)
		assert.NotNil(t, v.Value.C)
		assert.True(t, v.Value.C.SubC)
		assert.Nil(t, v.Payload.X)
		assert.Nil(t, v.Payload.Y)
		assert.NotNil(t, v.Payload.Z)
		assert.Equal(t, 11, v.Payload.Z.SubZ)
	})

	t.Run("failure/type_a_pattern_violation", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		err := json.Unmarshal([]byte(`{
			"type":"a",
			"value":{"sub_a":"HELLO","nested_type":"n1","nested_value":{"leaf_a":"leaf"}},
			"mode":"y",
			"payload":{"sub_y":true}
		}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern")
	})

	t.Run("failure/mode_z_minimum_violation", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		err := json.Unmarshal([]byte(`{
			"type":"c",
			"value":{"sub_c":true},
			"mode":"z",
			"payload":{"sub_z":9}
		}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be >=")
	})

	t.Run("failure/nested_type_n2_minimum_violation", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		err := json.Unmarshal([]byte(`{
			"type":"a",
			"value":{"sub_a":"hello","nested_type":"n2","nested_value":{"leaf_b":1}},
			"mode":"y",
			"payload":{"sub_y":true}
		}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be >=")
	})

	t.Run("failure/invalid_top_level_type", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		err := json.Unmarshal([]byte(`{
			"type":"invalid",
			"value":{"sub_c":true},
			"mode":"z",
			"payload":{"sub_z":11}
		}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"invalid"`)
	})

	t.Run("failure/invalid_top_level_mode", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		err := json.Unmarshal([]byte(`{
			"type":"c",
			"value":{"sub_c":true},
			"mode":"invalid",
			"payload":{"sub_z":11}
		}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"invalid"`)
	})

	t.Run("failure/type_a_invalid_nested_type", func(t *testing.T) {
		t.Parallel()

		var v testMultiOneOfEnvelope.MultiOneOfEnvelope
		err := json.Unmarshal([]byte(`{
			"type":"a",
			"value":{"sub_a":"hello","nested_type":"invalid","nested_value":{"leaf_a":"leaf"}},
			"mode":"y",
			"payload":{"sub_y":true}
		}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"invalid"`)
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

	t.Run("success/missing_type_selects_no_branch", func(t *testing.T) {
		t.Parallel()

		var v testOneOfEnvelopeRefEnumDiscriminatorOptionalType.OneOfEnvelopeRefEnumDiscriminatorOptionalType
		require.NoError(t, json.Unmarshal([]byte(`{"value":{"sub_a":"hello"}}`), &v))
		assert.Nil(t, v.Type)
		assert.Nil(t, v.Value.A)
		assert.Nil(t, v.Value.B)
	})
}

func TestRefSemanticInlineUnmarshalJSON(t *testing.T) {
	t.Parallel()

	basePayload := map[string]any{
		"inlineName":        "abc.example",
		"refName":           "def.example",
		"inlineCode":        "abcd",
		"refCode":           "bcde",
		"inlineStringConst": "stable",
		"refStringConst":    "stable",
		"inlineInteger":     5,
		"refInteger":        6,
		"inlineNumber":      3.5,
		"refNumber":         3.5,
		"inlineFlag":        true,
		"refFlag":           true,
		"inlineDateTime":    "2024-04-20T10:30:00Z",
		"refDateTime":       "2024-04-20T10:30:00Z",
	}

	cloneBase := func() map[string]any {
		out := make(map[string]any, len(basePayload))
		for key, val := range basePayload {
			out[key] = val
		}

		return out
	}

	toJSON := func(t *testing.T, m map[string]any) []byte {
		t.Helper()

		value, err := json.Marshal(m)
		require.NoError(t, err)

		return value
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var value testRefSemanticInline.RefSemanticInline
		require.NoError(t, json.Unmarshal(toJSON(t, cloneBase()), &value))
	})

	t.Run("failure/inline_pattern", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["inlineName"] = "INVALID"

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern match")
	})

	t.Run("failure/ref_pattern", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refName"] = "INVALID"

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern match")
	})

	t.Run("failure/ref_string_length", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refCode"] = "ab"

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be >=")
	})

	t.Run("failure/ref_string_const", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refStringConst"] = "unstable"

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be equal")
	})

	t.Run("failure/ref_integer_range", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refInteger"] = 10

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be <")
	})

	t.Run("failure/ref_number_const", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refNumber"] = 4.5

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be equal")
	})

	t.Run("failure/ref_boolean_const", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refFlag"] = false

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be equal to true")
	})

	t.Run("failure/inline_date_time_format", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["inlineDateTime"] = "bad"

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot parse")
	})

	t.Run("failure/ref_date_time_format", func(t *testing.T) {
		t.Parallel()

		payload := cloneBase()
		payload["refDateTime"] = "bad"

		var value testRefSemanticInline.RefSemanticInline
		err := json.Unmarshal(toJSON(t, payload), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot parse")
	})
}

func TestRefToPrimitiveStringUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var value testRefToPrimitiveString.RefToPrimitiveString
		require.NoError(t, json.Unmarshal([]byte(`{"inlineThing":"abc.example","refThing":"def.example"}`), &value))
	})

	t.Run("failure/inline_pattern", func(t *testing.T) {
		t.Parallel()

		var value testRefToPrimitiveString.RefToPrimitiveString
		err := json.Unmarshal([]byte(`{"inlineThing":"INVALID","refThing":"def.example"}`), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern match")
	})

	t.Run("failure/ref_pattern", func(t *testing.T) {
		t.Parallel()

		var value testRefToPrimitiveString.RefToPrimitiveString
		err := json.Unmarshal([]byte(`{"inlineThing":"abc.example","refThing":"INVALID"}`), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern match")
	})
}

func TestRefToEnumUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var value testRefToEnum.RefToEnum
		require.NoError(t, json.Unmarshal([]byte(`{"inlineThing":"x","refThing":"y"}`), &value))
	})

	t.Run("failure/inline_enum", func(t *testing.T) {
		t.Parallel()

		var value testRefToEnum.RefToEnum
		err := json.Unmarshal([]byte(`{"inlineThing":"z","refThing":"y"}`), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"z"`)
	})

	t.Run("failure/ref_enum", func(t *testing.T) {
		t.Parallel()

		var value testRefToEnum.RefToEnum
		err := json.Unmarshal([]byte(`{"inlineThing":"x","refThing":"z"}`), &value)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"z"`)
	})
}

func TestRefSemanticNamedUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var value testRefSemanticNamed.RefSemanticNamed
		require.NoError(
			t,
			json.Unmarshal(
				[]byte(`{"titleName":"abc.example","goJSONSchemaTypeName":"abc","xGoTypeName":"AB"}`),
				&value,
			),
		)
	})

	t.Run("failure/title_ref_pattern", func(t *testing.T) {
		t.Parallel()

		var value testRefSemanticNamed.RefSemanticNamed
		err := json.Unmarshal(
			[]byte(`{"titleName":"INVALID","goJSONSchemaTypeName":"abc","xGoTypeName":"AB"}`),
			&value,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern match")
	})

	t.Run("failure/x_go_type_ref_pattern", func(t *testing.T) {
		t.Parallel()

		var value testRefSemanticNamed.RefSemanticNamed
		err := json.Unmarshal(
			[]byte(`{"titleName":"abc.example","goJSONSchemaTypeName":"abc","xGoTypeName":"ab"}`),
			&value,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pattern match")
	})

}
