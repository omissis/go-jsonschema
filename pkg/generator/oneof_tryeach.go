package generator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// isTryEachOneOfCandidate reports whether a `oneOf` qualifies for the
// Phase-6 try-each fallback. We require every variant to be an object
// (or schema-less so it can be inferred as such) and to NOT contain
// nested compositions, because the dispatch emits one struct field per
// variant and tries to unmarshal into each in turn.
//
// Caller should only invoke this when detectDiscriminator already
// returned ok=false — try-each is the fallback strategy when a natural
// discriminator can't be identified.
func isTryEachOneOfCandidate(variants []*schemas.Type) bool {
	if len(variants) < 2 {
		return false
	}

	for _, v := range variants {
		if v == nil {
			return false
		}

		if len(v.OneOf)+len(v.AnyOf)+len(v.AllOf) > 0 {
			return false
		}

		// Variants with patternProperties can't be represented faithfully
		// by the generated struct: the variant has no storage for keys
		// that match a pattern, and the per-pattern type/format
		// constraints aren't enforced. Accepting the variant here would
		// mean inputs like {"name":"n","x-a":123} pass shape check (x-a
		// matches the pattern), the variant unmarshals successfully
		// (dropping x-a), and the round-trip silently loses data while
		// also failing to reject the value-type violation. Decline
		// detection — the schema falls through to the generic path.
		if len(v.PatternProperties) > 0 {
			return false
		}

		// Variants with no explicit type but Properties (or empty body)
		// are treated as object — generators infer them as such elsewhere.
		if len(v.Type) == 0 {
			if len(v.Properties) == 0 {
				return false
			}

			continue
		}

		if len(v.Type) != 1 || v.Type[0] != schemas.TypeNameObject {
			return false
		}
	}

	return true
}

// isTryEachCandidate reports whether the `oneOf` of t qualifies for the
// try-each fallback in either of its forms: object variants dispatched on the
// top-level key-set, or array variants dispatched on their element key-sets.
// Callers should only reach this once detectDiscriminator has declined.
//
// A t declaring patternProperties is declined for the same reason as a variant
// declaring them: oneOfVariantSchema folds t's patterns into every variant. A t
// declaring its own element schemas is declined too: the holder replaces the
// field that checked them, and only t's length bounds reach the shape check.
func isTryEachCandidate(t *schemas.Type) bool {
	if len(t.PatternProperties) > 0 {
		return false
	}

	if t.Items != nil || t.TupleItems != nil || t.AdditionalItems != nil {
		return false
	}

	return isTryEachOneOfCandidate(t.OneOf) || isTryEachArrayCandidate(t.OneOf)
}

// useTryEach reports whether t's oneOf gets the try-each holder: it must be a
// candidate, every variant must generate as a named type, no variant may
// redefine what t declares (variantRedefinesParent), and its decode must not be
// able to blow up (tryEachCanBlowUp).
func (g *schemaGenerator) useTryEach(t *schemas.Type) bool {
	return isTryEachCandidate(t) && !g.variantGeneratesInlineMap(t) &&
		!g.variantRedefinesParent(t, "") && !g.tryEachCanBlowUp(t)
}

// variantGeneratesInlineMap reports whether an object variant of t's oneOf has
// no properties, even with t's folded in, while DisableCustomTypesForMaps is
// set. Such a variant generates as an inline map rather than a named type, and
// the holder needs one named type per variant.
func (g *schemaGenerator) variantGeneratesInlineMap(t *schemas.Type) bool {
	if !g.config.DisableCustomTypesForMaps || isTryEachArrayCandidate(t.OneOf) {
		return false
	}

	for _, variant := range t.OneOf {
		if variant.Ref != "" {
			if variant = g.lookupRef(variant); variant == nil {
				continue
			}
		}

		if len(withParentObjectKeywords(t, variant).Properties) == 0 {
			return true
		}
	}

	return false
}

