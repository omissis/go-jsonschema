package tests_test

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	yamlv3 "gopkg.in/yaml.v3"

	// test2 "github.com/atombender/go-jsonschema/tests/data/core/time"
	test1 "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
)

func TestYamlV3Unmarshal(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("./data/extraImports/gopkgYAMLv3/gopkgYAMLv3.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var conf test1.GopkgYAMLv3

	if err := yamlv3.Unmarshal(data, &conf); err != nil {
		t.Fatal(err)
	}

	s := "example"
	n := 123.456
	i := 123
	b := true
	e := test1.GopkgYAMLv3MyEnumX

	assert.Equal(t, test1.GopkgYAMLv3{
		MyString:  &s,
		MyNumber:  &n,
		MyInteger: &i,
		MyBoolean: &b,
		MyNull:    nil,
		MyEnum:    &e,
	}, conf)
}

func TestYamlV3UnmarshalInvalidEnum(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("./data/extraImports/gopkgYAMLv3invalidEnum/gopkgYAMLv3invalidEnum.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var conf test1.GopkgYAMLv3

	err = yamlv3.Unmarshal(data, &conf)
	if err == nil {
		t.Fatal("Expected unmarshal error")
	}

	assert.Matches(t, err.Error(), "invalid value \\(expected one of .*\\): .*")
}

// func TestTimeUnmarshal(t *testing.T) {
// 	t.Parallel()

// 	var conf test2.DatetimeRef

// 	if err := json.Unmarshal([]byte(`{"time":"2023-08-02T17:53:08.614Z"}`), &conf); err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, conf.Time, "2023-08-02 17:53:08.614 +0000 UTC")
// }
