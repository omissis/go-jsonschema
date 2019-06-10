// Package schema defines JSON schema types.
//
// Code borrowed from https://github.com/alecthomas/jsonschema/
//
// Copyright (C) 2014 Alec Thomas
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package schemas

import (
	"encoding/json"
)

// Schema is the root schema.
// RFC draft-wright-json-schema-00, section 4.5
type Schema struct {
	*Type
	ID          string      `json:"id"`
	Definitions Definitions `json:"definitions,omitempty"`
}

// TypeList is a list of type names.
type TypeList []string

// UnmarshalJSON implements json.Unmarshaler.
func (t *TypeList) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '[' {
		var s []string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*t = TypeList(s)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != "" {
		*t = TypeList([]string{s})
	} else {
		*t = nil
	}
	return nil
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26
type Definitions map[string]*Type

// Type represents a JSON Schema object type.
type Type struct {
	// RFC draft-wright-json-schema-00
	Version string `json:"$schema,omitempty"` // section 6.1
	Ref     string `json:"$ref,omitempty"`    // section 7
	// RFC draft-wright-json-schema-validation-00, section 5
	MultipleOf           int              `json:"multipleOf,omitempty"`           // section 5.1
	Maximum              int              `json:"maximum,omitempty"`              // section 5.2
	ExclusiveMaximum     bool             `json:"exclusiveMaximum,omitempty"`     // section 5.3
	Minimum              int              `json:"minimum,omitempty"`              // section 5.4
	ExclusiveMinimum     bool             `json:"exclusiveMinimum,omitempty"`     // section 5.5
	MaxLength            int              `json:"maxLength,omitempty"`            // section 5.6
	MinLength            int              `json:"minLength,omitempty"`            // section 5.7
	Pattern              string           `json:"pattern,omitempty"`              // section 5.8
	AdditionalItems      *Type            `json:"additionalItems,omitempty"`      // section 5.9
	Items                *Type            `json:"items,omitempty"`                // section 5.9
	MaxItems             int              `json:"maxItems,omitempty"`             // section 5.10
	MinItems             int              `json:"minItems,omitempty"`             // section 5.11
	UniqueItems          bool             `json:"uniqueItems,omitempty"`          // section 5.12
	MaxProperties        int              `json:"maxProperties,omitempty"`        // section 5.13
	MinProperties        int              `json:"minProperties,omitempty"`        // section 5.14
	Required             []string         `json:"required,omitempty"`             // section 5.15
	Properties           map[string]*Type `json:"properties,omitempty"`           // section 5.16
	PatternProperties    map[string]*Type `json:"patternProperties,omitempty"`    // section 5.17
	AdditionalProperties json.RawMessage  `json:"additionalProperties,omitempty"` // section 5.18
	Dependencies         map[string]*Type `json:"dependencies,omitempty"`         // section 5.19
	Enum                 []interface{}    `json:"enum,omitempty"`                 // section 5.20
	Type                 TypeList         `json:"type,omitempty"`                 // section 5.21
	AllOf                []*Type          `json:"allOf,omitempty"`                // section 5.22
	AnyOf                []*Type          `json:"anyOf,omitempty"`                // section 5.23
	OneOf                []*Type          `json:"oneOf,omitempty"`                // section 5.24
	Not                  *Type            `json:"not,omitempty"`                  // section 5.25
	Definitions          Definitions      `json:"definitions,omitempty"`          // section 5.26
	// RFC draft-wright-json-schema-validation-00, section 6, 7
	Title       string      `json:"title,omitempty"`       // section 6.1
	Description string      `json:"description,omitempty"` // section 6.1
	Default     interface{} `json:"default,omitempty"`     // section 6.2
	Format      string      `json:"format,omitempty"`      // section 7
	// RFC draft-wright-json-schema-hyperschema-00, section 4
	Media          *Type  `json:"media,omitempty"`          // section 4.3
	BinaryEncoding string `json:"binaryEncoding,omitempty"` // section 4.3

	// ExtGoCustomType is the name of a (qualified or not) custom Go type
	// to use for the field.
	GoJSONSchemaExtension *GoJSONSchemaExtension `json:"goJSONSchema,omitempty"`
}

type GoJSONSchemaExtension struct {
	Type       *string  `json:"type,omitempty"`
	Identifier *string  `json:"identifier,omitempty"`
	Imports    []string `json:"imports,omitempty"`
}