// tryEachCanBlowUp reports whether two variants of t's oneOf can both pass the
// shape check while both leading back to t. Each level of such input is then
// decoded once per passing variant, so the work doubles with every level of
// nesting: 285 bytes nested 18 deep took four seconds to reject. Such a oneOf
// keeps the generic representation.
//
// Variants that cannot both pass, because one requires a key the other's
// `additionalProperties: false` refuses, decode at most one way per level and
// are kept, recursive or not.
func (g *schemaGenerator) tryEachCanBlowUp(t *schemas.Type) bool {
	arrayMode := isTryEachArrayCandidate(t.OneOf)
	shapes := make([]variantShape, 0, len(t.OneOf))

	for _, variant := range t.OneOf {
		if variant.Ref != "" {
			if variant = g.lookupRef(variant); variant == nil {
				continue
			}
		}

		merged := withParentObjectKeywords(t, variant)
		if !g.reachesSchema(merged, t, map[*schemas.Type]bool{}) {
			continue
		}

		shapeSource := merged
		if arrayMode {
			shapeSource = arrayElementSchema(merged)
		}

		shape := variantShapeFor(shapeSource)

		for _, other := range shapes {
			if !shapesExclude(shape, other) {
				return true
			}
		}

		shapes = append(shapes, shape)
	}

	return false
}

// shapesExclude reports whether no object can pass both shape checks: one of
// them requires a key the other, being strict, refuses.
func shapesExclude(a, b variantShape) bool {
	return refusesRequiredKey(a, b) || refusesRequiredKey(b, a)
}

// refusesRequiredKey reports whether strict refuses a key that other requires.
// A pattern might accept any key, so a shape with patterns refuses none.
func refusesRequiredKey(strict, other variantShape) bool {
	if !strict.strict || len(strict.patterns) > 0 {
		return false
	}

	for _, key := range other.required {
		if !slices.Contains(strict.knownProperties, key) {
			return true
		}
	}

	return false
}

// reachesSchema reports whether target can be reached from the subschemas of
// from, following $refs without generating anything.
func (g *schemaGenerator) reachesSchema(from, target *schemas.Type, seen map[*schemas.Type]bool) bool {
	if from == nil || seen[from] {
		return false
	}

	seen[from] = true

	next := []*schemas.Type{from.Items, from.AdditionalItems, from.AdditionalProperties}
	next = append(next, from.TupleItems...)
	next = append(next, from.AllOf...)
	next = append(next, from.AnyOf...)
	next = append(next, from.OneOf...)

	for _, name := range sortedKeys(from.Properties) {
		next = append(next, from.Properties[name])
	}

	for _, name := range sortedKeys(from.PatternProperties) {
		next = append(next, from.PatternProperties[name])
	}

	if from.Ref != "" {
		next = append(next, g.lookupRef(from))
	}

	for _, n := range next {
		if n == target || g.reachesSchema(n, target, seen) {
			return true
		}
	}

	return false
}

// lookupRef returns the schema a $ref points at without generating it, or nil
// when it cannot be found.
func (g *schemaGenerator) lookupRef(t *schemas.Type) *schemas.Type {
	if t.Ref == "#" {
		return (*schemas.Type)(g.schema.ObjectAsType)
	}

	defName, fileName, err := g.extractRefNames(t)
	if err != nil {
		return nil
	}

	schema := g.schema

	if fileName != "" {
		if schema, err = g.loader.Load(fileName, g.schemaFileName); err != nil {
			return nil
		}
	}

	if defName == "" {
		return (*schemas.Type)(schema.ObjectAsType)
	}

	def, err := resolveRefPath(schema.Definitions, defName)
	if err != nil {
		return nil
	}

	return def
}

