// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import yaml "gopkg.in/yaml.v3"

type AllOf3 struct {
	// Bar corresponds to the JSON schema field "bar".
	Bar float64 `json:"bar" yaml:"bar" mapstructure:"bar"`

	// Configurations corresponds to the JSON schema field "configurations".
	Configurations []interface{} `json:"configurations,omitempty" yaml:"configurations,omitempty" mapstructure:"configurations,omitempty"`

	// Foo corresponds to the JSON schema field "foo".
	Foo string `json:"foo" yaml:"foo" mapstructure:"foo"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AllOf3) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["bar"]; raw != nil && !ok {
		return fmt.Errorf("field bar in AllOf3: required")
	}
	if _, ok := raw["foo"]; raw != nil && !ok {
		return fmt.Errorf("field foo in AllOf3: required")
	}
	type Plain AllOf3
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AllOf3(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AllOf3) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["bar"]; raw != nil && !ok {
		return fmt.Errorf("field bar in AllOf3: required")
	}
	if _, ok := raw["foo"]; raw != nil && !ok {
		return fmt.Errorf("field foo in AllOf3: required")
	}
	type Plain AllOf3
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AllOf3(plain)
	return nil
}