// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "fmt"
import yaml "gopkg.in/yaml.v3"

type I16L int16

// UnmarshalJSON implements json.Unmarshaler.
func (j *I16L) UnmarshalJSON(value []byte) error {
	type Plain I16L
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 127 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 127)
	}
	if -129 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -129)
	}
	*j = I16L(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *I16L) UnmarshalYAML(value *yaml.Node) error {
	type Plain I16L
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 127 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 127)
	}
	if -129 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -129)
	}
	*j = I16L(plain)
	return nil
}

type I16U int16

// UnmarshalJSON implements json.Unmarshaler.
func (j *I16U) UnmarshalJSON(value []byte) error {
	type Plain I16U
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 128 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 128)
	}
	if -128 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -128)
	}
	*j = I16U(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *I16U) UnmarshalYAML(value *yaml.Node) error {
	type Plain I16U
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 128 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 128)
	}
	if -128 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -128)
	}
	*j = I16U(plain)
	return nil
}

type I32L int32

// UnmarshalJSON implements json.Unmarshaler.
func (j *I32L) UnmarshalJSON(value []byte) error {
	type Plain I32L
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 32767 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 32767)
	}
	if -32769 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -32769)
	}
	*j = I32L(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *I32L) UnmarshalYAML(value *yaml.Node) error {
	type Plain I32L
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 32767 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 32767)
	}
	if -32769 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -32769)
	}
	*j = I32L(plain)
	return nil
}

type I32U int32

// UnmarshalJSON implements json.Unmarshaler.
func (j *I32U) UnmarshalJSON(value []byte) error {
	type Plain I32U
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 32768 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 32768)
	}
	if -32768 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -32768)
	}
	*j = I32U(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *I32U) UnmarshalYAML(value *yaml.Node) error {
	type Plain I32U
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 32768 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 32768)
	}
	if -32768 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -32768)
	}
	*j = I32U(plain)
	return nil
}

type I64L int64

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *I64L) UnmarshalYAML(value *yaml.Node) error {
	type Plain I64L
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 2147483647 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 2147483647)
	}
	if -2147483649 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -2147483649)
	}
	*j = I64L(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *I64L) UnmarshalJSON(value []byte) error {
	type Plain I64L
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 2147483647 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 2147483647)
	}
	if -2147483649 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -2147483649)
	}
	*j = I64L(plain)
	return nil
}

type I64U int64

// UnmarshalJSON implements json.Unmarshaler.
func (j *I64U) UnmarshalJSON(value []byte) error {
	type Plain I64U
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 2147483648 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 2147483648)
	}
	if -2147483648 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -2147483648)
	}
	*j = I64U(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *I64U) UnmarshalYAML(value *yaml.Node) error {
	type Plain I64U
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 2147483648 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 2147483648)
	}
	if -2147483648 > plain {
		return fmt.Errorf("field %s: must be >= %v", "", -2147483648)
	}
	*j = I64U(plain)
	return nil
}