// arrayElementSchema returns the single schema describing every element of an
// array variant: the plain `items` schema, or the sole entry of a one-element
// tuple (the "tuple of exactly one" idiom). It returns nil for anything else —
// a heterogeneous tuple has no single element type, and an array with no
// `items` is unconstrained.
func arrayElementSchema(v *schemas.Type) *schemas.Type {
	if len(v.TupleItems) == 0 {
		return v.Items
	}

	// Same restriction as itemsSchema: only a CLOSED one-element tuple
	// describes every element. An open `items: [A]` leaves positions 1+
	// unconstrained, so its element shape cannot stand in for the whole
	// array when discriminating variants.
	if len(v.TupleItems) != 1 || !tupleIsClosed(v) {
		return nil
	}

	return v.TupleItems[0]
}

// isTryEachArrayCandidate reports whether a `oneOf` qualifies for the try-each
// fallback in its array form: every variant is an array whose elements are
// described by a single object schema carrying declared properties.
//
// Dispatch for these works the same way as the object form, except the shape
// pre-check runs against every element of the decoded array rather than against
// the top-level object. Restricting to object elements is what makes that check
// meaningful — it is the element key-sets that tell the variants apart (in
// CycloneDX's `licenseChoice`, for instance, `license` versus `expression`).
// Arrays of scalars carry no such signal and are declined.
func isTryEachArrayCandidate(variants []*schemas.Type) bool {
	if len(variants) < 2 {
		return false
	}

	for _, v := range variants {
		if v == nil {
			return false
		}

		if len(v.OneOf)+len(v.AnyOf)+len(v.AllOf) > 0 {
			return false
		}

		if len(v.Type) != 1 || v.Type[0] != schemas.TypeNameArray {
			return false
		}

		elem := arrayElementSchema(v)
		if elem == nil || len(elem.PatternProperties) > 0 || len(elem.Properties) == 0 {
			return false
		}

		if len(elem.OneOf)+len(elem.AnyOf)+len(elem.AllOf) > 0 {
			return false
		}

		if len(elem.Type) > 0 && (len(elem.Type) != 1 || elem.Type[0] != schemas.TypeNameObject) {
			return false
		}
	}

	return true
}

// variantShape captures the per-variant key-set used for shape checking.
// strict means additionalProperties is explicitly false, so any input key
// outside knownProperties (and not matching patternProperties) must
// disqualify the variant. When strict is false, extras are allowed and the
// shape check passes regardless of which keys are present.
//
// required lists the variant's required JSON property names. The shape
// check disqualifies a variant when any required key is missing from the
// input, regardless of strict — without this, two variants where one has
// a required field the other doesn't can both successfully decode the
// same partial input, producing an "ambiguous input" error instead of a
// clean dispatch.
type variantShape struct {
	knownProperties []string // sorted JSON property names
	patterns        []string // sorted patternProperties regex strings
	required        []string // sorted required property names
	strict          bool

	// Array mode only: the variant's length bounds.
	minItems    int
	maxItems    int
	hasMaxItems bool
}

// variantShapeFor extracts the per-variant shape used by the try-each
// pre-decode shape check: declared property names, patternProperties
// regexes, the typed/untyped/false flavor of additionalProperties, and
// whether the variant declares any patternProperties at all (which forces
// lenient extras handling, since the generator doesn't yet enforce
// patternProperties at runtime).
func variantShapeFor(v *schemas.Type) variantShape {
	props := make([]string, 0, len(v.Properties))
	for name := range v.Properties {
		props = append(props, name)
	}

	patterns := make([]string, 0, len(v.PatternProperties))
	for pat := range v.PatternProperties {
		patterns = append(patterns, pat)
	}

	required := append([]string{}, v.Required...)

	// Sort for deterministic emission.
	sortStrings(props)
	sortStrings(patterns)
	sortStrings(required)

	strict := isFalseSchema(v.AdditionalProperties)

	return variantShape{
		knownProperties: props,
		patterns:        patterns,
		required:        required,
		strict:          strict,
	}
}

// tighterMaxItems returns the smaller of the length caps two schemas imply, and
// whether either implies one.
func tighterMaxItems(a, b *schemas.Type) (int, bool) {
	aMax, aHas := effectiveMaxItems(a)
	bMax, bHas := effectiveMaxItems(b)

	switch {
	case aHas && bHas:
		return min(aMax, bMax), true

	case aHas:
		return aMax, true

	default:
		return bMax, bHas
	}
}

