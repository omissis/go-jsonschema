package generator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// conditionalDiscriminatorValidator emits per-discriminator runtime checks
// for object schemas that use `allOf` of `if/then[/else]` clauses keyed on a
// single discriminator property's `const` value. The struct itself is
// generated as a regular object (the allOf is treated as commentary for
// type-shape purposes); only the conditional-required-field checks are
// expressed at runtime.
//
// This is the third discriminator pattern the generator supports:
//   - oneOf + const → discriminator-holder (Phase 5)
//   - oneOf without natural discriminator → try-each (Phase 6)
//   - allOf of if/then/else on a const → single struct + per-branch
//     runtime validation (this file)
//
// Schema authors choose the allOf+if/then representation when they want a
// single Go struct shape, not a holder of variants. Respecting that intent
// is why this lives as a validator rather than emitting a holder type.
type conditionalDiscriminatorValidator struct {
	declName      string
	discriminator string
	branches      []conditionalBranch
}

// conditionalBranch is one element of the parent's `allOf`. It pairs the
// discriminator's expected value(s) with the `then` and (optional) `else`
// subschemas to apply when matched / not matched. matchValues is OR-ed:
// the branch fires if discStr matches ANY entry. The single-`const` form
// produces a one-element slice; the future enum form produces N entries
// in declaration order.
type conditionalBranch struct {
	matchValues []string
	thenSchema  *schemas.Type
	elseSchema  *schemas.Type
}

// desc satisfies the validator interface. Runs against the raw decoded map
// before the typed Plain decode (`beforeJSONUnmarshal: true`) so it can read
// the discriminator value off the wire and dispatch to the matching
// per-variant required-field check. Sets `hasError: true` because the
// emitted body returns an error when a required field is missing for the
// selected variant.
func (v *conditionalDiscriminatorValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: true,
	}
}

// generate emits the per-variant required-field checks. The emitted body
// peeks the discriminator key in the raw map, asserts it is a string, then
// runs each variant's required-field check inside a `if discStr == "<V>"`
// block. Returns a typed error when the discriminator is non-string or when
// a required field is missing for the matched variant. The `format` argument
// is unused — the body is identical for JSON and YAML because the raw map
// has already been decoded by the unmarshal-body helper.
func (v *conditionalDiscriminatorValidator) generate(out *codegen.Emitter, _ string) error {
	out.Printlnf(`rawDisc, hasDisc := %s[%q]`, varNameRawMap, v.discriminator)
	out.Printlnf(`if hasDisc {`)
	out.Indent(1)
	out.Printlnf(`discStr, isStr := rawDisc.(string)`)
	out.Printlnf(`if !isStr {`)
	out.Indent(1)
	out.Printlnf(
		`return fmt.Errorf("field %s in %s: must be a string for discriminator dispatch")`,
		v.discriminator, v.declName,
	)
	out.Indent(-1)
	out.Printlnf(`}`)

	for _, branch := range v.branches {
		v.emitBranch(out, branch)
	}

	out.Indent(-1)
	out.Printlnf(`}`)

	return nil
}

// emitBranch writes one if/else block: when discStr matches any of the
// branch's match values, run thenSchema's required checks; when it doesn't
// and an else is present, run elseSchema's required checks.
//
// Single-value branches (canonical const form) emit the same byte-for-byte
// shape as before — `if discStr == "X" { … }` plus a static `(when K='X')`
// error label. Multi-value branches (enum form) emit a `||`-chained
// predicate and a runtime `(when K='%s')` template that interpolates
// discStr at error time, so the user's error names the specific value
// that triggered the branch rather than the whole match set.
func (v *conditionalDiscriminatorValidator) emitBranch(out *codegen.Emitter, branch conditionalBranch) {
	out.Printlnf(`if %s {`, matchPredicate(branch.matchValues))
	out.Indent(1)

	thenLabel, thenRuntime := branchContextLabel(v.discriminator, branch.matchValues, false)
	v.emitRequiredChecks(out, branch.thenSchema, thenLabel, thenRuntime)
	out.Indent(-1)

	if branch.elseSchema != nil && len(branch.elseSchema.Required) > 0 {
		out.Printlnf(`} else {`)
		out.Indent(1)

		elseLabel, elseRuntime := branchContextLabel(v.discriminator, branch.matchValues, true)
		v.emitRequiredChecks(out, branch.elseSchema, elseLabel, elseRuntime)
		out.Indent(-1)
	}

	out.Printlnf(`}`)
}

