package generator

import (
	"errors"
	"fmt"
	"sort"

	"github.com/atombender/go-jsonschema/internal/x/text"
	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// ErrUnexpectedVariantType is returned when the schema generator produces
// a Go type for a oneOf variant that is not the expected *codegen.NamedType.
// This indicates an internal generator inconsistency (e.g. an alias or
// inline map type was returned where a named struct was required).
var ErrUnexpectedVariantType = errors.New("oneOf variant: expected *codegen.NamedType from generator")

// discriminatorResult bundles the outcome of detectDiscriminator so the
// call site can avoid named returns (which the project's linter flags).
// `ok` is the success flag; the other fields are only meaningful when ok
// is true.
type discriminatorResult struct {
	prop   string
	values []discriminatorValue
	ok     bool
}

// detectDiscriminator inspects a `oneOf` and returns the JSON property name
// that uniquely tags each variant, plus a list of (constValue, variantIndex)
// pairs in the original variant order. It returns ok=false when no natural
// discriminator exists; the caller should fall back to a different generation
// strategy (try-each, or interface{} field) in that case.
//
// The algorithm follows the plan in
// .claude/plans/when-i-asked-you-fuzzy-pearl.md, Phase 5:
//
//  1. Bail if any variant carries nested oneOf — those go to phase 6.
//  2. For each variant, build the candidate set
//     {prop : prop ∈ Properties ∧ prop ∈ Required ∧ Const is primitive}.
//     `enum: [x]` (single-value enum) counts as `const: x`.
//  3. Intersect the candidate sets across all variants.
//  4. For each shared candidate, collect const values across variants;
//     the property qualifies as a discriminator only when those values
//     are all distinct.
//  5. If multiple discriminators qualify, pick alphabetically first and warn.
func (g *schemaGenerator) detectDiscriminator(variants []*schemas.Type) discriminatorResult {
	if len(variants) < 2 {
		return discriminatorResult{}
	}

	// Step 1: resolve $ref and flatten allOf for each variant. This is
	// necessary so a variant like
	//   {"allOf":[{"$ref":"Base"},{"properties":{"kind":{"const":"x"}}}]}
	// has its discriminator-bearing const recognised — the candidate scan
	// only inspects the top-level Properties/Required of each variant. Bail
	// out of the whole detection if any variant carries nested oneOf or its
	// $ref/allOf can't be resolved (the caller falls back to phase 6
	// try-each in that case).
	flattened := make([]*schemas.Type, len(variants))

	for i, v := range variants {
		if v == nil {
			return discriminatorResult{}
		}

		f, ok := g.flattenForDiscriminator(v)
		if !ok {
			return discriminatorResult{}
		}

		if len(f.OneOf) > 0 {
			return discriminatorResult{}
		}

		flattened[i] = f
	}

	// Step 2: per-variant candidate sets.
	candidates := make([]map[string]any, len(variants))
	for i, v := range flattened {
		candidates[i] = primitiveConstCandidates(v)
		if len(candidates[i]) == 0 {
			return discriminatorResult{}
		}
	}

	// Step 3: intersect candidate property names across variants.
	shared := candidates[0]
	for _, c := range candidates[1:] {
		next := make(map[string]any, len(shared))

		for name := range shared {
			if _, in := c[name]; in {
				next[name] = nil
			}
		}

		shared = next

		if len(shared) == 0 {
			return discriminatorResult{}
		}
	}

	// Step 4: of the shared candidates, retain those whose values are distinct.
	qualifying := make([]string, 0, len(shared))

	for name := range shared {
		seen := make(map[string]int, len(variants))
		distinct := true

		for i, c := range candidates {
			key := jsonValueKey(c[name])
			if _, dup := seen[key]; dup {
				distinct = false

				break
			}

			seen[key] = i
		}

		if distinct {
			qualifying = append(qualifying, name)
		}
	}

	if len(qualifying) == 0 {
		return discriminatorResult{}
	}

	// Step 5: pick alphabetically first, warn on ties.
	sort.Strings(qualifying)
	prop := qualifying[0]

	if len(qualifying) > 1 {
		g.warner(fmt.Sprintf(
			"oneOf: multiple discriminator candidates %v; picking %q alphabetically",
			qualifying, prop,
		))
	}

	values := make([]discriminatorValue, len(variants))
	for i, c := range candidates {
		values[i] = discriminatorValue{
			constValue:   c[prop],
			variantIndex: i,
		}
	}

	return discriminatorResult{prop: prop, values: values, ok: true}
}

// flattenForDiscriminator returns a schema equivalent to v but with $ref
// resolved and allOf merged into the top level, so primitiveConstCandidates
// sees inherited Properties / Required as if they were declared inline.
//
// Returns ok=false on any structural problem (broken ref, allOf merge
// failure, nested compositions left after one round of flattening) — the
// caller treats that as "no discriminator" and falls through to phase 6.
//
// One round of flattening is sufficient for the common case
// `{allOf:[{$ref:Base},{inline}]}`. Schemas that nest allOf inside allOf
// would need iterative flattening, but those are rare and the phase 6
// try-each fallback handles them correctly even without discriminator
// detection.
func (g *schemaGenerator) flattenForDiscriminator(v *schemas.Type) (*schemas.Type, bool) {
	if v.Ref != "" {
		resolved, err := g.resolveRef(v)
		if err != nil {
			return nil, false
		}

		v = resolved
	}

	if len(v.AllOf) == 0 {
		return v, true
	}

	parts := make([]*schemas.Type, 0, len(v.AllOf))

	for _, child := range v.AllOf {
		// Recursively resolve each branch's ref so the merged result
		// includes their Properties / Required.
		f, ok := g.flattenForDiscriminator(child)
		if !ok {
			return nil, false
		}

		parts = append(parts, f)
	}

	merged, err := schemas.AllOf(parts, v)
	if err != nil {
		return nil, false
	}

	// schemas.AllOf preserves v.AllOf on the result; clear it so a caller
	// re-checking (or a later flattener) sees the merge as final.
	merged.AllOf = nil

	return merged, true
}

// discriminatorValue ties a variant index to the const value of its
// discriminator property.
type discriminatorValue struct {
	constValue   any
	variantIndex int
}

// primitiveConstCandidates returns the subset of v's required properties
// whose schema declares a primitive `const` (or single-value `enum`).
// Non-primitive const values (objects, arrays) are excluded because a
// JSON-token-level discriminator cannot match them with simple equality.
func primitiveConstCandidates(v *schemas.Type) map[string]any {
	required := make(map[string]struct{}, len(v.Required))
	for _, r := range v.Required {
		required[r] = struct{}{}
	}

	out := make(map[string]any, len(v.Properties))

	for name, prop := range v.Properties {
		if prop == nil {
			continue
		}

		if _, ok := required[name]; !ok {
			continue
		}

		val, hasConst := primitiveConstValue(prop)
		if !hasConst {
			continue
		}

		out[name] = val
	}

	return out
}

// primitiveConstValue returns the (value, true) of a property's primitive
// constraint when it has one — either a `const: x` declaration or a
// single-value `enum: [x]` (which is semantically equivalent). Returns
// (nil, false) when the property has neither, or when the declared value
// is non-primitive (object/array). Used by primitiveConstCandidates to
// decide whether a property qualifies as a discriminator candidate.
func primitiveConstValue(p *schemas.Type) (any, bool) {
	switch {
	case p.Const != nil:
		if isPrimitiveJSONValue(p.Const) {
			return p.Const, true
		}
	case len(p.Enum) == 1:
		if isPrimitiveJSONValue(p.Enum[0]) {
			return p.Enum[0], true
		}
	}

	return nil, false
}

// isPrimitiveJSONValue is a type predicate matching values that JSON's
// scalar tokens decode into via Go's encoding/json (or our schema parser):
// string, bool, float64, int, int64, or nil. Used to filter out
// composite-typed const values (objects, arrays) that can't be matched
// by a token-level discriminator dispatch.
func isPrimitiveJSONValue(v any) bool {
	switch v.(type) {
	case string, bool, float64, int, int64, nil:
		return true
	}

	return false
}

// jsonValueKey returns a stable string key for a primitive JSON value so the
// distinct-value check can compare across types — `1` (int) and `1.0` (float)
// must be considered equal because JSON does not distinguish them on the wire.
func jsonValueKey(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case bool:
		if x {
			return "bool:true"
		}

		return "bool:false"
	case string:
		return "string:" + x
	case float64:
		return fmt.Sprintf("number:%g", x)
	case int:
		return fmt.Sprintf("number:%g", float64(x))
	case int64:
		return fmt.Sprintf("number:%g", float64(x))
	}

	return fmt.Sprintf("other:%T:%v", v, v)
}