// sortStrings is a tiny wrapper to avoid importing "sort" purely for one call.
// We already import sort in oneof_discriminator.go for detectDiscriminator's
// alphabetical tie-break; this avoids re-importing here.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		j := i
		for j > 0 && s[j-1] > s[j] {
			s[j-1], s[j] = s[j], s[j-1]
			j--
		}
	}
}

// generateOneOfTryEach emits the same holder + variant types as the
// discriminator path, but dispatches by trying each variant in turn and
// counting successes. The shape check (per-variant allowed-key set) is a
// pre-filter that eliminates obviously-wrong variants before the unmarshal
// attempt, both for performance and for clearer error messages. Final
// outcome:
//
//   - exactly one variant unmarshals successfully → assign its pointer field
//   - zero variants succeed → return the most-recent unmarshal error
//   - more than one variant succeeds → return an ambiguous-input error
//
// MarshalJSON / MarshalYAML are identical to the discriminator path.
func (g *schemaGenerator) generateOneOfTryEach(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	holderName, nameCollided := g.output.uniqueTypeName(scope)

	if g.config.StructNameFromTitle && t.Title != "" {
		// The title supplies the name outright, so whatever the scope would
		// have collided with is no longer relevant.
		holderName = g.caser.Identifierize(t.Title)
		nameCollided = false
	}

	g.output.warnNameCollision(nameCollided, scope, holderName)

	// Registered before the variants are generated, so a variant referring
	// back to this oneOf (a recursive schema) resolves to the holder rather
	// than to a type that is never declared. The type is filled in below.
	holderDecl := &codegen.TypeDecl{
		Name:       holderName,
		Comment:    t.Description,
		SchemaType: t,
	}
	g.output.declsBySchema[t] = holderDecl
	g.output.declsByName[holderDecl.Name] = holderDecl

	arrayMode := isTryEachArrayCandidate(t.OneOf)
	bindings := make([]variantBinding, len(t.OneOf))
	shapes := make([]variantShape, len(t.OneOf))

	for i, variant := range t.OneOf {
		fieldName := fmt.Sprintf("Variant%d", i)
		variantScope := scope.add(fieldName)

		// A copy, carrying the parent's own properties: t.OneOf entries can
		// come from a resolved $ref, in which case the same *Type is
		// referenced from elsewhere in the schema graph. Setting the flag on
		// the original would contaminate the shared cache.
		variantClone, err := g.oneOfVariantSchema(t, variant)
		if err != nil {
			return nil, fmt.Errorf("oneOf try-each variant %d: %w", i, err)
		}

		variantClone.SetSubSchemaTypeElem()

		genType, err := g.generateDeclaredType(variantClone, variantScope)
		if err != nil {
			return nil, fmt.Errorf("oneOf try-each variant %d: %w", i, err)
		}

		nt, ok := genType.(*codegen.NamedType)
		if !ok {
			return nil, fmt.Errorf("%w (try-each variant %d, got %T)", ErrUnexpectedVariantType, i, genType)
		}

		bindings[i] = variantBinding{
			decl:      nt.Decl,
			fieldName: fieldName,
		}

		// In array mode the discriminating key-set belongs to the elements,
		// not to the array itself, so the shape check is built from the
		// element schema and applied to each decoded element in turn.
		shapeSource := variantClone
		if arrayMode {
			shapeSource = arrayElementSchema(variantClone)
		}

		shapes[i] = variantShapeFor(shapeSource)

		// The variant's own declared type does not check the array's length,
		// so its bounds go into the shape check: otherwise `[]` would match a
		// variant with minItems 1, and a one-element tuple would take any
		// number of elements. The bounds of the array holding the oneOf apply
		// too; the holder replaces the field that used to check them.
		if arrayMode {
			shapes[i].minItems = max(variantClone.MinItems, t.MinItems)
			shapes[i].maxItems, shapes[i].hasMaxItems = tighterMaxItems(variantClone, t)
		}
	}

	holderStruct := &codegen.StructType{}

	for _, b := range bindings {
		holderStruct.Fields = append(holderStruct.Fields, codegen.StructField{
			Name: b.fieldName,
			Type: &codegen.PointerType{Type: &codegen.NamedType{Decl: b.decl}},
			Tags: `json:"-" yaml:"-"`,
		})
	}

	holderDecl.Type = holderStruct
	g.output.file.Package.AddDecl(holderDecl)

	if g.config.OnlyModels {
		return &codegen.NamedType{Decl: holderDecl}, nil
	}

	g.output.file.Package.AddImport("encoding/json", "")
	g.output.file.Package.AddImport("fmt", "")

	// regexp is needed only when at least one variant has patternProperties
	// in its strict shape — that's what triggers the per-pattern
	// regexp.MatchString call inside emitTryEachVariantBlock.
	for _, s := range shapes {
		if s.strict && len(s.patterns) > 0 {
			g.output.file.Package.AddImport("regexp", "")

			break
		}
	}

	hasYAML := false

	for _, f := range g.formatters {
		if f.getName() == formatYAML {
			hasYAML = true

			break
		}
	}

	if hasYAML {
		g.output.file.Package.AddImport(YAMLPackage, "yaml")
	}

	addMethod := func(suffix string, impl func(*codegen.Emitter) error) {
		g.output.file.Package.AddDecl(&codegen.Method{
			Name: holderName + "_" + suffix,
			Impl: impl,
		})
	}

	addMethod("UnmarshalJSON", emitOneOfTryEachUnmarshalJSON(holderName, bindings, shapes, arrayMode))
	addMethod("MarshalJSON", emitOneOfDiscriminatorMarshalJSON(holderName, bindings))

	if hasYAML {
		addMethod("UnmarshalYAML", emitOneOfTryEachUnmarshalYAML(holderName, bindings, shapes, arrayMode))
		addMethod("MarshalYAML", emitOneOfDiscriminatorMarshalYAML(holderName, bindings))
	}

	return &codegen.NamedType{Decl: holderDecl}, nil
}

