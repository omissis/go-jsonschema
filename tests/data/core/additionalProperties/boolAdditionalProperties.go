// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "github.com/go-viper/mapstructure/v2"
import "reflect"
import "strings"

type BoolAdditionalProperties struct {
	// Name corresponds to the JSON schema field "name".
	Name *string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`

	AdditionalProperties map[string]bool `mapstructure:",remain"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *BoolAdditionalProperties) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	type Plain BoolAdditionalProperties
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if v, ok := raw[""]; !ok || v == nil {
		plain.AdditionalProperties = map[string]bool{}
	}
	st := reflect.TypeOf(Plain{})
	for i := range st.NumField() {
		delete(raw, st.Field(i).Name)
		delete(raw, strings.Split(st.Field(i).Tag.Get("json"), ",")[0])
	}
	if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {
		return err
	}
	*j = BoolAdditionalProperties(plain)
	return nil
}