// matchPredicate builds the `discStr == "X"` (or `discStr == "X" || ... ||
// discStr == "Y"`) predicate for a branch's match values. Order is
// preserved from the source schema so generated goldens are deterministic.
func matchPredicate(values []string) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("discStr == %q", v)
	}

	return strings.Join(parts, " || ")
}

// branchContextLabel returns (label, runtimeExpr) for the per-branch
// required-field error message.
//
// For single-value branches: label is the static `when K='V'` (or
// `when K!='V'` for the else side); runtimeExpr is empty so the caller
// emits the literal phrase into the generated fmt.Errorf format string —
// byte-equivalent to the pre-enum behavior.
//
// For multi-value branches: label is a `when K='%s'` (or `when K!='%s'`)
// template with one `%s` placeholder; runtimeExpr is `discStr` so the
// caller emits an extra fmt.Errorf arg at runtime. The error names the
// specific value that triggered the branch rather than the whole set.
func branchContextLabel(discriminator string, values []string, isElse bool) (string, string) {
	op := "='"
	if isElse {
		op = "!='"
	}

	if len(values) == 1 {
		return fmt.Sprintf("when %s%s%s'", discriminator, op, values[0]), ""
	}

	return fmt.Sprintf("when %s%s%%s'", discriminator, op), "discStr"
}

// emitRequiredChecks writes a presence check per field listed in the
// subschema's `required`. contextLabel is interpolated into the error.
//
// When runtimeContextExpr is empty, contextLabel is a static phrase
// (single-value branches — byte-equivalent to the pre-enum behavior).
// When non-empty, contextLabel is a Go format string with one `%s`
// placeholder and runtimeContextExpr is the Go expression supplying its
// value at runtime (multi-value branches).
func (v *conditionalDiscriminatorValidator) emitRequiredChecks(
	out *codegen.Emitter, sub *schemas.Type, contextLabel, runtimeContextExpr string,
) {
	if sub == nil {
		return
	}

	for _, req := range sub.Required {
		out.Printlnf(`if _, ok := %s[%q]; !ok {`, varNameRawMap, req)
		out.Indent(1)

		if runtimeContextExpr == "" {
			out.Printlnf(
				`return fmt.Errorf("field %s in %s (%s): required")`,
				req, v.declName, contextLabel,
			)
		} else {
			out.Printlnf(
				`return fmt.Errorf("field %s in %s (%s): required", %s)`,
				req, v.declName, contextLabel, runtimeContextExpr,
			)
		}

		out.Indent(-1)
		out.Printlnf(`}`)
	}
}

// detectConditionalDiscriminator inspects t for the canonical
// `allOf` of `if[const]/then[/else]` pattern and returns a populated
// validator when the schema qualifies. Returns false if any element of the
// pattern doesn't match — in which case the schema falls through to the
// existing routing (which, post-Phase 9, will warn at the silent-loss site).
//
// Detection is conservative: we accept only the shape we can compile
// end-to-end correctly, to avoid surprising users who relied on today's
// silent-fallback behavior.
func (g *schemaGenerator) detectConditionalDiscriminator(t *schemas.Type) (*conditionalDiscriminatorValidator, bool) {
	if len(t.AllOf) == 0 {
		return nil, false
	}

	if len(t.Type) != 1 || t.Type[0] != schemas.TypeNameObject {
		return nil, false
	}

	if len(t.Properties) == 0 {
		return nil, false
	}

	var (
		discriminator string
		branches      = make([]conditionalBranch, 0, len(t.AllOf))
	)

	for i, elem := range t.AllOf {
		key, matchValues, ok, declineReason := discriminatorIfClause(elem)
		if !ok {
			if declineReason != "" {
				g.warner(fmt.Sprintf(
					"conditional-discriminator detection declined on allOf[%d]: %s",
					i, declineReason,
				))
			}

			return nil, false
		}

		// Decline detection if then/else carries composition keywords or
		// nested conditionals that this validator can't compile end-to-end.
		// Without this guard, the generator silently accepts the branch and
		// emits per-variant required-field checks while ignoring the
		// nested constraints — under-validating schemas the user expected
		// to be rejected.
		if hasUnsupportedConditionalSubschema(elem.Then) || hasUnsupportedConditionalSubschema(elem.Else) {
			return nil, false
		}

		if i == 0 {
			discriminator = key
		} else if key != discriminator {
			return nil, false
		}

		branches = append(branches, conditionalBranch{
			matchValues: matchValues,
			thenSchema:  elem.Then,
			elseSchema:  elem.Else,
		})
	}

	if reason := discriminatorInvalidReason(t, discriminator, branches); reason != "" {
		// Structural match but the discriminator declaration is wrong:
		// surface the schema bug instead of silently routing back to the
		// allOf merge path (which would produce a less-useful Go type and
		// leave the user wondering why their conditional schema didn't
		// generate the expected dispatch).
		g.warner(fmt.Sprintf(
			"conditional-discriminator pattern detected on %q-keyed allOf, but %s; falling back to plain allOf merge",
			discriminator, reason,
		))

		return nil, false
	}

	return &conditionalDiscriminatorValidator{
		discriminator: discriminator,
		branches:      branches,
	}, true
}

