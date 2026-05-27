package generator

import (
	"fmt"
	"math"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

// hasAdditionalPropertiesField reports whether the generated struct will
// carry the synthetic `AdditionalProperties` catch-all field. Centralised so
// the formatters and the shared body all check the same predicate; a
// drift between the two caused real bugs in older revisions.
func hasAdditionalPropertiesField(declType *codegen.TypeDecl) bool {
	structType, ok := declType.Type.(*codegen.StructType)
	if !ok {
		return false
	}

	for _, f := range structType.Fields {
		if f.Name == additionalProperties {
			return true
		}
	}

	return false
}

// unmarshalContext captures the format-specific bits that distinguish
// UnmarshalJSON generation from UnmarshalYAML generation. Everything else
// in the body is identical across formats.
type unmarshalContext struct {
	// formatName is the lowercase name passed to validator.generate (e.g. "json", "yaml").
	formatName string
	// methodSuffix is the uppercase suffix used in the function name and doc
	// comment, e.g. "JSON" -> UnmarshalJSON / "json.Unmarshaler".
	methodSuffix string
	// paramDecl is the parameter declaration for the Unmarshal method,
	// e.g. "value []byte" or "value *yaml.Node".
	paramDecl string
	// decodeCall builds the format-specific decode expression that consumes
	// "value" into the named target. The result is wrapped in the standard
	// `if err := <expr>; err != nil { return err }` line.
	decodeCall func(targetVar string) string
}

// generateUnmarshalBody produces the body of an Unmarshal{JSON,YAML} method.
// It is the shared implementation behind jsonFormatter.generate and
// yamlFormatter.generate; the two formatters differ only by the unmarshalContext
// they pass in.
func generateUnmarshalBody(
	ctx unmarshalContext,
	output *output,
	declType *codegen.TypeDecl,
	validators []validator,
) func(*codegen.Emitter) error {
	var (
		beforeValidators []validator
		afterValidators  []validator
		forceBefore      bool
	)

	for _, v := range validators {
		d := v.desc()

		if d.beforeJSONUnmarshal {
			beforeValidators = append(beforeValidators, v)
		} else {
			afterValidators = append(afterValidators, v)
			forceBefore = forceBefore || d.requiresRawAfter
		}
	}

	return func(out *codegen.Emitter) error {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", ctx.methodSuffix, ctx.formatName)
		out.Printlnf("func (j *%s) Unmarshal%s(%s) error {", declType.Name, ctx.methodSuffix, ctx.paramDecl)
		out.Indent(1)

		hasAdditionalProperties := hasAdditionalPropertiesField(declType)

		if forceBefore || len(beforeValidators) != 0 || hasAdditionalProperties {
			out.Printlnf("var %s map[string]interface{}", varNameRawMap)
			out.Printlnf("if err := %s; err != nil {", ctx.decodeCall(varNameRawMap))
			out.Indent(1)
			out.Printlnf(`return fmt.Errorf("unmarshal raw %s: %%w", err)`, declType.Name)
			out.Indent(-1)
			out.Printlnf("}")
		}

		for _, v := range beforeValidators {
			if err := v.generate(out, ctx.formatName); err != nil {
				return fmt.Errorf("cannot generate before validators: %w", err)
			}
		}

		// Guard against `type Plain Plain` self-reference when the user's own
		// type happens to be named `Plain`, AND against collisions with any
		// other registered name. The previous form gated the rename loop on
		// `tp == declType.Name`, which silently passed `tp == "Plain"` through
		// when "Plain" was unique-by-name but identical to declType.Name.
		tp := typePlain
		for i := 0; (tp == declType.Name || !output.isUniqueTypeName(tp)) && i < math.MaxInt; i++ {
			tp = fmt.Sprintf("%s_%d", typePlain, i)
		}

		out.Printlnf("type %s %s", tp, declType.Name)
		out.Printlnf("var %s %s", varNamePlainStruct, tp)
		out.Printlnf("if err := %s; err != nil {", ctx.decodeCall(varNamePlainStruct))
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("unmarshal %s: %%w", err)`, declType.Name)
		out.Indent(-1)
		out.Printlnf("}")

		for _, v := range afterValidators {
			if err := v.generate(out, ctx.formatName); err != nil {
				return fmt.Errorf("cannot generate after validators: %w", err)
			}
		}

		if hasAdditionalProperties {
			// Strip declared properties from raw before passing it to
			// mapstructure for the catch-all map. We use only the active
			// format's wire name so that, e.g., a JSON additional-property
			// key whose name happens to match a Go identifier or an unrelated
			// tag isn't silently dropped. The synthetic AdditionalProperties
			// field carries no meaningful tag and must be skipped — otherwise
			// `delete(raw, "AdditionalProperties")` would eat that input key.
			// A field tagged `json:"-"` (or `yaml:"-"`) is also skipped so a
			// legitimate additional-property key literally named "-" is
			// preserved.
			//
			// Match case-insensitively because encoding/json and yaml.v3 both
			// accept case-variant keys for declared fields (with exact case
			// taking precedence). Without case-insensitive pruning, a payload
			// like {"foo": ..., "FOO": ...} where `foo` is declared would
			// leak the `FOO` variant into AdditionalProperties.
			tagName := formatJSON
			if ctx.formatName == formatYAML {
				tagName = formatYAML
			}

			out.Printlnf("st := reflect.TypeOf(%s{})", tp)
			out.Printlnf("for i := 0; i < st.NumField(); i++ {")
			out.Indent(1)
			out.Printlnf("f := st.Field(i)")
			out.Printlnf(`if f.Name == %q { continue }`, additionalProperties)
			out.Printlnf(`name := strings.Split(f.Tag.Get(%q), ",")[0]`, tagName)
			out.Printlnf(`if name == "-" { continue }`)
			out.Printlnf(`if name == "" { name = f.Name }`)
			out.Printlnf("for k := range %s {", varNameRawMap)
			out.Indent(1)
			out.Printlnf(`if strings.EqualFold(k, name) { delete(%s, k) }`, varNameRawMap)
			out.Indent(-1)
			out.Printlnf("}")
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf(
				"if err := mapstructure.Decode(%s, &%s.AdditionalProperties); err != nil {",
				varNameRawMap, varNamePlainStruct,
			)
			out.Indent(1)
			out.Printlnf(`return fmt.Errorf("decode additional properties for %s: %%w", err)`, declType.Name)
			out.Indent(-1)
			out.Printlnf("}")
		}

		out.Printlnf("*j = %s(%s)", declType.Name, varNamePlainStruct)
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}