// variantBinding pairs a generated variant decl with its holder field name
// and the JSON-encoded form of its discriminator value (for the dispatch
// switch in UnmarshalJSON).
type variantBinding struct {
	decl      *codegen.TypeDecl // generated struct for the variant
	fieldName string            // field name in the holder, e.g. "Dog"
	jsonLitGo string            // Go literal that matches the JSON token, e.g. `"\"dog\""` or `"42"`
	// Go literal for the YAML peek (decoded scalar), e.g. `"dog"` / `"42"`.
	// Unused when constIsNull (null is dispatched via the YAML tag instead).
	yamlLitGo string
	// constValue is the underlying Go value of the discriminator's `const`
	// for this variant (e.g. string, float64, int, bool, nil). Retained so
	// emit code can decide between lexeme-based dispatch (for strings,
	// booleans, null) and value-based dispatch (for numerics, where JSON
	// allows equivalent representations like `1`, `1.0`, `1e0`).
	constValue any
	// constIsNull marks the binding as the `const: null` variant. The YAML
	// emit dispatches it via a `Tag == "!!null"` check rather than the
	// string switch, so a literal empty-string discriminator value cannot
	// collide with the null variant (and so a missing-field error is
	// reported instead of silently routing to null).
	constIsNull bool
}

// bindingsAreNumeric reports whether every non-null binding's discriminator
// const is a number (float64, int, int64). When true, emit code dispatches
// on a parsed numeric value rather than the raw token text — necessary
// because JSON inputs `1`, `1.0`, `1e0` are equivalent integer encodings
// but produce distinct token text and so wouldn't all match a single
// `case "1":` in a string-switch.
//
// Mixed-type discriminators (e.g. some strings, some numbers — unusual but
// not formally forbidden) keep the lexeme-based dispatch since there's no
// single normalization that works across types.
func bindingsAreNumeric(bindings []variantBinding) bool {
	hasNumeric := false

	for _, b := range bindings {
		if b.constIsNull {
			continue
		}

		switch b.constValue.(type) {
		case float64, int, int64:
			hasNumeric = true
		default:
			return false
		}
	}

	return hasNumeric
}