// discriminatorIfClause matches an allOf element of shape
// `{if: {properties: {<K>: {const|enum: <string V>|[<V1>, ...]}}},
// then: <subschema>, else?: <subschema>}`. Returns (K, values, true) on
// match where `values` is `[V]` for the const form or the enum members
// (deduplicated, declaration order preserved) for the enum form.
//
// declineReason is non-empty for "looked enum-shaped but failed
// validation" cases (non-string values, mixed types, empty enum) so the
// caller can emit a named warning. It stays empty for "not enum-shaped at
// all" decisions which fall through silently.
//
// The `if` subschema is required to be EXACTLY the single-property
// const-or-enum check — extra constraints (`required`, additional
// `properties`, nested composition keywords, type/length/pattern/etc.)
// are rejected by ifClauseHasExtraConstraints because the generated
// dispatch only checks the discriminator's value. When both `const` and
// `enum` are declared on the same property (overlapping but legal in
// JSON Schema), `const` wins as the more specific assertion.
func discriminatorIfClause(elem *schemas.Type) (string, []string, bool, string) {
	if elem == nil || elem.If == nil || elem.Then == nil {
		return "", nil, false, ""
	}

	if len(elem.If.Properties) != 1 {
		return "", nil, false, ""
	}

	if ifClauseHasExtraConstraints(elem.If) {
		return "", nil, false, ""
	}

	var (
		key string
		sub *schemas.Type
	)

	for k, v := range elem.If.Properties {
		key, sub = k, v
	}

	if sub == nil {
		return "", nil, false, ""
	}

	// Const path takes precedence over enum (more specific assertion).
	if sub.Const != nil {
		constStr, isStr := sub.Const.(string)
		if !isStr {
			return "", nil, false, fmt.Sprintf(
				"if-clause property %q has non-string const value", key,
			)
		}

		return key, []string{constStr}, true, ""
	}

	// Enum path: every member must be a string and at least one must
	// remain after dedup.
	if sub.Enum != nil {
		strs, reason := stringEnumValues(sub.Enum)
		if reason != "" {
			return "", nil, false, fmt.Sprintf("if-clause property %q %s", key, reason)
		}

		return key, strs, true, ""
	}

	// Neither const nor enum — silent decline (the schema isn't claiming
	// to be a discriminator at all).
	return "", nil, false, ""
}

// stringEnumValues filters an `enum` value list to its string members
// (deduplicated, declaration order preserved). Returns (nil, reason) when
// the list is empty, contains non-string values, or empties out after
// filtering — the conditional-discriminator dispatch is string-based, so
// any non-string would have to fall through.
func stringEnumValues(enum []any) ([]string, string) {
	if len(enum) == 0 {
		return nil, "enum is empty"
	}

	out := make([]string, 0, len(enum))
	seen := make(map[string]struct{}, len(enum))

	for i, e := range enum {
		s, ok := e.(string)
		if !ok {
			return nil, fmt.Sprintf("enum value at position %d is not a string", i)
		}

		if _, dup := seen[s]; dup {
			continue
		}

		seen[s] = struct{}{}
		out = append(out, s)
	}

	return out, ""
}

// ifClauseHasExtraConstraints reports whether an `if` subschema declares
// anything beyond the single-property check that the conditional-discriminator
// validator can dispatch on. Returns true for `required`, additional
// validation keywords, and any composition keyword — all of which would mean
// the branch's true entry condition is more than the discriminator's value,
// so generating dispatch on the value alone would under-validate.
//
// Caller is responsible for the `len(if.Properties) == 1` check; this helper
// only inspects the OTHER fields. `Type` is intentionally not checked because
// the schemas parser auto-infers `type: object` whenever `properties` is
// declared (see model.go's UnmarshalJSON), so requiring it to be empty would
// reject every legitimate canonical-shape if clause.
func ifClauseHasExtraConstraints(ifc *schemas.Type) bool {
	if ifc == nil {
		return false
	}

	return ifClauseHasExtraComposition(ifc) ||
		ifClauseHasExtraValidation(ifc) ||
		ifClauseHasExtraStructure(ifc)
}

