package tests_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"

	testStringAddl "github.com/atombender/go-jsonschema/tests/data/core/additionalProperties"
	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
)

// TestYamlAdditionalPropertiesIsCaseSensitive pins the pruning rule used when
// filling AdditionalProperties for YAML.
//
// yaml.v3 binds keys to fields case-SENSITIVELY, unlike encoding/json. So given
// a declared `name`, the key `Name` is left unbound by the decoder and is a
// genuine additional property. The generated UnmarshalYAML must therefore prune
// `raw` by exact match; pruning case-insensitively (as the JSON path correctly
// does) would silently swallow `Name` instead of surfacing it.
func TestYamlAdditionalPropertiesIsCaseSensitive(t *testing.T) {
	t.Parallel()

	var v testStringAddl.StringAdditionalProperties

	if err := yamlv3.Unmarshal([]byte("name: bound\nName: extra\n"), &v); err != nil {
		t.Fatal(err)
	}

	if v.Name == nil || *v.Name != "bound" {
		t.Fatalf("declared field `name` should bind exactly; got %v", v.Name)
	}

	got, ok := v.AdditionalProperties["Name"]
	if !ok {
		t.Fatalf("case-variant key `Name` should survive as an additional property, got %#v", v.AdditionalProperties)
	}

	if got != "extra" {
		t.Fatalf("AdditionalProperties[\"Name\"] = %q, want %q", got, "extra")
	}

	if _, leaked := v.AdditionalProperties["name"]; leaked {
		t.Fatalf("declared key `name` should have been pruned, got %#v", v.AdditionalProperties)
	}
}

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
