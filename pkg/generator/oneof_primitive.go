package generator

import (
	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// primitiveKind is a bit-set of JSON kinds the wrapper-type emission
// strategy can dispatch on. Used by both the primitive `oneOf` path and
// the multi-type union path. `integer` is intentionally not represented
// here: schemas including `integer` are routed away from the wrapper path
// because the wire-level JSON token kind cannot distinguish `1` from `1.5`,
// and silently widening would violate the schema.
type primitiveKind uint8

const (
	primitiveKindString primitiveKind = 1 << iota
	primitiveKindNumber
	primitiveKindBoolean
	primitiveKindNull
)

func (k primitiveKind) has(other primitiveKind) bool { return k&other == other }

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

	// seen tracks the primitive kinds we've already accepted in this
	// variant list. Duplicates mean two branches dispatch on the same JSON
	// token kind, which the wrapper cannot represent: collapsing them via
	// the kinds bitset would let one input satisfy multiple branches,
	// violating JSON Schema's oneOf "exactly one matches" rule. Decline
	// detection in that case and let the generic path handle (or reject)
	// the schema.
	seen := make(map[string]struct{}, len(t.OneOf))

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

		if primitiveHasValidationConstraints(variant) {
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

		if _, dup := seen[variant.Type[0]]; dup {
			return false
		}

		seen[variant.Type[0]] = struct{}{}
	}

	return true
}

// primitiveHasValidationConstraints reports whether a schema declares any
// constraint that the primitive-wrapper emission strategy can't honor.
// Called on both single oneOf variants AND multi-type union schemas (the
// constraint set is the same in both contexts). The wrapper only
// dispatches on the JSON token kind, so anything needing e.g. `format`,
// `minimum`, or `pattern` checked would silently pass invalid values if
// routed through this path.
func primitiveHasValidationConstraints(v *schemas.Type) bool {
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
func primitiveOneOfKinds(t *schemas.Type) primitiveKind {
	var kinds primitiveKind

	for _, variant := range t.OneOf {
		switch variant.Type[0] {
		case schemas.TypeNameString:
			kinds |= primitiveKindString
		case schemas.TypeNameNumber:
			kinds |= primitiveKindNumber
		case schemas.TypeNameBoolean:
			kinds |= primitiveKindBoolean
		case schemas.TypeNameNull:
			kinds |= primitiveKindNull
		}
	}

	return kinds
}

// generateOneOfPrimitive emits a wrapper type and the methods needed to
// round-trip a primitive `oneOf` schema. The generated type stores the
// decoded value in `value any` and tracks whether the wrapper has been
// populated (by Unmarshal{JSON,YAML}) in `present bool` so that three
// states stay distinct: untouched / decoded `null` / decoded primitive.
// Without `present` the helpers and `omitzero` round-trip would conflate
// the first two — explicit nulls would silently become "missing".
// UnmarshalJSON dispatches on the JSON token kind so that overlapping
// schemas (e.g. number+integer) cannot match more than one variant by
// accident.
//
// Precondition: callers MUST NOT invoke this in `OnlyModels` mode — the
// wrapper struct has unexported `value`/`present` fields and is unusable
// without the marshalers/accessors emitted below. `OnlyModels` semantics
// require the model type to be readable/writable by external code, which
// this wrapper isn't. Callers in `generateDeclaredType` already gate on
// `!g.config.OnlyModels`; this function panics rather than silently
// emitting a broken type if that guard is ever bypassed.
func (g *schemaGenerator) generateOneOfPrimitive(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if g.config.OnlyModels {
		// Defensive: should be unreachable per the contract above.
		// Panicking here turns a future routing bug (silent broken
		// codegen) into a loud failure at codegen time.
		panic("generateOneOfPrimitive called in OnlyModels mode; caller must guard")
	}

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
				{Name: "present", Type: codegen.PrimitiveType{Type: "bool"}},
			},
		},
		SchemaType: t,
	}

	g.output.declsBySchema[t] = decl
	g.output.declsByName[decl.Name] = decl
	g.output.file.Package.AddDecl(decl)

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

	if kinds.has(primitiveKindString) {
		addMethod("AsString", emitOneOfPrimitiveAsString(name))
	}

	if kinds.has(primitiveKindNumber) {
		addMethod("AsNumber", emitOneOfPrimitiveAsNumber(name))
	}

	if kinds.has(primitiveKindBoolean) {
		addMethod("AsBool", emitOneOfPrimitiveAsBool(name))
	}

	if kinds.has(primitiveKindNull) {
		addMethod("IsNull", emitOneOfPrimitiveIsNull(name))
	}

	return &codegen.NamedType{Decl: decl}, nil
}

