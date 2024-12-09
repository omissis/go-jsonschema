// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import "regexp"

type Pattern struct {
	// MyNullableString corresponds to the JSON schema field "myNullableString".
	MyNullableString *string `json:"myNullableString,omitempty" yaml:"myNullableString,omitempty" mapstructure:"myNullableString,omitempty"`

	// MyString corresponds to the JSON schema field "myString".
	MyString string `json:"myString" yaml:"myString" mapstructure:"myString"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Pattern) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["myString"]; raw != nil && !ok {
		return fmt.Errorf("field myString in Pattern: required")
	}
	type Plain Pattern
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if plain.MyNullableString != nil {
		if matched, _ := regexp.MatchString(`^0x[0-9a-f]{10}$`, string(*plain.MyNullableString)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", `^0x[0-9a-f]{10}$`, "MyNullableString")
		}
	}
	if matched, _ := regexp.MatchString(`^0x[0-9a-f]{10}\.$`, string(plain.MyString)); !matched {
		return fmt.Errorf("field %s pattern match: must match %s", `^0x[0-9a-f]{10}\.$`, "MyString")
	}
	*j = Pattern(plain)
	return nil
}

// Verify checks all fields on the struct match the schema.
func (plain *Pattern) Verify() error {
	if plain.MyNullableString != nil {
		if matched, _ := regexp.MatchString(`^0x[0-9a-f]{10}$`, string(*plain.MyNullableString)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", `^0x[0-9a-f]{10}$`, "MyNullableString")
		}
	}
	if matched, _ := regexp.MatchString(`^0x[0-9a-f]{10}\.$`, string(plain.MyString)); !matched {
		return fmt.Errorf("field %s pattern match: must match %s", `^0x[0-9a-f]{10}\.$`, "MyString")
	}
	return nil
}
