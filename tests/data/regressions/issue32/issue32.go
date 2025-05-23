// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import yaml "gopkg.in/yaml.v3"

type TestObject struct {
	// Config corresponds to the JSON schema field "config".
	Config map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty" mapstructure:"config,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Owner corresponds to the JSON schema field "owner".
	Owner string `json:"owner" yaml:"owner" mapstructure:"owner"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *TestObject) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in TestObject: required")
	}
	if _, ok := raw["owner"]; raw != nil && !ok {
		return fmt.Errorf("field owner in TestObject: required")
	}
	type Plain TestObject
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = TestObject(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *TestObject) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["name"]; raw != nil && !ok {
		return fmt.Errorf("field name in TestObject: required")
	}
	if _, ok := raw["owner"]; raw != nil && !ok {
		return fmt.Errorf("field owner in TestObject: required")
	}
	type Plain TestObject
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = TestObject(plain)
	return nil
}