// generateOneOfDiscriminator emits a holder type with per-variant pointer
// fields, plus the variant types themselves, plus UnmarshalJSON/MarshalJSON
// (and YAML counterparts when ExtraImports is on) that dispatch on the
// discriminator property.
//
// Example shape for `oneOf:[{kind:"dog", barkAt}, {kind:"cat", purr}]`:
//
//	type Animal struct { Dog *AnimalDog; Cat *AnimalCat }
//	type AnimalDog struct { Kind string `json:"kind"`; BarkAt string `json:"barkAt"` }
//	type AnimalCat struct { Kind string `json:"kind"`; Purr   bool   `json:"purr"`   }
//
// UnmarshalJSON peeks the discriminator property, dispatches into the
// matching variant, and assigns its pointer field. MarshalJSON scans the
// pointer fields and marshals the single non-nil variant (errors on zero
// or more than one set).
func (g *schemaGenerator) generateOneOfDiscriminator(
	t *schemas.Type,
	scope nameScope,
	prop string,
	values []discriminatorValue,
) (codegen.Type, error) {
	holderName := g.output.uniqueTypeName(scope)
	if g.config.StructNameFromTitle && t.Title != "" {
		holderName = g.caser.Identifierize(t.Title)
	}

	bindings := make([]variantBinding, len(t.OneOf))

	for i, variant := range t.OneOf {
		dv := values[i]
		fieldName := variantFieldName(g.caser, dv.constValue)
		variantScope := scope.add(fieldName)

		// Force generation of the variant as its own struct (not inlined).
		// Clone the schema first so the mutation doesn't contaminate any
		// shared cache entry: t.OneOf entries can come from a resolved $ref,
		// in which case the same *Type may be referenced from elsewhere in
		// the schema graph.
		variantClone := *variant
		variantClone.SetSubSchemaTypeElem()

		genType, err := g.generateDeclaredType(&variantClone, variantScope)
		if err != nil {
			return nil, fmt.Errorf("oneOf variant %d (%v): %w", i, dv.constValue, err)
		}

		nt, ok := genType.(*codegen.NamedType)
		if !ok {
			return nil, fmt.Errorf("%w (variant %d, got %T)", ErrUnexpectedVariantType, i, genType)
		}

		bindings[i] = variantBinding{
			decl:        nt.Decl,
			fieldName:   fieldName,
			jsonLitGo:   jsonGoLiteralForJSONToken(dv.constValue),
			yamlLitGo:   jsonGoLiteralForYAMLScalar(dv.constValue),
			constValue:  dv.constValue,
			constIsNull: dv.constValue == nil,
		}
	}

	// Build the holder struct: one pointer field per variant. The fields
	// carry json:"-" / yaml:"-" so the standard struct decoder ignores
	// them — our custom Unmarshal* methods dispatch instead.
	holderStruct := &codegen.StructType{}

	for _, b := range bindings {
		holderStruct.Fields = append(holderStruct.Fields, codegen.StructField{
			Name: b.fieldName,
			Type: &codegen.PointerType{Type: &codegen.NamedType{Decl: b.decl}},
			Tags: `json:"-" yaml:"-"`,
		})
	}

	holderDecl := &codegen.TypeDecl{
		Name:       holderName,
		Comment:    t.Description,
		Type:       holderStruct,
		SchemaType: t,
	}

	g.output.declsBySchema[t] = holderDecl
	g.output.declsByName[holderDecl.Name] = holderDecl
	g.output.file.Package.AddDecl(holderDecl)

	if g.config.OnlyModels {
		return &codegen.NamedType{Decl: holderDecl}, nil
	}

	g.output.file.Package.AddImport("encoding/json", "")
	g.output.file.Package.AddImport("fmt", "")

	hasYAML := false

	for _, f := range g.formatters {
		if f.getName() == formatYAML {
			hasYAML = true

			break
		}
	}

	if hasYAML {
		g.output.file.Package.AddImport(YAMLPackage, "yaml")

		if bindingsAreNumeric(bindings) {
			// strconv.ParseFloat is used in the YAML emit to parse the
			// scalar text into a float64 for value-based dispatch.
			g.output.file.Package.AddImport("strconv", "")
		}
	}

	addMethod := func(suffix string, impl func(*codegen.Emitter) error) {
		g.output.file.Package.AddDecl(&codegen.Method{
			Name: holderName + "_" + suffix,
			Impl: impl,
		})
	}

	addMethod("UnmarshalJSON", emitOneOfDiscriminatorUnmarshalJSON(holderName, prop, bindings))
	addMethod("MarshalJSON", emitOneOfDiscriminatorMarshalJSON(holderName, bindings))

	if hasYAML {
		addMethod("UnmarshalYAML", emitOneOfDiscriminatorUnmarshalYAML(holderName, prop, bindings))
		addMethod("MarshalYAML", emitOneOfDiscriminatorMarshalYAML(holderName, bindings))
	}

	return &codegen.NamedType{Decl: holderDecl}, nil
}