// ifClauseHasExtraComposition reports any composition / conditional /
// reference keyword on the if subschema. These are the highest-impact
// cases: if a user nests `allOf` / `anyOf` / `oneOf` / `not` / `if` /
// `then` / `else` / `$ref` inside the if clause, the dispatch on a
// single discriminator value can't represent the actual condition.
func ifClauseHasExtraComposition(ifc *schemas.Type) bool {
	return len(ifc.AllOf) > 0 ||
		len(ifc.AnyOf) > 0 ||
		len(ifc.OneOf) > 0 ||
		ifc.Not != nil ||
		ifc.If != nil ||
		ifc.Then != nil ||
		ifc.Else != nil ||
		ifc.Ref != ""
}

// ifClauseHasExtraValidation reports any per-value validation keyword
// (string/number/array constraints, enum/const, format, pattern). These
// declare additional shape requirements beyond the discriminator value.
func ifClauseHasExtraValidation(ifc *schemas.Type) bool {
	return ifc.Const != nil || ifc.ConstIsSet ||
		ifc.Enum != nil ||
		ifc.Pattern != "" ||
		ifc.Format != "" ||
		ifc.MinLength != 0 || ifc.MaxLength != 0 ||
		ifc.MinItems != 0 || ifc.MaxItems != 0 ||
		ifc.UniqueItems ||
		ifc.Minimum != nil || ifc.Maximum != nil ||
		ifc.ExclusiveMinimum != nil || ifc.ExclusiveMaximum != nil ||
		ifc.MultipleOf != nil
}

// ifClauseHasExtraStructure reports any object-shape constraint (extra
// required fields, additional/pattern properties, dependent requirements,
// item shapes). These restrict the input shape beyond just the
// discriminator key.
func ifClauseHasExtraStructure(ifc *schemas.Type) bool {
	return len(ifc.Required) > 0 ||
		ifc.AdditionalProperties != nil ||
		len(ifc.PatternProperties) > 0 ||
		ifc.MinProperties != 0 || ifc.MaxProperties != 0 ||
		ifc.Items != nil ||
		ifc.AdditionalItems != nil ||
		len(ifc.DependentRequired) > 0 ||
		len(ifc.DependentSchemas) > 0
}

// discriminatorInvalidReason returns "" when the parent schema validly
// declares the discriminator property (string-typed, string-enum, listed
// in required, and every branch's const value is a member of the enum).
// Otherwise it returns a human-readable reason; callers should surface the
// reason via the warner so a malformed conditional-discriminator schema
// doesn't silently fall through to the plain allOf merge path.
func discriminatorInvalidReason(t *schemas.Type, discriminator string, branches []conditionalBranch) string {
	prop, ok := t.Properties[discriminator]
	if !ok {
		return fmt.Sprintf("property %q is not declared on the parent schema", discriminator)
	}

	if len(prop.Type) != 1 || prop.Type[0] != schemas.TypeNameString {
		return fmt.Sprintf("property %q must be declared as `type: string`", discriminator)
	}

	if prop.Enum == nil {
		return fmt.Sprintf("property %q must declare an `enum` listing the valid values", discriminator)
	}

	enumValues := make(map[string]struct{}, len(prop.Enum))

	for _, e := range prop.Enum {
		if s, ok := e.(string); ok {
			enumValues[s] = struct{}{}
		}
	}

	for branchIdx, b := range branches {
		for _, val := range b.matchValues {
			if _, ok := enumValues[val]; !ok {
				return fmt.Sprintf(
					"branch %d value %q is not a member of the %q enum",
					branchIdx, val, discriminator,
				)
			}
		}
	}

	if !slices.Contains(t.Required, discriminator) {
		return fmt.Sprintf("property %q must be listed in `required`", discriminator)
	}

	return ""
}

// hasUnsupportedConditionalSubschema returns true if the given then/else
// subschema carries composition keywords or nested conditionals that the
// conditional-discriminator validator can't compile end-to-end. Used as a
// detection guard so the schema falls through to the interface{} path
// (and a fidelity warning) instead of generating per-variant required-only
// code that silently ignores the nested constraints.
func hasUnsupportedConditionalSubschema(sub *schemas.Type) bool {
	if sub == nil {
		return false
	}

	return len(sub.AllOf) > 0 ||
		len(sub.AnyOf) > 0 ||
		len(sub.OneOf) > 0 ||
		sub.If != nil ||
		sub.Then != nil ||
		sub.Else != nil
}
