package generator

import (
	"fmt"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// warnFallback emits a structured warning when a schema is about to be
// represented as `interface{}` while still declaring keywords whose presence
// implies the user expected enforcement. Returns true when a warning was
// emitted (purely for caller convenience — the helper always permits the
// caller to fall back regardless).
//
// The warning is suppressed when the schema has nothing meaningful to enforce
// (e.g. a genuinely-empty `{}` schema); this keeps signal-to-noise high.
//
// trigger is a short human-readable reason ("if/then/else not compiled",
// "allOf merge produced empty result", etc.) used to anchor the warning so
// users can find the responsible keyword.
func (g *schemaGenerator) warnFallback(t *schemas.Type, scope nameScope, trigger string) {
	dropped := collectDroppedKeywords(t)

	// collectDroppedKeywords leaves a conditional to the trigger, but a type
	// that declares nothing besides one still loses it, as does one whose
	// trigger names something else.
	if conditional, has := findConditional(t, ownConditional); has {
		dropped = append(dropped, conditional)
	}

	if len(dropped) == 0 {
		return
	}

	msg := fmt.Sprintf(
		"schema fidelity: type %s falls back to interface{} due to %s; declared but not enforced: %s",
		scope.string(), trigger, strings.Join(dropped, ", "),
	)

	if affected := affectedConfigsFor(dropped, g.config); len(affected) > 0 {
		msg += "; affected configuration: " + strings.Join(affected, ", ")
	}

	g.warner(msg)
}

// warnNotEnforced emits a structured warning when a type keeps its shape but
// drops something it declares, which the generated code then does not check;
// reason says what. warnFallback covers a type that degrades to interface{}
// as a whole.
func (g *schemaGenerator) warnNotEnforced(scope nameScope, reason string) {
	g.warner(fmt.Sprintf("schema fidelity: type %s does not enforce all it declares: %s", scope.string(), reason))
}

// warnUncompiledOneOf warns when t declares a oneOf that no oneOf strategy
// compiled while t itself became a struct: the struct holds t's own
// properties and checks none of the branches, such as branches that only list
// required, so a value matching none or several of them passes.
func (g *schemaGenerator) warnUncompiledOneOf(t *schemas.Type, theType codegen.Type, scope nameScope) {
	if len(t.OneOf) == 0 || g.config.OnlyModels {
		return
	}

	if _, isStruct := theType.(*codegen.StructType); isStruct {
		g.warnNotEnforced(scope, fmt.Sprintf(
			"oneOf of %d branches not compiled, so a value matching none or several of them passes",
			len(t.OneOf),
		))
	}
}

// warnUncompiledConditional warns when t declares an if/then/else or a not
// of its own, which no code path compiles outside the conditional
// discriminator pattern. It leaves an allOf to generateAllOfType, which
// reports the allOf's own conditional together with its branches'; a $ref to
// the schema it refers to; and a schema that the goJSONSchema extension maps to
// a Go type of the user's to that type. A schema is reported once, though it
// may pass through more than one of the generator's entry points.
func (g *schemaGenerator) warnUncompiledConditional(t *schemas.Type, scope nameScope) {
	if g.config.OnlyModels || len(t.AllOf) > 0 || t.Ref != "" {
		return
	}

	if ext := t.GoJSONSchemaExtension; ext != nil && ext.Type != nil {
		return
	}

	if _, reported := g.reportedConditionals[t]; reported {
		return
	}

	if reason := ownConditionalReason(t); reason != "" {
		g.reportedConditionals[t] = struct{}{}
		g.warnNotEnforced(scope, reason)
	}
}

// warnUncompiledVariants reports the if/then/else or not of a primitive
// oneOf's variants, which the wrapper does not compile: it checks only the
// JSON kind of each, and declares no type of theirs that would report them.
func (g *schemaGenerator) warnUncompiledVariants(t *schemas.Type, scope nameScope) {
	for _, variant := range t.OneOf {
		if _, reported := g.reportedConditionals[variant]; reported {
			continue
		}

		if reason := ownConditionalReason(variant); reason != "" {
			g.reportedConditionals[variant] = struct{}{}
			g.warnNotEnforced(scope, reason+" (in oneOf branch)")
		}
	}
}

// The conditionals ownConditional names.
const (
	conditionalIfThenElse  = "if/then/else"
	conditionalFalseSchema = "false schema"
	conditionalNot         = `"not"`
)

// ownConditional names t's own if/then/else, false schema or not, if it
// constrains a value, or returns "": a conditional whose then and else impose
// nothing is vacuous, so dropping it loses nothing. The schema `false` is
// read as a not of the empty schema, and is named as such.
func ownConditional(t *schemas.Type) string {
	switch {
	case (t.If != nil || t.Then != nil || t.Else != nil) && (imposes(t.Then) || imposes(t.Else)):
		return conditionalIfThenElse
	case isFalseSchema(t):
		return conditionalFalseSchema
	case t.Not != nil:
		return conditionalNot
	}

	return ""
}

// ownConditionalReason says why the conditional ownConditional names goes
// unchecked, or returns "" if t has none.
func ownConditionalReason(t *schemas.Type) string {
	switch ownConditional(t) {
	case conditionalIfThenElse:
		return "if/then/else not compiled"
	case conditionalFalseSchema:
		return "false schema not compiled, so any value passes"
	case conditionalNot:
		return `"not" keyword not compiled`
	}

	return ""
}

// constrainingComposition is hasUnsupportedComposition limited to what
// actually constrains a value, as ownConditionalReason decides it for t and
// each of its branches.
func constrainingComposition(t *schemas.Type) (string, bool) {
	return findConditional(t, ownConditionalReason)
}

// findConditional returns what describe says of t, or else of the first of its
// branches, at any depth, of which it says anything, followed by the branches
// that lead to it.
func findConditional(t *schemas.Type, describe func(*schemas.Type) string) (string, bool) {
	if description := describe(t); description != "" {
		return description, true
	}

	for _, branches := range []struct {
		keyword schemas.SubSchemaType
		list    []*schemas.Type
	}{
		{schemas.SubSchemaTypeAllOf, t.AllOf},
		{schemas.SubSchemaTypeAnyOf, t.AnyOf},
		{schemas.SubSchemaTypeOneOf, t.OneOf},
	} {
		for _, child := range branches.list {
			if description, has := findConditional(child, describe); has {
				return description + " (in " + string(branches.keyword) + " branch)", true
			}
		}
	}

	return "", false
}

// imposes reports whether a then or else subschema constrains a value.
func imposes(t *schemas.Type) bool {
	if t == nil {
		return false
	}

	if len(collectDroppedKeywords(t)) > 0 {
		return true
	}

	_, has := constrainingComposition(t)

	return has
}

// collectDroppedKeywords lists schema keywords whose presence indicates the
// user expects the generator to enforce something. The list returned here is
// what gets silently lost when the type degrades to interface{}. Recurses
// into allOf/anyOf/oneOf branches so constraints declared inside a branch
// are surfaced even when the parent is a thin composition wrapper.
//
// Does NOT recurse into if/then/else/not — those are conditional or negated
// schemas whose nested keywords aren't declarations on the parent type;
// their unsupported-ness is reported via hasUnsupportedComposition's
// trigger string instead. Results are deduplicated by literal keyword
// string and visited-set protected against cyclic schemas.
func collectDroppedKeywords(t *schemas.Type) []string {
	var (
		dropped []string
		seen    = make(map[string]struct{})
		visited = make(map[*schemas.Type]struct{})
	)

	add := func(kw string) {
		if _, dup := seen[kw]; dup {
			return
		}

		seen[kw] = struct{}{}

		dropped = append(dropped, kw)
	}

	var walk func(t *schemas.Type)

	walk = func(t *schemas.Type) {
		if t == nil {
			return
		}

		if _, v := visited[t]; v {
			return
		}

		visited[t] = struct{}{}

		if t.AdditionalProperties != nil && isStrictAdditionalProperties(t.AdditionalProperties) {
			add("additionalProperties: false")
		}

		if len(t.Required) > 0 {
			add("required")
		}

		if t.Format != "" {
			add("format: " + t.Format)
		}

		if t.Pattern != "" {
			add("pattern")
		}

		if t.MinLength != 0 || t.MaxLength != 0 {
			add("string length constraint")
		}

		if t.Minimum != nil || t.Maximum != nil ||
			t.ExclusiveMinimum != nil || t.ExclusiveMaximum != nil ||
			t.MultipleOf != nil {
			add("numeric constraint")
		}

		if t.MinItems != 0 || t.MaxItems != 0 || t.UniqueItems {
			add("array constraint")
		}

		if len(t.Properties) > 0 {
			add(fmt.Sprintf("%d declared property(ies)", len(t.Properties)))
		}

		if t.Enum != nil {
			add("enum")
		}

		if t.ConstIsSet {
			add("const")
		}

		// A branch that refers to another schema carries that schema's
		// constraints, which go unchecked with the branch.
		if t.Ref != "" {
			add("$ref " + t.Ref)
		}

		for _, child := range t.AllOf {
			walk(child)
		}

		for _, child := range t.AnyOf {
			walk(child)
		}

		for _, child := range t.OneOf {
			walk(child)
		}
	}

	walk(t)

	return dropped
}

// affectedConfigsFor maps the dropped keywords to the configuration knobs
// they would have triggered, so the user knows which `Config.*` settings
// became no-ops on this type. Always returns the always-on validators
// (required, etc.) when their keyword is dropped.
func affectedConfigsFor(dropped []string, cfg Config) []string {
	var affected []string

	for _, kw := range dropped {
		switch {
		case strings.HasPrefix(kw, "additionalProperties"):
			if cfg.StrictAdditionalProperties != StrictAdditionalPropertiesOff {
				affected = append(affected, "Config.StrictAdditionalProperties")
			}

		case strings.HasPrefix(kw, "format:"):
			if cfg.FormatValidation.Enabled {
				affected = append(affected, "Config.FormatValidation")
			}

		case kw == "required":
			affected = append(affected, "required-field check")

		case kw == "pattern", strings.HasPrefix(kw, "string length"):
			affected = append(affected, "string validator")

		case strings.HasPrefix(kw, "numeric"):
			affected = append(affected, "numeric validator")

		case strings.HasPrefix(kw, "array"):
			affected = append(affected, "array validator")
		}
	}

	return dedupAffected(affected)
}

// isStrictAdditionalProperties reports whether the additionalProperties
// schema means literal `false` (i.e. "no extra properties allowed"). A
// non-nil typed AdditionalProperties (like `additionalProperties: {type:
// "string"}`) is not in scope — that path generates a typed-map field today.
//
// JSON Schema parses `false` into `&Type{Not: &Type{}}` (the "matches
// nothing" representation); see pkg/schemas/model.go's UnmarshalJSON path.
func isStrictAdditionalProperties(t *schemas.Type) bool {
	return t != nil && t.Not != nil && len(t.Type) == 0 &&
		len(t.Properties) == 0 && len(t.AllOf) == 0 &&
		len(t.AnyOf) == 0 && len(t.OneOf) == 0
}

// hasUnsupportedComposition reports whether t (or any of its allOf/anyOf/
// oneOf children) contains a composition keyword that the generator does
// not currently compile end-to-end. Used by warnFallback callers to enrich
// the trigger string.
func hasUnsupportedComposition(t *schemas.Type) (string, bool) {
	if t.If != nil || t.Then != nil || t.Else != nil {
		return "if/then/else not compiled", true
	}

	if t.Not != nil {
		return `"not" keyword not compiled`, true
	}

	for _, child := range t.AllOf {
		if reason, has := hasUnsupportedComposition(child); has {
			return reason + " (in allOf branch)", true
		}
	}

	for _, child := range t.AnyOf {
		if reason, has := hasUnsupportedComposition(child); has {
			return reason + " (in anyOf branch)", true
		}
	}

	for _, child := range t.OneOf {
		if reason, has := hasUnsupportedComposition(child); has {
			return reason + " (in oneOf branch)", true
		}
	}

	return "", false
}

// dedupAffected removes duplicate entries while preserving first-seen order.
// Used by affectedConfigsFor to keep the warning's "affected configuration:"
// list stable across runs and free of repeats when several dropped keywords
// map to the same Config knob.
func dedupAffected(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, s := range in {
		if _, dup := seen[s]; dup {
			continue
		}

		seen[s] = struct{}{}

		out = append(out, s)
	}

	return out
}

// compositionFallbackTrigger explains why a composition could not be compiled,
// for use as the trigger text of a fidelity warning. Returns "" when t is not a
// composition, so callers can use it as the gate.
//
// The oneOf case is the one that bites in practice: the primitive wrapper
// dispatches on JSON token kind and has nowhere to enforce per-variant
// keywords, so a single variant carrying `format` or `minimum` disqualifies the
// whole schema and it degrades to interface{}. Naming the offending variant is
// what makes that actionable — the alternative is a silent `interface{}` that
// looks deliberate.
func compositionFallbackTrigger(t *schemas.Type) string {
	if t == nil {
		return ""
	}

	switch {
	case len(t.OneOf) > 1:
		for i, v := range t.OneOf {
			if v != nil && variantHasValidationConstraints(v) {
				return fmt.Sprintf(
					"oneOf variant %d declares constraints the primitive wrapper cannot enforce", i,
				)
			}
		}

		return "oneOf not compiled"

	case len(t.AnyOf) > 0:
		return "anyOf not compiled"

	case len(t.AllOf) > 0:
		return "allOf not compiled"
	}

	return ""
}

// warnCompositionFallback emits a fidelity warning when a composition degrades
// to interface{}. No-op for non-compositions, and — via warnFallback — for
// schemas that declare nothing worth losing.
func (g *schemaGenerator) warnCompositionFallback(t *schemas.Type, scope nameScope) {
	if trigger := compositionFallbackTrigger(t); trigger != "" {
		g.warnFallback(t, scope, trigger)
	}
}
