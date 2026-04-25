package generator

import (
	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// oneOfKind is a bit-set of JSON kinds that the variants of a primitive
// `oneOf` cover. `integer` is intentionally not represented here: schemas
// that include an `integer` variant are routed away from the wrapper path
// because the wire-level JSON token kind cannot distinguish `1` from `1.5`,
// and silently widening `oneOf:[string,integer]` to accept `1.5` would
// violate the schema.
type oneOfKind uint8

const (
	oneOfKindString oneOfKind = 1 << iota
	oneOfKindNumber
	oneOfKindBoolean
	oneOfKindNull
)

func (k oneOfKind) has(other oneOfKind) bool { return k&other == other }

// isPrimitiveOneOf reports whether t is a `oneOf` whose variants are all
// simple JSON primitives suitable for the wrapper-type emission strategy.
// Variants must declare exactly one type from {string, number, boolean, null},
// must not contain nested compositions (oneOf/anyOf/allOf), and must not
// introduce constraints that require their own validators (format, minimum,
// maximum, minLength, maxLength, pattern, multipleOf, items, properties,
// patternProperties, additionalProperties). Schemas including an `integer`
// variant are also rejected because the wire-level JSON token cannot
// distinguish `1` from `1.5`, so the wrapper would silently widen
// `oneOf:[string,integer]` to accept `1.5`.
func isPrimitiveOneOf(t *schemas.Type) bool {
	if t == nil || len(t.OneOf) < 2 {
		return false
	}

	for _, variant := range t.OneOf {
		if variant == nil || len(variant.Type) != 1 {
			return false
		}

		if len(variant.OneOf)+len(variant.AnyOf)+len(variant.AllOf) > 0 {
			return false
		}

		if variant.Ref != "" || variant.Enum != nil || variant.Const != nil {
			return false
		}

		if variantHasValidationConstraints(variant) {
			return false
		}

		switch variant.Type[0] {
		case schemas.TypeNameString,
			schemas.TypeNameNumber,
			schemas.TypeNameBoolean,
			schemas.TypeNameNull:
		default:
			return false
		}
	}

	return true
}

// variantHasValidationConstraints reports whether a primitive `oneOf` variant
// declares any constraint that the wrapper-type emission strategy can't
// honor. The wrapper only dispatches on the JSON token kind, so a variant
// that needs e.g. `format`, `minimum`, or `pattern` checked would silently
// pass invalid values if routed through this path.
func variantHasValidationConstraints(v *schemas.Type) bool {
	if v.Format != "" || v.Pattern != "" {
		return true
	}

	if v.MinLength != 0 || v.MaxLength != 0 {
		return true
	}

	if v.Minimum != nil || v.Maximum != nil ||
		v.ExclusiveMinimum != nil || v.ExclusiveMaximum != nil ||
		v.MultipleOf != nil {
		return true
	}

	if len(v.Properties) > 0 || len(v.PatternProperties) > 0 ||
		v.AdditionalProperties != nil || len(v.Required) > 0 {
		return true
	}

	if v.Items != nil || v.MinItems != 0 || v.MaxItems != 0 || v.UniqueItems {
		return true
	}

	return false
}

// primitiveOneOfKinds returns the set of JSON kinds the variants of t cover.
// Caller must have already filtered out schemas containing `integer` via
// isPrimitiveOneOf — this function only handles the surviving primitive set.
func primitiveOneOfKinds(t *schemas.Type) oneOfKind {
	var kinds oneOfKind

	for _, variant := range t.OneOf {
		switch variant.Type[0] {
		case schemas.TypeNameString:
			kinds |= oneOfKindString
		case schemas.TypeNameNumber:
			kinds |= oneOfKindNumber
		case schemas.TypeNameBoolean:
			kinds |= oneOfKindBoolean
		case schemas.TypeNameNull:
			kinds |= oneOfKindNull
		}
	}

	return kinds
}

// generateOneOfPrimitive emits a wrapper type and the methods needed to
// round-trip a primitive `oneOf` schema. The generated type stores the
// decoded value in a single unexported `value any` field and exposes typed
// accessors plus IsZero/Value helpers; UnmarshalJSON dispatches on the
// JSON token kind so that overlapping schemas (e.g. number+integer) cannot
// match more than one variant by accident.
func (g *schemaGenerator) generateOneOfPrimitive(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	kinds := primitiveOneOfKinds(t)

	name := g.output.uniqueTypeName(scope)
	if g.config.StructNameFromTitle && t.Title != "" {
		name = g.caser.Identifierize(t.Title)
	}

	decl := &codegen.TypeDecl{
		Name:    name,
		Comment: t.Description,
		Type: &codegen.StructType{
			Fields: []codegen.StructField{
				{Name: "value", Type: codegen.EmptyInterfaceType{}},
			},
		},
		SchemaType: t,
	}

	g.output.declsBySchema[t] = decl
	g.output.declsByName[decl.Name] = decl
	g.output.file.Package.AddDecl(decl)

	if g.config.OnlyModels {
		return &codegen.NamedType{Decl: decl}, nil
	}

	g.output.file.Package.AddImport("encoding/json", "")
	g.output.file.Package.AddImport("fmt", "")
	g.output.file.Package.AddImport("bytes", "")

	hasYAMLFormatter := false

	for _, f := range g.formatters {
		if f.getName() == formatYAML {
			hasYAMLFormatter = true

			break
		}
	}

	if hasYAMLFormatter {
		g.output.file.Package.AddImport(YAMLPackage, "yaml")
	}

	addMethod := func(suffix string, impl func(*codegen.Emitter) error) {
		g.output.file.Package.AddDecl(&codegen.Method{
			Name: name + "_" + suffix,
			Impl: impl,
		})
	}

	addMethod("UnmarshalJSON", emitOneOfPrimitiveUnmarshalJSON(name, kinds))
	addMethod("MarshalJSON", emitOneOfPrimitiveMarshalJSON(name, kinds))

	if hasYAMLFormatter {
		addMethod("UnmarshalYAML", emitOneOfPrimitiveUnmarshalYAML(name, kinds))
		addMethod("MarshalYAML", emitOneOfPrimitiveMarshalYAML(name, kinds))
	}

	addMethod("Value", emitOneOfPrimitiveValue(name))
	addMethod("IsZero", emitOneOfPrimitiveIsZero(name))

	if kinds.has(oneOfKindString) {
		addMethod("AsString", emitOneOfPrimitiveAsString(name))
	}

	if kinds.has(oneOfKindNumber) {
		addMethod("AsNumber", emitOneOfPrimitiveAsNumber(name))
	}

	if kinds.has(oneOfKindBoolean) {
		addMethod("AsBool", emitOneOfPrimitiveAsBool(name))
	}

	if kinds.has(oneOfKindNull) {
		addMethod("IsNull", emitOneOfPrimitiveIsNull(name))
	}

	return &codegen.NamedType{Decl: decl}, nil
}

func emitOneOfPrimitiveUnmarshalJSON(typeName string, kinds oneOfKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("UnmarshalJSON implements json.Unmarshaler.")
		out.Printlnf("func (j *%s) UnmarshalJSON(value []byte) error {", typeName)
		out.Indent(1)

		out.Printlnf("dec := json.NewDecoder(bytes.NewReader(value))")

		if kinds.has(oneOfKindNumber) {
			out.Printlnf("dec.UseNumber()")
		}

		out.Printlnf("tok, err := dec.Token()")
		out.Printlnf("if err != nil { return err }")
		out.Printlnf("switch tok.(type) {")

		if kinds.has(oneOfKindString) {
			out.Printlnf("case string:")
			out.Indent(1)
			out.Printlnf("var v string")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(oneOfKindNumber) {
			out.Printlnf("case json.Number:")
			out.Indent(1)
			out.Printlnf("var v float64")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(oneOfKindBoolean) {
			out.Printlnf("case bool:")
			out.Indent(1)
			out.Printlnf("var v bool")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(oneOfKindNull) {
			out.Printlnf("case nil:")
			out.Indent(1)
			out.Printlnf("j.value = nil")
			out.Indent(-1)
		}

		out.Printlnf("default:")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: unsupported JSON value of type %%T", tok)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveMarshalJSON(typeName string, kinds oneOfKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("MarshalJSON implements json.Marshaler.")
		out.Printlnf("func (j *%s) MarshalJSON() ([]byte, error) {", typeName)
		out.Indent(1)

		if kinds.has(oneOfKindNull) {
			// Schema includes null as a variant — emit null for the unset
			// value so round-trip preserves the explicit-null variant.
			out.Printlnf("if j == nil || j.value == nil { return []byte(\"null\"), nil }")
		} else {
			// Schema does not include null — refuse to marshal an unset
			// value rather than emit a null that the matching UnmarshalJSON
			// would reject. Round-trip would otherwise be broken.
			out.Printlnf("if j == nil || j.value == nil {")
			out.Indent(1)
			out.Printlnf(`return nil, fmt.Errorf("%s: cannot marshal unset value (schema does not allow null)")`, typeName)
			out.Indent(-1)
			out.Printlnf("}")
		}

		out.Printlnf("return json.Marshal(j.value)")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveUnmarshalYAML(typeName string, kinds oneOfKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("UnmarshalYAML implements yaml.Unmarshaler.")
		out.Printlnf("func (j *%s) UnmarshalYAML(value *yaml.Node) error {", typeName)
		out.Indent(1)
		out.Printlnf("if value.Kind != yaml.ScalarNode {")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: expected scalar YAML node")`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("switch value.Tag {")

		if kinds.has(oneOfKindString) {
			out.Printlnf(`case "!!str":`)
			out.Indent(1)
			out.Printlnf("var v string")
			out.Printlnf("if err := value.Decode(&v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(oneOfKindNumber) {
			out.Printlnf(`case "!!int", "!!float":`)
			out.Indent(1)
			out.Printlnf("var v float64")
			out.Printlnf("if err := value.Decode(&v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(oneOfKindBoolean) {
			out.Printlnf(`case "!!bool":`)
			out.Indent(1)
			out.Printlnf("var v bool")
			out.Printlnf("if err := value.Decode(&v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(oneOfKindNull) {
			out.Printlnf(`case "!!null":`)
			out.Indent(1)
			out.Printlnf("j.value = nil")
			out.Indent(-1)
		}

		out.Printlnf("default:")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: unsupported YAML scalar tag %%q", value.Tag)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveMarshalYAML(typeName string, kinds oneOfKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("MarshalYAML implements yaml.Marshaler.")
		out.Printlnf("func (j *%s) MarshalYAML() (interface{}, error) {", typeName)
		out.Indent(1)

		if kinds.has(oneOfKindNull) {
			out.Printlnf("if j == nil { return nil, nil }")
		} else {
			out.Printlnf("if j == nil || j.value == nil {")
			out.Indent(1)
			out.Printlnf(`return nil, fmt.Errorf("%s: cannot marshal unset value (schema does not allow null)")`, typeName)
			out.Indent(-1)
			out.Printlnf("}")
		}

		out.Printlnf("return j.value, nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveValue(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("Value returns the decoded primitive payload.")
		out.Printlnf("func (j *%s) Value() any {", typeName)
		out.Indent(1)
		out.Printlnf("if j == nil { return nil }")
		out.Printlnf("return j.value")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveIsZero(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("IsZero reports whether the value is unset; supports the encoding/json `omitzero` tag.")
		out.Printlnf("func (j *%s) IsZero() bool {", typeName)
		out.Indent(1)
		out.Printlnf("if j == nil { return true }")
		out.Printlnf("return j.value == nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveAsString(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("AsString returns the value as a string and reports whether it was a string.")
		out.Printlnf("func (j *%s) AsString() (string, bool) {", typeName)
		out.Indent(1)
		out.Printlnf("if j == nil { return \"\", false }")
		out.Printlnf("v, ok := j.value.(string)")
		out.Printlnf("return v, ok")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveAsNumber(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("AsNumber returns the value as a float64 and reports whether it was numeric.")
		out.Printlnf("func (j *%s) AsNumber() (float64, bool) {", typeName)
		out.Indent(1)
		out.Printlnf("if j == nil { return 0, false }")
		out.Printlnf("v, ok := j.value.(float64)")
		out.Printlnf("return v, ok")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveAsBool(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("AsBool returns the value as a bool and reports whether it was a bool.")
		out.Printlnf("func (j *%s) AsBool() (bool, bool) {", typeName)
		out.Indent(1)
		out.Printlnf("if j == nil { return false, false }")
		out.Printlnf("v, ok := j.value.(bool)")
		out.Printlnf("return v, ok")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveIsNull(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("IsNull reports whether the decoded value was an explicit JSON null.")
		out.Printlnf("func (j *%s) IsNull() bool {", typeName)
		out.Indent(1)
		out.Printlnf("return j == nil || j.value == nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}
