package tests_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"

	testAdditionalProperties "github.com/atombender/go-jsonschema/tests/data/core/additionalProperties"
	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
	"github.com/stretchr/testify/assert"
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
