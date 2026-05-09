package generator

import (
	"fmt"
	"strings"

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
func (g *schemaGenerator) warnFallback(t *schemas.Type, scope nameScope, trigger string) bool {
	dropped := collectDroppedKeywords(t)
	if len(dropped) == 0 {
		return false
	}

	msg := fmt.Sprintf(
		"schema fidelity: type %s falls back to interface{} due to %s; declared but not enforced: %s",
		scope.string(), trigger, strings.Join(dropped, ", "),
	)

	if affected := affectedConfigsFor(dropped, g.config); len(affected) > 0 {
		msg += "; affected configuration: " + strings.Join(affected, ", ")
	}

	g.warner(msg)

	return true
}

// collectDroppedKeywords lists schema keywords whose presence indicates the
// user expects the generator to enforce something. The list returned here is
// what gets silently lost when the type degrades to interface{}.
func collectDroppedKeywords(t *schemas.Type) []string {
	var dropped []string

	if t.AdditionalProperties != nil && isStrictAdditionalProperties(t.AdditionalProperties) {
		dropped = append(dropped, "additionalProperties: false")
	}

	if len(t.Required) > 0 {
		dropped = append(dropped, "required")
	}

	if t.Format != "" {
		dropped = append(dropped, "format: "+t.Format)
	}

	if t.Pattern != "" {
		dropped = append(dropped, "pattern")
	}

	if t.MinLength != 0 || t.MaxLength != 0 {
		dropped = append(dropped, "string length constraint")
	}

	if t.Minimum != nil || t.Maximum != nil ||
		t.ExclusiveMinimum != nil || t.ExclusiveMaximum != nil ||
		t.MultipleOf != nil {
		dropped = append(dropped, "numeric constraint")
	}

	if t.MinItems != 0 || t.MaxItems != 0 || t.UniqueItems {
		dropped = append(dropped, "array constraint")
	}

	if len(t.Properties) > 0 {
		dropped = append(dropped, fmt.Sprintf("%d declared property(ies)", len(t.Properties)))
	}

	if t.Enum != nil {
		dropped = append(dropped, "enum")
	}

	if t.Const != nil {
		dropped = append(dropped, "const")
	}

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
