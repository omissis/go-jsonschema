package tests_test

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	yamlv3 "gopkg.in/yaml.v3"

	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
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

	assert.Equal(t, test.GopkgYAMLv3{
		MyString:  &s,
		MyNumber:  &n,
		MyInteger: &i,
		MyBoolean: &b,
		MyNull:    nil,
	}, conf)
}