// tryEachFormat parameterises emitOneOfTryEachBody on the format-specific
// pieces (signature, raw decode call, per-variant decode call). JSON and
// YAML bodies are otherwise structurally identical — sharing this helper
// keeps them in sync and avoids the dupl lint flag.
type tryEachFormat struct {
	methodSig   string // e.g. `func (j *T) UnmarshalJSON(value []byte) error {`
	rawDecode   string // e.g. `json.Unmarshal(value, &raw)`
	variantCall string // e.g. `json.Unmarshal(value, &v)`
	comment     []string
}

// emitOneOfTryEachUnmarshalJSON returns the codegen.Emitter callback that
// writes the holder type's UnmarshalJSON. The body is generated by the
// shared emitOneOfTryEachBody helper with format-specific snippets supplied
// via tryEachFormat (decode call, doc comment).
func emitOneOfTryEachUnmarshalJSON(
	typeName string,
	bindings []variantBinding,
	shapes []variantShape,
	arrayMode bool,
) func(*codegen.Emitter) error {
	return emitOneOfTryEachBody(typeName, bindings, shapes, arrayMode, tryEachFormat{
		methodSig:   fmt.Sprintf("func (j *%s) UnmarshalJSON(value []byte) error {", typeName),
		rawDecode:   "json.Unmarshal(value, &rawTarget)",
		variantCall: "json.Unmarshal(value, &v)",
		comment: []string{
			"UnmarshalJSON implements json.Unmarshaler. With no natural",
			"discriminator we try each variant in turn after a per-variant shape check;",
			"success requires exactly one variant to unmarshal without error.",
		},
	})
}

