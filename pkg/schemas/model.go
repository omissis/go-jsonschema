// Package schemas defines JSON schema types.
//
// Code borrowed from https://github.com/alecthomas/jsonschema/
//
// # Copyright (C) 2014 Alec Thomas
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
	"fmt"
	"reflect"
	"slices"
	"strings"

	"dario.cat/mergo"
)

var (
	ErrCannotMergeTypes = fmt.Errorf("cannot merge types")
	ErrEmptyTypesList   = fmt.Errorf("types list is empty")
)

// Schema is the root schema.
type Schema struct {
	*ObjectAsType

	ID          string      `json:"$id"` // RFC draft-wright-json-schema-01, section-9.2.
	LegacyID    string      `json:"id"`  // RFC draft-wright-json-schema-00, section 4.5.
	Definitions Definitions `json:"$defs,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler for Schema struct.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var unmarshSchema unmarshalerSchema
	if err := json.Unmarshal(data, &unmarshSchema); err != nil {
		return fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	// Fall back to id if $id is not present.
	if unmarshSchema.ID == "" {
		unmarshSchema.ID = unmarshSchema.LegacyID
	}

	// Take care of legacy fields.
	var legacySchema struct {
		Definitions Definitions `json:"definitions,omitempty"`
	}

	if err := json.Unmarshal(data, &legacySchema); err != nil {
		return fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	if unmarshSchema.Definitions == nil && legacySchema.Definitions != nil {
		unmarshSchema.Definitions = legacySchema.Definitions
	}

	*s = Schema(unmarshSchema)

	return nil
}

type (
	unmarshalerSchema Schema
	ObjectAsType      Type
)

// TypeList is a list of type names.
type TypeList []string

// UnmarshalJSON implements json.Unmarshaler.
func (t *TypeList) UnmarshalJSON(value []byte) error {
	if len(value) > 0 && value[0] == '[' {
		var s []string
		if err := json.Unmarshal(value, &s); err != nil {
			return fmt.Errorf("failed to unmarshal type list: %w", err)
		}

		*t = s

		return nil
	}

	var s string
	if err := json.Unmarshal(value, &s); err != nil {
		return fmt.Errorf("failed to unmarshal type list: %w", err)
	}

	if s != "" {
		*t = []string{s}
	} else {
		*t = nil
	}

	return nil
}

func (t *TypeList) Equals(b TypeList) bool {
	if t == nil {
		return false
	}

	if len(*t) != len(b) {
		return false
	}

	for i := range *t {
		if (*t)[i] != b[i] {
			return false
		}
	}

	return true
}

func (t *TypeList) Add(s string) {
	if !slices.Contains(*t, s) {
		*t = append(*t, s)
	}
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26.
type Definitions map[string]*Type

type SubSchemaType string

const (
	SubSchemaTypeAllOf SubSchemaType = "allOf"
	SubSchemaTypeAnyOf SubSchemaType = "anyOf"
	SubSchemaTypeOneOf SubSchemaType = "oneOf"
	SubSchemaTypeNot   SubSchemaType = "not"
)

// Type represents a JSON Schema object type.
type Type struct {
	// RFC draft-wright-json-schema-00.
	Version string `json:"$schema,omitempty"` // Section 6.1.
	Ref     string `json:"$ref,omitempty"`    // Section 7.
	// RFC draft-wright-json-schema-validation-00, section 5.
	MultipleOf           *float64         `json:"multipleOf,omitempty"`           // Section 5.1.
	Maximum              *float64         `json:"maximum,omitempty"`              // Section 5.2.
	ExclusiveMaximum     *any             `json:"exclusiveMaximum,omitempty"`     // Section 5.3. Changed in draft 4.
	Minimum              *float64         `json:"minimum,omitempty"`              // Section 5.4.
	ExclusiveMinimum     *any             `json:"exclusiveMinimum,omitempty"`     // Section 5.5. Changed in draft 4.
	MaxLength            int              `json:"maxLength,omitempty"`            // Section 5.6.
	MinLength            int              `json:"minLength,omitempty"`            // Section 5.7.
	Pattern              string           `json:"pattern,omitempty"`              // Section 5.8.
	AdditionalItems      *Type            `json:"additionalItems,omitempty"`      // Section 5.9.
	Items                *Type            `json:"items,omitempty"`                // Section 5.9.
	MaxItems             int              `json:"maxItems,omitempty"`             // Section 5.10.
	MinItems             int              `json:"minItems,omitempty"`             // Section 5.11.
	UniqueItems          bool             `json:"uniqueItems,omitempty"`          // Section 5.12.
	MaxProperties        int              `json:"maxProperties,omitempty"`        // Section 5.13.
	MinProperties        int              `json:"minProperties,omitempty"`        // Section 5.14.
	Required             []string         `json:"required,omitempty"`             // Section 5.15.
	Properties           map[string]*Type `json:"properties,omitempty"`           // Section 5.16.
	PatternProperties    map[string]*Type `json:"patternProperties,omitempty"`    // Section 5.17.
	AdditionalProperties *Type            `json:"additionalProperties,omitempty"` // Section 5.18.
	Enum                 []any            `json:"enum,omitempty"`                 // Section 5.20.
	Type                 TypeList         `json:"type,omitempty"`                 // Section 5.21.
	// RFC draft-bhutton-json-schema-01, section 10.
	AllOf []*Type `json:"allOf,omitempty"` // Section 10.2.1.1.
	AnyOf []*Type `json:"anyOf,omitempty"` // Section 10.2.1.2.
	OneOf []*Type `json:"oneOf,omitempty"` // Section 10.2.1.3.
	Not   *Type   `json:"not,omitempty"`   // Section 10.2.1.4.
	// RFC draft-wright-json-schema-validation-00, section 6, 7.
	Title       string `json:"title,omitempty"`       // Section 6.1.
	Description string `json:"description,omitempty"` // Section 6.1.
	Default     any    `json:"default,omitempty"`     // Section 6.2.
	Format      string `json:"format,omitempty"`      // Section 7.
	// RFC draft-wright-json-schema-hyperschema-00, section 4.
	Media          *Type  `json:"media,omitempty"`          // Section 4.3.
	BinaryEncoding string `json:"binaryEncoding,omitempty"` // Section 4.3.
	// RFC draft-handrews-json-schema-validation-02, section 6.
	DependentRequired map[string][]string `json:"dependentRequired,omitempty"` // Section 6.5.4.
	// RFC draft-handrews-json-schema-validation-02, appendix A.
	Definitions      Definitions      `json:"$defs,omitempty"`
	DependentSchemas map[string]*Type `json:"dependentSchemas,omitempty"`

	// TODO: add correct section where "readOnly" is mentioned in the spec
	//       I'm not sure which section I should put here, but I did notice in the 2020-12 validation schema changelog,
	//       under the "draft-handrews-json-schema-validation-00" item it mentions "readOnly" as having been moved
	//       from hyper-schema to validation meta-data...
	ReadOnly bool `json:"readOnly,omitempty"`

	// ExtGoCustomType is the name of a (qualified or not) custom Go type
	// to use for the field.
	GoJSONSchemaExtension *GoJSONSchemaExtension `json:"goJSONSchema,omitempty"` //nolint:tagliatelle // breaking change

	// SubSchemaType marks the type as being a subschema type.
	subSchemaType     SubSchemaType `json:"-"`
	subSchemasCount   int           `json:"-"`
	subSchemaTypeElem bool          `json:"-"`

	// Flags.
	Dereferenced bool `json:"-"` // Marks that his type has been dereferenced.
}

func (value *Type) SetSubSchemaType(sst SubSchemaType) {
	value.subSchemaType = sst
}

func (value *Type) GetSubSchemaType() SubSchemaType {
	return value.subSchemaType
}

func (value *Type) SetSubSchemasCount(ssc int) {
	value.subSchemasCount = ssc
}

func (value *Type) GetSubSchemasCount() int {
	return value.subSchemasCount
}

func (value *Type) IsSubSchemaTypeElem() bool {
	return value.subSchemaTypeElem
}

func (value *Type) SetSubSchemaTypeElem() {
	value.subSchemaTypeElem = true
}

func (value *Type) ConvertAllRefs(absolutePath string) error {
	val := reflect.ValueOf(value).Elem()

	return updateAllRefsValues(&val, absolutePath)
}

// UnmarshalJSON accepts booleans as schemas where `true` is equivalent to `{}`
// and `false` is equivalent to `{"not": {}}`.
func (value *Type) UnmarshalJSON(raw []byte) error {
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		if b {
			*value = Type{}
		} else {
			*value = Type{Not: &Type{}}
		}

		return nil
	}

	var obj ObjectAsType
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("failed to unmarshal type: %w", err)
	}

	// Take care of legacy fields from older RFC versions.
	legacyObj := struct {
		// RFC draft-wright-json-schema-validation-00, section 5.
		Dependencies map[string]*Type `json:"dependencies,omitempty"`
		Definitions  Definitions      `json:"definitions,omitempty"` // Section 5.26.
	}{}
	if err := json.Unmarshal(raw, &legacyObj); err != nil {
		return fmt.Errorf("failed to unmarshal type: %w", err)
	}

	if legacyObj.Definitions != nil && obj.Definitions == nil {
		obj.Definitions = legacyObj.Definitions
	}

	if legacyObj.Dependencies != nil && obj.DependentSchemas == nil {
		obj.DependentSchemas = legacyObj.Dependencies
	}

	*value = Type(obj)

	return nil
}

func AllOf(types []*Type, baseType *Type) (*Type, error) {
	typ, err := MergeTypes(types, baseType)
	if err != nil {
		return nil, err
	}

	typ.subSchemaType = SubSchemaTypeAllOf

	return typ, nil
}

func AnyOf(types []*Type, baseType *Type) (*Type, error) {
	typ, err := MergeTypes(types, baseType)
	if err != nil {
		return nil, err
	}

	typ.Required = mergeRequiredUnion(types, baseType)
	typ.subSchemaType = SubSchemaTypeAnyOf
	typ.subSchemasCount = len(types)

	return typ, nil
}

func MergeTypes(types []*Type, baseType *Type) (*Type, error) {
	if len(types) == 0 {
		return nil, ErrEmptyTypesList
	}

	result := &Type{}

	if isPrimitiveTypeList(types, result.Type) {
		return result, nil
	}

	opts := []func(*mergo.Config){
		mergo.WithAppendSlice,
		mergo.WithTransformers(typeListTransformer{}),
	}

	if err := mergo.Merge(result, baseType, opts...); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCannotMergeTypes, err)
	}

	for _, t := range types {
		if err := mergo.Merge(result, t, opts...); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCannotMergeTypes, err)
		}
	}

	return result, nil
}

func mergeRequiredUnion(types []*Type, baseType *Type) []string {
	required := make([]string, len(baseType.Required))
	copy(required, baseType.Required)

	for _, r := range types[0].Required {
		valid := true

		for _, t := range types {
			if !slices.Contains(t.Required, r) {
				valid = false

				break
			}
		}

		if valid && !slices.Contains(required, r) {
			required = append(required, r) //nolint:makezero
		}
	}

	return required
}

func updateAllRefsValues(structValue *reflect.Value, refPath string) error {
	switch structValue.Kind() { //nolint:exhaustive
	case reflect.Struct:
		for i := range structValue.NumField() {
			field := structValue.Field(i)
			name := structValue.Type().Field(i).Name

			switch field.Kind() { //nolint:exhaustive
			case reflect.String:
				fieldVal := field.String()
				if name == "Ref" && fieldVal != "" && field.CanSet() {
					if strings.HasPrefix(fieldVal, "#") {
						field.SetString(refPath + fieldVal)
					}
				}

			default:
				if err := updateAllRefsValues(&field, refPath); err != nil {
					return fmt.Errorf("struct error: %w", err)
				}
			}
		}

	case reflect.Ptr:
		elem := structValue.Elem()
		if !structValue.IsNil() {
			if err := updateAllRefsValues(&elem, refPath); err != nil {
				return fmt.Errorf("ptr error: %w", err)
			}
		}

	case reflect.Map:
		for _, key := range structValue.MapKeys() {
			val := structValue.MapIndex(key)
			if err := updateAllRefsValues(&val, refPath); err != nil {
				return fmt.Errorf("map error: %w", err)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := range structValue.Len() {
			field := structValue.Index(i)
			if err := updateAllRefsValues(&field, refPath); err != nil {
				return fmt.Errorf("slice error: %w", err)
			}
		}
	}

	return nil
}

type typeListTransformer struct{}

func (t typeListTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeFor[TypeList]() {
		return func(dst, src reflect.Value) error {
			return nil
		}
	}

	return nil
}

type GoJSONSchemaExtension struct {
	Type       *string           `json:"type,omitempty"`
	Identifier *string           `json:"identifier,omitempty"`
	Nillable   bool              `json:"nillable,omitempty"`
	Imports    []string          `json:"imports,omitempty"`
	ExtraTags  map[string]string `json:"extraTags,omitempty"`
}