// emitOneOfDiscriminatorUnmarshalJSON returns the codegen callback that
// writes the holder type's UnmarshalJSON. The body resets the holder,
// peeks the discriminator JSON token via a json.RawMessage scratch
// struct, then dispatches: string-typed discriminators go through a
// `switch string(peek.Discriminator)` lexeme match; numeric-typed
// discriminators (per bindingsAreNumeric) parse the token into a
// float64 and switch on the value so equivalent encodings (1 / 1.0 /
// 1e0) all match.
func emitOneOfDiscriminatorUnmarshalJSON(
	typeName, prop string,
	bindings []variantBinding,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("UnmarshalJSON implements json.Unmarshaler. It peeks the")
		out.Commentf("discriminator property %q and dispatches the decode into the", prop)
		out.Commentf("matching variant.")
		out.Printlnf("func (j *%s) UnmarshalJSON(value []byte) error {", typeName)
		out.Indent(1)
		out.Commentf("Reset to zero value so reusing the same holder across multiple")
		out.Commentf("Unmarshal calls doesn't leave a previous winner set alongside the")
		out.Commentf("new one (which would violate the one-variant-set invariant and")
		out.Commentf("break the corresponding MarshalJSON).")
		out.Printlnf("*j = %s{}", typeName)
		out.Printlnf("var peek struct {")
		out.Indent(1)
		out.Printlnf("Discriminator json.RawMessage `json:%q`", prop)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if err := json.Unmarshal(value, &peek); err != nil {")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("unmarshal %s: %%w", err)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if len(peek.Discriminator) == 0 {")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: missing discriminator field %s")`, typeName, prop)
		out.Indent(-1)
		out.Printlnf("}")

		if bindingsAreNumeric(bindings) {
			// Numeric-typed discriminator: parse the raw JSON token as a
			// number and switch on the value, so equivalent encodings
			// (1 / 1.0 / 1e0) all match the same case.
			out.Printlnf("var disc float64")
			out.Printlnf("if err := json.Unmarshal(peek.Discriminator, &disc); err != nil {")
			out.Indent(1)
			out.Printlnf(
				`return fmt.Errorf("%s: %s discriminator must be numeric: %%w", err)`,
				typeName, prop,
			)
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf("switch disc {")
		} else {
			out.Printlnf("switch string(peek.Discriminator) {")
		}

		for _, b := range bindings {
			caseLit := b.jsonLitGo
			if bindingsAreNumeric(bindings) {
				caseLit = numericCaseLiteral(b.constValue)
			}

			out.Printlnf("case %s:", caseLit)
			out.Indent(1)
			out.Printlnf("var v %s", b.decl.Name)
			out.Printlnf("if err := json.Unmarshal(value, &v); err != nil {")
			out.Indent(1)
			out.Printlnf(`return fmt.Errorf("%s.%s: %%w", err)`, typeName, b.fieldName)
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf("j.%s = &v", b.fieldName)
			out.Indent(-1)
		}

		out.Printlnf("default:")
		out.Indent(1)

		if bindingsAreNumeric(bindings) {
			out.Printlnf(
				`return fmt.Errorf("%s: unknown %s value %%v", disc)`,
				typeName, prop,
			)
		} else {
			out.Printlnf(
				`return fmt.Errorf("%s: unknown %s value %%s", string(peek.Discriminator))`,
				typeName, prop,
			)
		}

		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

// numericCaseLiteral renders a Go switch-case literal for a numeric
// discriminator constValue. Integral values emit as untyped int (e.g.
// `1`), non-integral as `%g` (e.g. `1.5`). The case value is implicitly
// converted to float64 by the surrounding switch on a `float64` discriminant.
func numericCaseLiteral(v any) string {
	switch x := v.(type) {
	case float64:
		return numberSuffix(x)
	case int:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	}

	return fmt.Sprintf("%v", v)
}

// emitOneOfDiscriminatorMarshalJSON returns the codegen callback that
// writes the holder type's MarshalJSON. The body counts non-nil variant
// pointers and errors when the count isn't exactly one (zero = ambiguous
// holder; >1 = invalid round-trip), then marshals the single matched
// variant's contents directly via json.Marshal.
func emitOneOfDiscriminatorMarshalJSON(
	typeName string,
	bindings []variantBinding,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("MarshalJSON implements json.Marshaler. Exactly one variant")
		out.Commentf("pointer must be non-nil; otherwise marshaling errors.")
		out.Printlnf("func (j %s) MarshalJSON() ([]byte, error) {", typeName)
		out.Indent(1)
		out.Printlnf("set := 0")

		for _, b := range bindings {
			out.Printlnf("if j.%s != nil { set++ }", b.fieldName)
		}

		out.Printlnf("if set != 1 {")
		out.Indent(1)
		out.Printlnf(`return nil, fmt.Errorf("%s: exactly one variant must be set, got %%d", set)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")

		for _, b := range bindings {
			out.Printlnf("if j.%s != nil { return json.Marshal(j.%s) }", b.fieldName, b.fieldName)
		}

		out.Printlnf("return nil, nil // unreachable")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

// emitOneOfDiscriminatorUnmarshalYAML mirrors the JSON path but for YAML
// inputs. It peeks the discriminator into a yaml.Node so it can
// distinguish missing (Kind == 0) from explicit-null (Tag == "!!null")
// from a scalar-text value, then dispatches: null variants are matched
// via the YAML tag check; numeric-typed discriminators parse the scalar
// text via strconv.ParseFloat and switch on the value; string-typed
// discriminators switch on the lexeme.
func emitOneOfDiscriminatorUnmarshalYAML(
	typeName, prop string,
	bindings []variantBinding,
) func(*codegen.Emitter) error {
	// Split out the (optional) null variant: it dispatches via the YAML tag
	// rather than through the scalar-text switch. Mixing them leads to a
	// `case "":` literal that collides with both an explicit empty-string
	// scalar and a missing-field decode (yaml.v3 produces `""` for both
	// when the peek field is typed `string`).
	var nullBinding *variantBinding

	scalarBindings := make([]variantBinding, 0, len(bindings))

	for i := range bindings {
		if bindings[i].constIsNull {
			nullBinding = &bindings[i]

			continue
		}

		scalarBindings = append(scalarBindings, bindings[i])
	}

	return func(out *codegen.Emitter) error {
		out.Commentf("UnmarshalYAML mirrors UnmarshalJSON: peek the discriminator")
		out.Commentf("then dispatch to the matching variant.")
		out.Printlnf("func (j *%s) UnmarshalYAML(value *yaml.Node) error {", typeName)
		out.Indent(1)
		out.Commentf("Reset to zero value so reusing the same holder across multiple")
		out.Commentf("Unmarshal calls doesn't leave a previous winner set alongside the")
		out.Commentf("new one (which would violate the one-variant-set invariant and")
		out.Commentf("break the corresponding MarshalYAML).")
		out.Printlnf("*j = %s{}", typeName)
		// Peek into a yaml.Node so we can distinguish missing vs explicit
		// null (Tag == "!!null") vs an actual scalar value via .Value.
		out.Printlnf("var peek struct {")
		out.Indent(1)
		out.Printlnf("Discriminator yaml.Node `yaml:%q`", prop)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if err := value.Decode(&peek); err != nil {")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("unmarshal %s: %%w", err)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if peek.Discriminator.Kind == 0 {")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: missing discriminator field %s")`, typeName, prop)
		out.Indent(-1)
		out.Printlnf("}")

		if nullBinding != nil {
			out.Printlnf(`if peek.Discriminator.Tag == "!!null" {`)
			out.Indent(1)
			emitYAMLDispatchAssign(out, typeName, *nullBinding)
			out.Printlnf("return nil")
			out.Indent(-1)
			out.Printlnf("}")
		}

		if bindingsAreNumeric(scalarBindings) {
			// Numeric-typed discriminator: parse the YAML scalar text as a
			// number so equivalent encodings (1 / 1.0 / 1e0) all match
			// the same case. yaml.v3 always exposes the scalar text via
			// .Value; converting via strconv keeps the dispatch consistent
			// with the JSON path (which uses json.Unmarshal into float64).
			out.Printlnf("disc, err := strconv.ParseFloat(peek.Discriminator.Value, 64)")
			out.Printlnf("if err != nil {")
			out.Indent(1)
			out.Printlnf(
				`return fmt.Errorf("%s: %s discriminator must be numeric: %%w", err)`,
				typeName, prop,
			)
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf("switch disc {")
		} else {
			out.Printlnf("switch peek.Discriminator.Value {")
		}

		for _, b := range scalarBindings {
			caseLit := b.yamlLitGo
			if bindingsAreNumeric(scalarBindings) {
				caseLit = numericCaseLiteral(b.constValue)
			}

			out.Printlnf("case %s:", caseLit)
			out.Indent(1)
			emitYAMLDispatchAssign(out, typeName, b)
			out.Indent(-1)
		}

		out.Printlnf("default:")
		out.Indent(1)

		if bindingsAreNumeric(scalarBindings) {
			out.Printlnf(
				`return fmt.Errorf("%s: unknown %s value %%v", disc)`,
				typeName, prop,
			)
		} else {
			out.Printlnf(
				`return fmt.Errorf("%s: unknown %s value %%q", peek.Discriminator.Value)`,
				typeName, prop,
			)
		}

		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

// emitYAMLDispatchAssign emits the per-case body that decodes the full YAML
// value into the variant's struct and assigns the holder's pointer field.
// Shared between the null-tag branch and the scalar-value switch.
func emitYAMLDispatchAssign(out *codegen.Emitter, typeName string, b variantBinding) {
	out.Printlnf("var v %s", b.decl.Name)
	out.Printlnf("if err := value.Decode(&v); err != nil {")
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("%s.%s: %%w", err)`, typeName, b.fieldName)
	out.Indent(-1)
	out.Printlnf("}")
	out.Printlnf("j.%s = &v", b.fieldName)
}

// emitOneOfDiscriminatorMarshalYAML mirrors emitOneOfDiscriminatorMarshalJSON
// for YAML output: same exactly-one-variant invariant, but returns the
// matched variant's payload as an interface{} (yaml.v3 then handles the
// actual encoding) rather than producing a []byte directly.
func emitOneOfDiscriminatorMarshalYAML(
	typeName string,
	bindings []variantBinding,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("MarshalYAML mirrors MarshalJSON.")
		out.Printlnf("func (j %s) MarshalYAML() (interface{}, error) {", typeName)
		out.Indent(1)
		out.Printlnf("set := 0")

		for _, b := range bindings {
			out.Printlnf("if j.%s != nil { set++ }", b.fieldName)
		}

		out.Printlnf("if set != 1 {")
		out.Indent(1)
		out.Printlnf(`return nil, fmt.Errorf("%s: exactly one variant must be set, got %%d", set)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")

		for _, b := range bindings {
			out.Printlnf("if j.%s != nil { return j.%s, nil }", b.fieldName, b.fieldName)
		}

		out.Printlnf("return nil, nil // unreachable")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

// variantFieldName builds the holder's field name for a variant. The string
// case routes through Caser.Identifierize so that --capitalization overrides,
// Go-keyword guards, and leading-non-letter handling stay consistent with
// the rest of the generator. Synthetic suffixes (Const42, True, Null) are
// already valid Go identifiers and bypass the caser.
func variantFieldName(caser *text.Caser, constValue any) string {
	switch x := constValue.(type) {
	case string:
		return caser.Identifierize(x)
	case bool:
		if x {
			return "True"
		}

		return "False"
	case nil:
		return "Null"
	case float64:
		return fmt.Sprintf("Const%s", numberSuffix(x))
	case int:
		return fmt.Sprintf("Const%d", x)
	case int64:
		return fmt.Sprintf("Const%d", x)
	}

	return fmt.Sprintf("ConstV%v", constValue)
}

// numberSuffix renders a float64 as a Go numeric literal in its shortest
// canonical form: integer-valued floats become unsuffixed integer text
// ("1" rather than "1.000000"), non-integer values use %g compact format.
// Used when constructing JSON-token literals and Go switch-case values
// for numeric discriminator dispatch.
func numberSuffix(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}

	return fmt.Sprintf("%g", f)
}

// jsonGoLiteralForJSONToken returns a Go string literal whose value is the
// JSON-encoded form of constValue. The `peek` struct in UnmarshalJSON receives
// the raw JSON token bytes via json.RawMessage, so the case must compare
// against the wire form: strings include the surrounding quotes, numbers
// and booleans are stringified naturally.
func jsonGoLiteralForJSONToken(constValue any) string {
	switch x := constValue.(type) {
	case string:
		// Want a Go literal whose VALUE is `"x"` (with quotes). %q produces
		// `"x"` from `x`; we then need to wrap that result in a Go string,
		// so use %q twice would over-escape. Use Sprintf with the literal
		// JSON quoting:
		return fmt.Sprintf("%q", fmt.Sprintf("%q", x))
	case bool:
		return fmt.Sprintf("%q", fmt.Sprintf("%t", x))
	case nil:
		return `"null"`
	case float64:
		return fmt.Sprintf("%q", numberSuffix(x))
	case int:
		return fmt.Sprintf("%q", fmt.Sprintf("%d", x))
	case int64:
		return fmt.Sprintf("%q", fmt.Sprintf("%d", x))
	}

	return fmt.Sprintf("%q", fmt.Sprintf("%v", constValue))
}

// jsonGoLiteralForYAMLScalar returns a Go literal matching the decoded
// scalar form a YAML decoder would produce for the discriminator value.
// We peek into a `string` field so YAML's tag-based scalar coercion gives
// us the canonical text; for booleans/numbers/null the YAML library
// stringifies the same way the JSON wire form does.
func jsonGoLiteralForYAMLScalar(constValue any) string {
	switch x := constValue.(type) {
	case string:
		return fmt.Sprintf("%q", x)
	case bool:
		return fmt.Sprintf("%q", fmt.Sprintf("%t", x))
	case nil:
		return `""`
	case float64:
		return fmt.Sprintf("%q", numberSuffix(x))
	case int:
		return fmt.Sprintf("%q", fmt.Sprintf("%d", x))
	case int64:
		return fmt.Sprintf("%q", fmt.Sprintf("%d", x))
	}

	return fmt.Sprintf("%q", fmt.Sprintf("%v", constValue))
}
