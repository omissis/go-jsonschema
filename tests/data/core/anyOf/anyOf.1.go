// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "errors"
import "fmt"
import yaml "gopkg.in/yaml.v3"

type AnyOf1 struct {
	// Configurations corresponds to the JSON schema field "configurations".
	Configurations []AnyOf1ConfigurationsElem `json:"configurations,omitempty" yaml:"configurations,omitempty" mapstructure:"configurations,omitempty"`

	// Flags corresponds to the JSON schema field "flags".
	Flags interface{} `json:"flags,omitempty" yaml:"flags,omitempty" mapstructure:"flags,omitempty"`
}

type AnyOf1ConfigurationsElem struct {
	// Bar corresponds to the JSON schema field "bar".
	Bar float64 `json:"bar" yaml:"bar" mapstructure:"bar"`

	// Baz corresponds to the JSON schema field "baz".
	Baz *bool `json:"baz,omitempty" yaml:"baz,omitempty" mapstructure:"baz,omitempty"`

	// Foo corresponds to the JSON schema field "foo".
	Foo string `json:"foo" yaml:"foo" mapstructure:"foo"`
}

type AnyOf1ConfigurationsElem_0 struct {
	// Foo corresponds to the JSON schema field "foo".
	Foo string `json:"foo" yaml:"foo" mapstructure:"foo"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AnyOf1ConfigurationsElem_0) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["foo"]; raw != nil && !ok {
		return fmt.Errorf("field foo in AnyOf1ConfigurationsElem_0: required")
	}
	type Plain AnyOf1ConfigurationsElem_0
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem_0(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AnyOf1ConfigurationsElem_0) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["foo"]; raw != nil && !ok {
		return fmt.Errorf("field foo in AnyOf1ConfigurationsElem_0: required")
	}
	type Plain AnyOf1ConfigurationsElem_0
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem_0(plain)
	return nil
}

type AnyOf1ConfigurationsElem_1 struct {
	// Bar corresponds to the JSON schema field "bar".
	Bar float64 `json:"bar" yaml:"bar" mapstructure:"bar"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AnyOf1ConfigurationsElem_1) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["bar"]; raw != nil && !ok {
		return fmt.Errorf("field bar in AnyOf1ConfigurationsElem_1: required")
	}
	type Plain AnyOf1ConfigurationsElem_1
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem_1(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AnyOf1ConfigurationsElem_1) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["bar"]; raw != nil && !ok {
		return fmt.Errorf("field bar in AnyOf1ConfigurationsElem_1: required")
	}
	type Plain AnyOf1ConfigurationsElem_1
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem_1(plain)
	return nil
}

type AnyOf1ConfigurationsElem_2 struct {
	// Baz corresponds to the JSON schema field "baz".
	Baz *bool `json:"baz,omitempty" yaml:"baz,omitempty" mapstructure:"baz,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AnyOf1ConfigurationsElem_2) UnmarshalJSON(value []byte) error {
	type Plain AnyOf1ConfigurationsElem_2
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem_2(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AnyOf1ConfigurationsElem_2) UnmarshalYAML(value *yaml.Node) error {
	type Plain AnyOf1ConfigurationsElem_2
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem_2(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AnyOf1ConfigurationsElem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	var anyOf1ConfigurationsElem_0 AnyOf1ConfigurationsElem_0
	var anyOf1ConfigurationsElem_1 AnyOf1ConfigurationsElem_1
	var anyOf1ConfigurationsElem_2 AnyOf1ConfigurationsElem_2
	var errs []error
	if err := anyOf1ConfigurationsElem_0.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOf1ConfigurationsElem_1.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOf1ConfigurationsElem_2.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain AnyOf1ConfigurationsElem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AnyOf1ConfigurationsElem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	var anyOf1ConfigurationsElem_0 AnyOf1ConfigurationsElem_0
	var anyOf1ConfigurationsElem_1 AnyOf1ConfigurationsElem_1
	var anyOf1ConfigurationsElem_2 AnyOf1ConfigurationsElem_2
	var errs []error
	if err := anyOf1ConfigurationsElem_0.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOf1ConfigurationsElem_1.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOf1ConfigurationsElem_2.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain AnyOf1ConfigurationsElem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AnyOf1ConfigurationsElem(plain)
	return nil
}