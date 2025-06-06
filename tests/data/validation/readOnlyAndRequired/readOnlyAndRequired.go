// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import yaml "gopkg.in/yaml.v3"

type ReadOnlyAndRequired struct {
	// MyReadOnlyRequiredString corresponds to the JSON schema field
	// "myReadOnlyRequiredString".
	MyReadOnlyRequiredString string `json:"myReadOnlyRequiredString" yaml:"myReadOnlyRequiredString" mapstructure:"myReadOnlyRequiredString"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ReadOnlyAndRequired) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["myReadOnlyRequiredString"]; raw != nil && !ok {
		return fmt.Errorf("field myReadOnlyRequiredString in ReadOnlyAndRequired: required")
	}
	if _, ok := raw["myReadOnlyRequiredString"]; raw != nil && ok {
		return fmt.Errorf("field myReadOnlyRequiredString in ReadOnlyAndRequired: read only")
	}
	type Plain ReadOnlyAndRequired
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = ReadOnlyAndRequired(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *ReadOnlyAndRequired) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["myReadOnlyRequiredString"]; raw != nil && !ok {
		return fmt.Errorf("field myReadOnlyRequiredString in ReadOnlyAndRequired: required")
	}
	if _, ok := raw["myReadOnlyRequiredString"]; raw != nil && ok {
		return fmt.Errorf("field myReadOnlyRequiredString in ReadOnlyAndRequired: read only")
	}
	type Plain ReadOnlyAndRequired
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = ReadOnlyAndRequired(plain)
	return nil
}