type Restricted struct {
	// I16L corresponds to the JSON schema field "i16l".
	I16L I16L `json:"i16l" yaml:"i16l" mapstructure:"i16l"`

	// I16U corresponds to the JSON schema field "i16u".
	I16U I16U `json:"i16u" yaml:"i16u" mapstructure:"i16u"`

	// I32L corresponds to the JSON schema field "i32l".
	I32L I32L `json:"i32l" yaml:"i32l" mapstructure:"i32l"`

	// I32U corresponds to the JSON schema field "i32u".
	I32U I32U `json:"i32u" yaml:"i32u" mapstructure:"i32u"`

	// I64L corresponds to the JSON schema field "i64l".
	I64L I64L `json:"i64l" yaml:"i64l" mapstructure:"i64l"`

	// I64U corresponds to the JSON schema field "i64u".
	I64U I64U `json:"i64u" yaml:"i64u" mapstructure:"i64u"`

	// U16 corresponds to the JSON schema field "u16".
	U16 U16 `json:"u16" yaml:"u16" mapstructure:"u16"`

	// U32 corresponds to the JSON schema field "u32".
	U32 U32 `json:"u32" yaml:"u32" mapstructure:"u32"`

	// U64 corresponds to the JSON schema field "u64".
	U64 U64 `json:"u64" yaml:"u64" mapstructure:"u64"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Restricted) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if _, ok := raw["i16l"]; raw != nil && !ok {
		return fmt.Errorf("field i16l in Restricted: required")
	}
	if _, ok := raw["i16u"]; raw != nil && !ok {
		return fmt.Errorf("field i16u in Restricted: required")
	}
	if _, ok := raw["i32l"]; raw != nil && !ok {
		return fmt.Errorf("field i32l in Restricted: required")
	}
	if _, ok := raw["i32u"]; raw != nil && !ok {
		return fmt.Errorf("field i32u in Restricted: required")
	}
	if _, ok := raw["i64l"]; raw != nil && !ok {
		return fmt.Errorf("field i64l in Restricted: required")
	}
	if _, ok := raw["i64u"]; raw != nil && !ok {
		return fmt.Errorf("field i64u in Restricted: required")
	}
	if _, ok := raw["u16"]; raw != nil && !ok {
		return fmt.Errorf("field u16 in Restricted: required")
	}
	if _, ok := raw["u32"]; raw != nil && !ok {
		return fmt.Errorf("field u32 in Restricted: required")
	}
	if _, ok := raw["u64"]; raw != nil && !ok {
		return fmt.Errorf("field u64 in Restricted: required")
	}
	type Plain Restricted
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Restricted(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Restricted) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	if _, ok := raw["i16l"]; raw != nil && !ok {
		return fmt.Errorf("field i16l in Restricted: required")
	}
	if _, ok := raw["i16u"]; raw != nil && !ok {
		return fmt.Errorf("field i16u in Restricted: required")
	}
	if _, ok := raw["i32l"]; raw != nil && !ok {
		return fmt.Errorf("field i32l in Restricted: required")
	}
	if _, ok := raw["i32u"]; raw != nil && !ok {
		return fmt.Errorf("field i32u in Restricted: required")
	}
	if _, ok := raw["i64l"]; raw != nil && !ok {
		return fmt.Errorf("field i64l in Restricted: required")
	}
	if _, ok := raw["i64u"]; raw != nil && !ok {
		return fmt.Errorf("field i64u in Restricted: required")
	}
	if _, ok := raw["u16"]; raw != nil && !ok {
		return fmt.Errorf("field u16 in Restricted: required")
	}
	if _, ok := raw["u32"]; raw != nil && !ok {
		return fmt.Errorf("field u32 in Restricted: required")
	}
	if _, ok := raw["u64"]; raw != nil && !ok {
		return fmt.Errorf("field u64 in Restricted: required")
	}
	type Plain Restricted
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Restricted(plain)
	return nil
}

type U16 uint16

// UnmarshalJSON implements json.Unmarshaler.
func (j *U16) UnmarshalJSON(value []byte) error {
	type Plain U16
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 256 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 256)
	}
	*j = U16(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *U16) UnmarshalYAML(value *yaml.Node) error {
	type Plain U16
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 256 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 256)
	}
	*j = U16(plain)
	return nil
}

type U32 uint32

// UnmarshalJSON implements json.Unmarshaler.
func (j *U32) UnmarshalJSON(value []byte) error {
	type Plain U32
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 65536 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 65536)
	}
	*j = U32(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *U32) UnmarshalYAML(value *yaml.Node) error {
	type Plain U32
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 65536 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 65536)
	}
	*j = U32(plain)
	return nil
}

type U64 uint64

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *U64) UnmarshalYAML(value *yaml.Node) error {
	type Plain U64
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	if 4294967296 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 4294967296)
	}
	*j = U64(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *U64) UnmarshalJSON(value []byte) error {
	type Plain U64
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	if 4294967296 < plain {
		return fmt.Errorf("field %s: must be <= %v", "", 4294967296)
	}
	*j = U64(plain)
	return nil
}
