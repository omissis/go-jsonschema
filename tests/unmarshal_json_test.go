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
	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
	testNullTypes "github.com/atombender/go-jsonschema/tests/data/validateNullTypes"
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

// TestJsonUnmarshalValidateNullTypes pins the opt-in `type` enforcement for
// explicit nulls, and — more importantly — pins the cases that must keep
// accepting null.
//
// Per draft-07 a null is invalid only because `type` says so: §6.5.3 defines
// `required` as testing presence by name (so `{"x": null}` satisfies it), and
// keywords are vacuously true for instances of a type they do not target. A
// blanket "reject null" would therefore be a spec violation.
func TestJsonUnmarshalValidateNullTypes(t *testing.T) {
	t.Parallel()

	t.Run("null is rejected where type excludes it", func(t *testing.T) {
		t.Parallel()

		var v testNullTypes.NullTypes

		// The container itself: `type: object` does not admit null.
		require.Error(t, json.Unmarshal([]byte(`null`), &v))

		// A present-but-null property whose type excludes null.
		err := json.Unmarshal([]byte(`{"name":null}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be null")

		err = json.Unmarshal([]byte(`{"name":"a","tags":null}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be null")
	})

	t.Run("omission and valid values are unaffected", func(t *testing.T) {
		t.Parallel()

		var v testNullTypes.NullTypes

		require.NoError(t, json.Unmarshal([]byte(`{"name":"a"}`), &v))

		// Still a required-field error, not a null error.
		err := json.Unmarshal([]byte(`{}`), &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("null stays valid where the schema admits it", func(t *testing.T) {
		t.Parallel()

		var v testNullTypes.NullAllowed

		// `{"type": ["string", "null"]}` admits null deliberately.
		require.NoError(t, json.Unmarshal([]byte(`{"nullable":null}`), &v))

		// No `type` keyword constrains nothing, so null is valid.
		require.NoError(t, json.Unmarshal([]byte(`{"untyped":null}`), &v))

		// But a plain `type: string` still rejects it.
		require.Error(t, json.Unmarshal([]byte(`{"strict":null}`), &v))
	})
}