// emitOneOfTryEachUnmarshalYAML mirrors emitOneOfTryEachUnmarshalJSON for
// YAML inputs, delegating to the same shared body with YAML-specific
// snippets (yaml.Node decode call, MarshalYAML-aware doc comment).
func emitOneOfTryEachUnmarshalYAML(
	typeName string,
	bindings []variantBinding,
	shapes []variantShape,
	arrayMode bool,
) func(*codegen.Emitter) error {
	return emitOneOfTryEachBody(typeName, bindings, shapes, arrayMode, tryEachFormat{
		methodSig:   fmt.Sprintf("func (j *%s) UnmarshalYAML(value *yaml.Node) error {", typeName),
		rawDecode:   "value.Decode(&rawTarget)",
		variantCall: "value.Decode(&v)",
		comment: []string{
			"UnmarshalYAML mirrors UnmarshalJSON: try each variant after a",
			"shape check; exactly one must unmarshal without error.",
		},
	})
}

// emitOneOfTryEachBody is the shared emitter that writes the actual
// Unmarshal{JSON,YAML} body for a try-each holder. The flow is:
//
//  1. reset the holder to zero (variant-pointer leakage prevention),
//  2. decode raw into a generic map for the per-variant shape check,
//  3. emit one shape-checked attempt block per variant via
//     emitTryEachVariantBlock (success path increments matched and assigns
//     the variant pointer; shape mismatch skips the variant entirely),
//  4. dispatch on the final matched count: 0 ⇒ joined error, 1 ⇒ success,
//     >1 ⇒ ambiguous-input error with all variant pointers reset.
//
// The format-specific bits (method signature, decode call, doc comment)
// come from tryEachFormat so a single body generator serves both JSON and
// YAML callers.
func emitOneOfTryEachBody(
	typeName string,
	bindings []variantBinding,
	shapes []variantShape,
	arrayMode bool,
	f tryEachFormat,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		for _, line := range f.comment {
			out.Commentf("%s", line)
		}

		out.Printlnf("%s", f.methodSig)
		out.Indent(1)
		out.Commentf("Reset to zero value so reusing the same holder across multiple")
		out.Commentf("Unmarshal calls doesn't leave a previous winner set alongside the")
		out.Commentf("new one (which would violate the one-variant-set invariant and")
		out.Commentf("break the corresponding Marshal).")
		out.Printlnf("*j = %s{}", typeName)

		// Array variants dispatch on their ELEMENT key-sets, so decode a
		// slice of element maps; the per-variant shape check then walks it.
		// The object form keeps a single top-level map.
		if arrayMode {
			out.Printlnf("var rawElems []map[string]interface{}")
		} else {
			out.Printlnf("var raw map[string]interface{}")
		}

		rawVar := "raw"
		if arrayMode {
			rawVar = "rawElems"
		}

		out.Printlnf("if err := %s; err != nil {", strings.ReplaceAll(f.rawDecode, "rawTarget", rawVar))
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("unmarshal %s: %%w", err)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("matched := 0")
		out.Printlnf("var lastErr error")

		for i, b := range bindings {
			out.Newline()
			out.Commentf("Variant %d: %s", i, b.decl.Name)
			emitTryEachVariantBlock(out, b, shapes[i], f.variantCall, arrayMode)
		}

		out.Newline()
		out.Printlnf("if matched == 0 {")
		out.Indent(1)
		out.Printlnf("if lastErr != nil {")
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("%s: no oneOf variant matched: %%w", lastErr)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf(`return fmt.Errorf("%s: no oneOf variant matched")`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if matched > 1 {")
		out.Indent(1)

		for _, b := range bindings {
			out.Printlnf("j.%s = nil", b.fieldName)
		}

		out.Printlnf(`return fmt.Errorf("%s: ambiguous input — %%d oneOf variants matched", matched)`, typeName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

// emitTryEachVariantBlock writes the per-variant attempt block: a brace
// scope containing the shape pre-check (declared/pattern key sweep against
// the raw map, gated by the variant's additionalProperties flavor), the
// format-specific decode call, and the success-path assignment of the
// variant pointer plus matched++ side effect. Variants that fail either
// the shape check or the decode are silently skipped — the surrounding
// body's matched-count check decides the final outcome.
func emitTryEachVariantBlock(
	out *codegen.Emitter,
	b variantBinding,
	s variantShape,
	variantCall string,
	arrayMode bool,
) {
	out.Printlnf("{")
	out.Indent(1)
	out.Printlnf("shapeOK := true")

	// The checks below are written against a map variable named `raw`. In
	// array mode every element must satisfy the variant's element shape, so
	// bind `raw` to each element in turn and reuse the same emission
	// verbatim. The array's length is checked first: an empty array satisfies
	// the element checks of every variant, so only minItems tells them apart.
	// An element shape with neither required keys nor a strict key set checks
	// nothing, and an empty loop would not compile: its `raw` goes unused.
	elemLoop := arrayMode && (len(s.required) > 0 || s.strict)

	if arrayMode {
		if s.minItems > 0 {
			out.Printlnf("if len(rawElems) < %d { shapeOK = false }", s.minItems)
		}

		if s.hasMaxItems {
			out.Printlnf("if len(rawElems) > %d { shapeOK = false }", s.maxItems)
		}
	}

	if elemLoop {
		out.Printlnf("for _, raw := range rawElems {")
		out.Indent(1)
	}

	// Required-key check runs regardless of strict mode: a variant whose
	// required field is missing from the input cannot match, even when
	// extras are otherwise allowed.
	for _, req := range s.required {
		out.Printlnf(`if _, ok := raw[%q]; !ok { shapeOK = false }`, req)
	}

	if s.strict {
		// Emit a switch-on-key with all known properties as cases; the
		// default arm checks each patternProperties regex (if any) and
		// only flips shapeOK to false when no pattern matches either.
		out.Printlnf("for k := range raw {")
		out.Indent(1)
		out.Printlnf("switch k {")

		if len(s.knownProperties) > 0 {
			quoted := make([]string, len(s.knownProperties))
			for i, p := range s.knownProperties {
				quoted[i] = fmt.Sprintf("%q", p)
			}

			out.Printlnf("case %s:", joinCommas(quoted))
		}

		out.Printlnf("default:")
		out.Indent(1)

		if len(s.patterns) > 0 {
			out.Printlnf("matchedPattern := false")

			for _, pat := range s.patterns {
				out.Printlnf("if m, _ := regexp.MatchString(%q, k); m { matchedPattern = true }", pat)
			}

			out.Printlnf("if !matchedPattern { shapeOK = false }")
		} else {
			out.Printlnf("shapeOK = false")
		}

		out.Indent(-1)
		out.Printlnf("}")
		out.Indent(-1)
		out.Printlnf("}")
	}

	if elemLoop {
		out.Indent(-1)
		out.Printlnf("}")
	}

	out.Printlnf("if shapeOK {")
	out.Indent(1)
	out.Printlnf("var v %s", b.decl.Name)
	out.Printlnf("if err := %s; err == nil {", variantCall)
	out.Indent(1)
	out.Printlnf("j.%s = &v", b.fieldName)
	out.Printlnf("matched++")
	out.Indent(-1)
	out.Printlnf("} else {")
	out.Indent(1)
	out.Printlnf("lastErr = err")
	out.Indent(-1)
	out.Printlnf("}")
	out.Indent(-1)
	out.Printlnf("}")
	out.Indent(-1)
	out.Printlnf("}")
}

// joinCommas joins already-quoted Go literals with ", ". Used to render
// the shape-check key list as a Go-source switch-statement body, where
// each element is the output of strconv.Quote (i.e. already includes its
// own surrounding quotes). Avoiding strings.Join lets us keep the inputs
// pre-quoted at their construction sites, so callers don't have to double
// up on escaping.
func joinCommas(quoted []string) string {
	var b strings.Builder

	for i, q := range quoted {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(q)
	}

	return b.String()
}
