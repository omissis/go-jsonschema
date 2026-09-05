package generator

import (
	"fmt"
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

// arrayElementSchema returns the single schema describing every element of an
// array variant: the plain `items` schema, or the sole entry of a one-element
// tuple (the "tuple of exactly one" idiom). It returns nil for anything else —
// a heterogeneous tuple has no single element type, and an array with no
// `items` is unconstrained.
func arrayElementSchema(v *schemas.Type) *schemas.Type {
	if len(v.TupleItems) == 1 {
		return v.TupleItems[0]
	}

	if len(v.TupleItems) > 0 {
		return nil
	}

	return v.Items
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

	strict := v.AdditionalProperties != nil && v.AdditionalProperties.Not != nil

	return variantShape{
		knownProperties: props,
		patterns:        patterns,
		required:        required,
		strict:          strict,
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
	holderName := g.output.uniqueTypeName(scope)
	if g.config.StructNameFromTitle && t.Title != "" {
		holderName = g.caser.Identifierize(t.Title)
	}

	arrayMode := isTryEachArrayCandidate(t.OneOf)
	bindings := make([]variantBinding, len(t.OneOf))
	shapes := make([]variantShape, len(t.OneOf))

	for i, variant := range t.OneOf {
		fieldName := fmt.Sprintf("Variant%d", i)
		variantScope := scope.add(fieldName)

		// Clone before mutating: t.OneOf entries can come from a resolved
		// $ref, in which case the same *Type is referenced from elsewhere
		// in the schema graph. Setting the flag on the original would
		// contaminate the shared cache.
		variantClone := *variant
		variantClone.SetSubSchemaTypeElem()

		genType, err := g.generateDeclaredType(&variantClone, variantScope)
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
		shapeSource := &variantClone
		if arrayMode {
			shapeSource = arrayElementSchema(&variantClone)
		}

		shapes[i] = variantShapeFor(shapeSource)
	}

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
	methodSig    string // e.g. `func (j *T) UnmarshalJSON(value []byte) error {`
	rawDecode    string // e.g. `json.Unmarshal(value, &raw)`
	variantCall  string // e.g. `json.Unmarshal(value, &v)`
	commentLine1 string
	commentLine2 string
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
		methodSig:    fmt.Sprintf("func (j *%s) UnmarshalJSON(value []byte) error {", typeName),
		rawDecode:    "json.Unmarshal(value, &rawTarget)",
		variantCall:  "json.Unmarshal(value, &v)",
		commentLine1: "UnmarshalJSON implements json.Unmarshaler. With no natural",
		commentLine2: "discriminator we try each variant in turn after a per-variant shape check;",
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
		methodSig:    fmt.Sprintf("func (j *%s) UnmarshalYAML(value *yaml.Node) error {", typeName),
		rawDecode:    "value.Decode(&rawTarget)",
		variantCall:  "value.Decode(&v)",
		commentLine1: "UnmarshalYAML mirrors UnmarshalJSON: try each variant after a",
		commentLine2: "shape check; exactly one must unmarshal without error.",
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
		out.Commentf("%s", f.commentLine1)
		out.Commentf("%s", f.commentLine2)
		out.Commentf("success requires exactly one variant to unmarshal without error.")
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
	// verbatim. An empty array vacuously satisfies every variant and so is
	// reported as ambiguous, which is what `oneOf` semantics imply for it.
	if arrayMode {
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

	if arrayMode {
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
