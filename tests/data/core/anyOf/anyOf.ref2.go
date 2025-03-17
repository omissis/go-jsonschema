// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package test

import "encoding/json"
import "errors"
import "fmt"
import yaml "gopkg.in/yaml.v3"

type AnyOfRef2 struct {
	// Qux2 corresponds to the JSON schema field "qux2".
	Qux2 []AnyOfRef2Qux2Elem `json:"qux2,omitempty" yaml:"qux2,omitempty" mapstructure:"qux2,omitempty"`
}

type AnyOfRef2Qux2Elem struct {
	// Content corresponds to the JSON schema field "content".
	Content []interface{} `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *AnyOfRef2Qux2Elem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	var anyOfRef2Qux2Elem_0 AnyOfRef2Qux2Elem_0
	var anyOfRef2Qux2Elem_1 AnyOfRef2Qux2Elem_1
	var anyOfRef2Qux2Elem_2 AnyOfRef2Qux2Elem_2
	var errs []error
	if err := anyOfRef2Qux2Elem_0.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOfRef2Qux2Elem_1.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOfRef2Qux2Elem_2.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain AnyOfRef2Qux2Elem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = AnyOfRef2Qux2Elem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *AnyOfRef2Qux2Elem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	var anyOfRef2Qux2Elem_0 AnyOfRef2Qux2Elem_0
	var anyOfRef2Qux2Elem_1 AnyOfRef2Qux2Elem_1
	var anyOfRef2Qux2Elem_2 AnyOfRef2Qux2Elem_2
	var errs []error
	if err := anyOfRef2Qux2Elem_0.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOfRef2Qux2Elem_1.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := anyOfRef2Qux2Elem_2.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain AnyOfRef2Qux2Elem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = AnyOfRef2Qux2Elem(plain)
	return nil
}

type Bar2 struct {
	// Content corresponds to the JSON schema field "content".
	Content []Bar2ContentElem `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

type Bar2ContentElem struct {
	// Content corresponds to the JSON schema field "content".
	Content []interface{} `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Bar2ContentElem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	var bar2ContentElem_0 Bar2ContentElem_0
	var bar2ContentElem_1 Bar2ContentElem_1
	var bar2ContentElem_2 Bar2ContentElem_2
	var errs []error
	if err := bar2ContentElem_0.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := bar2ContentElem_1.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := bar2ContentElem_2.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain Bar2ContentElem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Bar2ContentElem(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Bar2ContentElem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	var bar2ContentElem_0 Bar2ContentElem_0
	var bar2ContentElem_1 Bar2ContentElem_1
	var bar2ContentElem_2 Bar2ContentElem_2
	var errs []error
	if err := bar2ContentElem_0.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := bar2ContentElem_1.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := bar2ContentElem_2.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain Bar2ContentElem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Bar2ContentElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Bar2) UnmarshalJSON(value []byte) error {
	type Plain Bar2
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Bar2(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Bar2) UnmarshalYAML(value *yaml.Node) error {
	type Plain Bar2
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Bar2(plain)
	return nil
}

type Baz2 struct {
	// Content corresponds to the JSON schema field "content".
	Content []Baz2ContentElem `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

type Baz2ContentElem struct {
	// Content corresponds to the JSON schema field "content".
	Content []interface{} `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Baz2ContentElem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	var baz2ContentElem_0 Baz2ContentElem_0
	var baz2ContentElem_1 Baz2ContentElem_1
	var baz2ContentElem_2 Baz2ContentElem_2
	var errs []error
	if err := baz2ContentElem_0.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := baz2ContentElem_1.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := baz2ContentElem_2.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain Baz2ContentElem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Baz2ContentElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Baz2ContentElem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	var baz2ContentElem_0 Baz2ContentElem_0
	var baz2ContentElem_1 Baz2ContentElem_1
	var baz2ContentElem_2 Baz2ContentElem_2
	var errs []error
	if err := baz2ContentElem_0.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := baz2ContentElem_1.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := baz2ContentElem_2.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain Baz2ContentElem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Baz2ContentElem(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Baz2) UnmarshalYAML(value *yaml.Node) error {
	type Plain Baz2
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Baz2(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Baz2) UnmarshalJSON(value []byte) error {
	type Plain Baz2
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Baz2(plain)
	return nil
}

type Foo2 struct {
	// Content corresponds to the JSON schema field "content".
	Content []Foo2ContentElem `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

type Foo2ContentElem struct {
	// Content corresponds to the JSON schema field "content".
	Content []interface{} `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`
}

type Foo2ContentElem_0 = Foo2
type Bar2ContentElem_1 = Bar2
type Bar2ContentElem_2 = Baz2

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Foo2ContentElem) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	var foo2ContentElem_0 Foo2ContentElem_0
	var foo2ContentElem_1 Foo2ContentElem_1
	var foo2ContentElem_2 Foo2ContentElem_2
	var errs []error
	if err := foo2ContentElem_0.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := foo2ContentElem_1.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if err := foo2ContentElem_2.UnmarshalYAML(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain Foo2ContentElem
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Foo2ContentElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Foo2ContentElem) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}
	var foo2ContentElem_0 Foo2ContentElem_0
	var foo2ContentElem_1 Foo2ContentElem_1
	var foo2ContentElem_2 Foo2ContentElem_2
	var errs []error
	if err := foo2ContentElem_0.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := foo2ContentElem_1.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if err := foo2ContentElem_2.UnmarshalJSON(value); err != nil {
		errs = append(errs, err)
	}
	if len(errs) == 3 {
		return fmt.Errorf("all validators failed: %s", errors.Join(errs...))
	}
	type Plain Foo2ContentElem
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Foo2ContentElem(plain)
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (j *Foo2) UnmarshalYAML(value *yaml.Node) error {
	type Plain Foo2
	var plain Plain
	if err := value.Decode(&plain); err != nil {
		return err
	}
	*j = Foo2(plain)
	return nil
}

type Foo2ContentElem_2 = Baz2
type Foo2ContentElem_1 = Bar2
type Bar2ContentElem_0 = Foo2
type AnyOfRef2Qux2Elem_0 = Foo2
type AnyOfRef2Qux2Elem_1 = Bar2
type AnyOfRef2Qux2Elem_2 = Baz2
type Baz2ContentElem_0 = Foo2

// UnmarshalJSON implements json.Unmarshaler.
func (j *Foo2) UnmarshalJSON(value []byte) error {
	type Plain Foo2
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}
	*j = Foo2(plain)
	return nil
}

type Baz2ContentElem_2 = Baz2
type Baz2ContentElem_1 = Bar2