func emitOneOfPrimitiveUnmarshalJSON(typeName string, kinds primitiveKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("UnmarshalJSON implements json.Unmarshaler.")
		out.Printlnf("func (j *%s) UnmarshalJSON(value []byte) error {", typeName)
		out.Indent(1)

		out.Printlnf("dec := json.NewDecoder(bytes.NewReader(value))")

		if kinds.has(primitiveKindNumber) {
			out.Printlnf("dec.UseNumber()")
		}

		out.Printlnf("tok, err := dec.Token()")
		out.Printlnf("if err != nil { return err }")
		out.Printlnf("switch tok.(type) {")

		if kinds.has(primitiveKindString) {
			out.Printlnf("case string:")
			out.Indent(1)
			out.Printlnf("var v string")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(primitiveKindNumber) {
			out.Printlnf("case json.Number:")
			out.Indent(1)
			out.Printlnf("var v float64")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(primitiveKindBoolean) {
			out.Printlnf("case bool:")
			out.Indent(1)
			out.Printlnf("var v bool")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(primitiveKindNull) {
			out.Printlnf("case nil:")
			out.Indent(1)
			// Validate the full payload to reject trailing garbage like
			// `null 123`. dec.Token() above only consumed the leading null
			// token; without this, the other branches' json.Unmarshal call
			// asymmetry would silently accept malformed input here.
			out.Printlnf("var v any")
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil { return err }")
			out.Printlnf("j.value = nil")
			out.Indent(-1)
		}

		out.Printlnf("default:")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: unsupported JSON value of type %%T", tok)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("j.present = true")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveMarshalJSON(typeName string, kinds primitiveKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("MarshalJSON implements json.Marshaler.")
		out.Printlnf("func (j *%s) MarshalJSON() ([]byte, error) {", typeName)
		out.Indent(1)

		if kinds.has(primitiveKindNull) {
			// Unset (no Unmarshal call has touched j) → emit null. The
			// receiver's surrounding struct is responsible for using
			// `omitzero` if it wants to omit the field entirely.
			out.Printlnf("if j == nil || !j.present { return []byte(\"null\"), nil }")
			// Explicit JSON null was decoded → preserve it.
			out.Printlnf("if j.value == nil { return []byte(\"null\"), nil }")
		} else {
			// Schema does not include null → refuse to marshal either an
			// unset wrapper or a wrapper whose value is nil (which would
			// round-trip as null and the matching UnmarshalJSON would
			// reject). Forces round-trip correctness.
			out.Printlnf("if j == nil || !j.present {")
			out.Indent(1)
			out.Printlnf(`return nil, fmt.Errorf("%s: cannot marshal unset value (schema does not allow null)")`, typeName)
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf("if j.value == nil {")
			out.Indent(1)
			out.Printlnf(`return nil, fmt.Errorf("%s: cannot marshal nil value (schema does not allow null)")`, typeName)
			out.Indent(-1)
			out.Printlnf("}")
		}

		out.Printlnf("return json.Marshal(j.value)")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveUnmarshalYAML(typeName string, kinds primitiveKind) func(*codegen.Emitter) error {
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

		if kinds.has(primitiveKindString) {
			out.Printlnf(`case "!!str":`)
			out.Indent(1)
			out.Printlnf("var v string")
			out.Printlnf("if err := value.Decode(&v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(primitiveKindNumber) {
			out.Printlnf(`case "!!int", "!!float":`)
			out.Indent(1)
			out.Printlnf("var v float64")
			out.Printlnf("if err := value.Decode(&v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(primitiveKindBoolean) {
			out.Printlnf(`case "!!bool":`)
			out.Indent(1)
			out.Printlnf("var v bool")
			out.Printlnf("if err := value.Decode(&v); err != nil { return err }")
			out.Printlnf("j.value = v")
			out.Indent(-1)
		}

		if kinds.has(primitiveKindNull) {
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
		out.Printlnf("j.present = true")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveMarshalYAML(typeName string, kinds primitiveKind) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("MarshalYAML implements yaml.Marshaler.")
		out.Printlnf("func (j *%s) MarshalYAML() (interface{}, error) {", typeName)
		out.Indent(1)

		if kinds.has(primitiveKindNull) {
			// Unset → emit null. Explicit null also yields nil (yaml.v3
			// emits `null` for a nil interface).
			out.Printlnf("if j == nil || !j.present { return nil, nil }")
		} else {
			out.Printlnf("if j == nil || !j.present {")
			out.Indent(1)
			out.Printlnf(`return nil, fmt.Errorf("%s: cannot marshal unset value (schema does not allow null)")`, typeName)
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf("if j.value == nil {")
			out.Indent(1)
			out.Printlnf(`return nil, fmt.Errorf("%s: cannot marshal nil value (schema does not allow null)")`, typeName)
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
		out.Commentf(
			"IsZero reports whether the wrapper has not been populated by " +
				"Unmarshal{JSON,YAML}; supports the encoding/json `omitzero` tag. " +
				"Note: an explicitly-decoded JSON `null` is NOT zero — see IsNull.",
		)
		out.Printlnf("func (j *%s) IsZero() bool {", typeName)
		out.Indent(1)
		out.Printlnf("return j == nil || !j.present")
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
		out.Printlnf("if j == nil || !j.present { return \"\", false }")
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
		out.Printlnf("if j == nil || !j.present { return 0, false }")
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
		out.Printlnf("if j == nil || !j.present { return false, false }")
		out.Printlnf("v, ok := j.value.(bool)")
		out.Printlnf("return v, ok")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func emitOneOfPrimitiveIsNull(typeName string) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf(
			"IsNull reports whether the wrapper was populated with an explicit " +
				"JSON `null`. Returns false for unset wrappers (use IsZero) and " +
				"for wrappers holding a non-null primitive.",
		)
		out.Printlnf("func (j *%s) IsNull() bool {", typeName)
		out.Indent(1)
		out.Printlnf("return j != nil && j.present && j.value == nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}
