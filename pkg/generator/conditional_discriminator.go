package generator

import (
	"fmt"
	"slices"

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
// discriminator's expected value with the `then` and (optional) `else`
// subschemas to apply when matched / not matched.
type conditionalBranch struct {
	constValue string
	thenSchema *schemas.Type
	elseSchema *schemas.Type
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

// emitBranch writes one if/else block: when discStr matches the branch's
// const, run thenSchema's required checks; when it doesn't and an else is
// present, run elseSchema's required checks.
func (v *conditionalDiscriminatorValidator) emitBranch(out *codegen.Emitter, branch conditionalBranch) {
	out.Printlnf(`if discStr == %q {`, branch.constValue)
	out.Indent(1)
	v.emitRequiredChecks(out, branch.thenSchema, fmt.Sprintf("when %s='%s'", v.discriminator, branch.constValue))
	out.Indent(-1)

	if branch.elseSchema != nil && len(branch.elseSchema.Required) > 0 {
		out.Printlnf(`} else {`)
		out.Indent(1)
		v.emitRequiredChecks(out, branch.elseSchema, fmt.Sprintf("when %s!='%s'", v.discriminator, branch.constValue))
		out.Indent(-1)
	}

	out.Printlnf(`}`)
}

// emitRequiredChecks writes a presence check per field listed in the
// subschema's `required`. The contextLabel is interpolated into the error
// to make it clear which discriminator branch triggered the requirement.
func (v *conditionalDiscriminatorValidator) emitRequiredChecks(
	out *codegen.Emitter, sub *schemas.Type, contextLabel string,
) {
	if sub == nil {
		return
	}

	for _, req := range sub.Required {
		out.Printlnf(`if _, ok := %s[%q]; !ok {`, varNameRawMap, req)
		out.Indent(1)
		out.Printlnf(
			`return fmt.Errorf("field %s in %s (%s): required")`,
			req, v.declName, contextLabel,
		)
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
		key, constStr, ok := singleConstIfClause(elem)
		if !ok {
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
			constValue: constStr,
			thenSchema: elem.Then,
			elseSchema: elem.Else,
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

// singleConstIfClause matches an allOf element of shape
// `{if: {properties: {<K>: {const: <string V>}}}, then: <subschema>, else?: <subschema>}`.
// Returns (K, V, true) on match. The `then` is required to be present; the
// `else` is optional and not validated here (we rely on the presence check
// at emit time).
func singleConstIfClause(elem *schemas.Type) (string, string, bool) {
	if elem == nil || elem.If == nil || elem.Then == nil {
		return "", "", false
	}

	if len(elem.If.Properties) != 1 {
		return "", "", false
	}

	var (
		key string
		sub *schemas.Type
	)

	for k, v := range elem.If.Properties {
		key, sub = k, v
	}

	if sub == nil || sub.Const == nil {
		return "", "", false
	}

	constStr, ok := sub.Const.(string)
	if !ok {
		return "", "", false
	}

	return key, constStr, true
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

	for _, b := range branches {
		if _, ok := enumValues[b.constValue]; !ok {
			return fmt.Sprintf(
				"branch const %q is not a member of the %q enum",
				b.constValue, discriminator,
			)
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
