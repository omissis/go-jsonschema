// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import "github.com/go-viper/mapstructure/v2"
import yaml "gopkg.in/yaml.v3"
import "reflect"
import "strings"

type Alpha struct {
	// Beta corresponds to the JSON schema field "beta".
	Beta AlphaBeta `json:"beta,omitempty" yaml:"beta,omitempty" mapstructure:"beta,omitempty"`

	// Eta corresponds to the JSON schema field "eta".
	Eta *Eta `json:"eta,omitempty" yaml:"eta,omitempty" mapstructure:"eta,omitempty"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

type AlphaBeta interface{}

type Beta interface{}

type Eta struct {
	// Epsilon corresponds to the JSON schema field "epsilon".
	Epsilon string `json:"epsilon" yaml:"epsilon" mapstructure:"epsilon"`

	// Theta corresponds to the JSON schema field "theta".
	Theta Theta `json:"theta" yaml:"theta" mapstructure:"theta"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Eta) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["epsilon"]; raw != nil && !ok {
		return fmt.Errorf("field epsilon in Eta: required")
	}
	if _, ok := raw["theta"]; raw != nil && !ok {
		return fmt.Errorf("field theta in Eta: required")
	}
	type Plain Eta
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = Eta(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Eta) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["epsilon"]; raw != nil && !ok {
		return fmt.Errorf("field epsilon in Eta: required")
	}
	if _, ok := raw["theta"]; raw != nil && !ok {
		return fmt.Errorf("field theta in Eta: required")
	}
	type Plain Eta
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = Eta(plain)
	return nil
}

type Iota struct {
	// DESCRIPTION
	Kappa *TITLE `json:"kappa,omitempty" yaml:"kappa,omitempty" mapstructure:"kappa,omitempty"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

type IotaKappaLambdaElem struct {
	// Sigma corresponds to the JSON schema field "sigma".
	Sigma *Alpha `json:"sigma,omitempty" yaml:"sigma,omitempty" mapstructure:"sigma,omitempty"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

type Properties struct {
	// Iota corresponds to the JSON schema field "iota".
	Iota Iota `json:"iota" yaml:"iota" mapstructure:"iota"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Properties) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["iota"]; raw != nil && !ok {
		return fmt.Errorf("field iota in Properties: required")
	}
	type Plain Properties
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = Properties(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Properties) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["iota"]; raw != nil && !ok {
		return fmt.Errorf("field iota in Properties: required")
	}
	type Plain Properties
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = Properties(plain)
	return nil
}

// DESCRIPTION
type TITLE struct {
	// Lambda corresponds to the JSON schema field "lambda".
	Lambda []IotaKappaLambdaElem `json:"lambda,omitempty" yaml:"lambda,omitempty" mapstructure:"lambda,omitempty"`

	AdditionalProperties interface{} `mapstructure:",remain"`
}

type Theta int

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Theta) UnmarshalYAML(value *yaml.Node) error {
	type Plain Theta
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 65535 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 65535)
	}
	if 0 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", 0)
	}
	*j = Theta(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Theta) UnmarshalJSON(value []byte) error {
	type Plain Theta
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 65535 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 65535)
	}
	if 0 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", 0)
	}
	*j = Theta(plain)
	return nil
}
