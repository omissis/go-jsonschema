package tests_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	testAdditionalProperties "github.com/atombender/go-jsonschema/tests/data/core/additionalProperties"
	testAllOf "github.com/atombender/go-jsonschema/tests/data/core/allOf"
	testAnyOf "github.com/atombender/go-jsonschema/tests/data/core/anyOf"
	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
)

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
							{Foo: "hello"},
							{Bar: 2.2},
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
							{Foo: "ciao"},
							{Bar: 200.0},
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
			target: &testAnyOf.AnyOf1{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAnyOf.AnyOf1{
						Configurations: []testAnyOf.AnyOf1ConfigurationsElem{
							{Foo: "ciao"},
							{Bar: 2.0},
							{Baz: ptr(false)},
						},
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
			desc: "allOf - 1",
			json: `{
				"configurations": [
					{
						"foo": "hello",
						"bar": 2.2
					}
				]
			}`,
			target: &testAllOf.AllOf{},
			assertFn: func(target any) {
				assert.Equal(
					t,
					&testAllOf.AllOf{
						Configurations: []testAllOf.AllOfConfigurationsElem{
							{Foo: "hello", Bar: 2.2},
						},
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
