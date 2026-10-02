package tests_test

import (
	"net/netip"
	"os"
	"reflect"
	"strings"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"

	test "github.com/atombender/go-jsonschema/tests/data/extraImports/gopkgYAMLv3"
	testIpDefault "github.com/atombender/go-jsonschema/tests/data/validation/ipDefault"
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

func TestYamlV3UnmarshalIPDefault(t *testing.T) {
	t.Parallel()

	yamlData := []byte(`{}`)

	var conf testIpDefault.IpDefault

	if err := yamlv3.Unmarshal(yamlData, &conf); err != nil {
		t.Fatal(err)
	}

	want := testIpDefault.IpDefault{
		Ipv4:         netip.MustParseAddr("127.0.0.1"),
		Ipv6:         netip.MustParseAddr("2001:db8::68"),
		NullableIpv4: ptr(netip.MustParseAddr("192.168.1.1")),
	}

	if !reflect.DeepEqual(conf, want) {
		t.Errorf("Unmarshalled YAML data does not match expected\nWant: %#v\nGot:  %#v", want, conf)
	}
}
