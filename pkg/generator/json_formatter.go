package generator

import (
	"fmt"
	"math"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

const (
	formatJSON = "json"
)

var ErrCannotUnmarshalEnum = fmt.Errorf("cannot unmarshal enum")

type jsonFormatter struct{}

func (jf *jsonFormatter) generate(
	output *output,
	declType *codegen.TypeDecl,
	validators []validator,
) func(*codegen.Emitter) error {
	var (
		beforeValidators []validator
		afterValidators  []validator
	)

	forceBefore := false

	for _, v := range validators {
		desc := v.desc()
		if desc.beforeJSONUnmarshal {
			beforeValidators = append(beforeValidators, v)
		} else {
			afterValidators = append(afterValidators, v)
			forceBefore = forceBefore || desc.requiresRawAfter
		}
	}

	return func(out *codegen.Emitter) error {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatJSON), formatJSON)
		out.Printlnf("func (j *%s) Unmarshal%s(value []byte) error {", declType.Name, strings.ToUpper(formatJSON))
		out.Indent(1)

		if forceBefore || len(beforeValidators) != 0 {
			out.Printlnf("var %s map[string]interface{}", varNameRawMap)
			out.Printlnf("if err := %s.Unmarshal(value, &%s); err != nil { return err }", formatJSON, varNameRawMap)
		}

		for _, v := range beforeValidators {
			if err := v.generate(out, "json"); err != nil {
				return fmt.Errorf("cannot generate before validators: %w", err)
			}
		}

		tp := typePlain

		if tp == declType.Name {
			for i := 0; !output.isUniqueTypeName(tp) && i < math.MaxInt; i++ {
				tp = fmt.Sprintf("%s_%d", typePlain, i)
			}
		}

		out.Printlnf("type %s %s", tp, declType.Name)
		out.Printlnf("var %s %s", varNamePlainStruct, tp)
		out.Printlnf("if err := %s.Unmarshal(value, &%s); err != nil { return err }",
			formatJSON, varNamePlainStruct)

		for _, v := range afterValidators {
			if err := v.generate(out, "json"); err != nil {
				return fmt.Errorf("cannot generate after validators: %w", err)
			}
		}

		if structType, ok := declType.Type.(*codegen.StructType); ok {
			for _, f := range structType.Fields {
				if f.Name == additionalProperties {
					out.Printlnf("st := reflect.TypeOf(Plain{})")
					out.Printlnf("for i := range st.NumField() {")
					out.Indent(1)
					out.Printlnf("delete(raw, st.Field(i).Name)")
					out.Printlnf("delete(raw, strings.Split(st.Field(i).Tag.Get(\"json\"), \",\")[0])")
					out.Indent(-1)
					out.Printlnf("}")
					out.Printlnf("if err := mapstructure.Decode(raw, &plain.AdditionalProperties); err != nil {")
					out.Indent(1)
					out.Printlnf("return err")
					out.Indent(-1)
					out.Printlnf("}")

					break
				}
			}
		}

		out.Printlnf("*j = %s(%s)", declType.Name, varNamePlainStruct)
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (jf *jsonFormatter) enumMarshal(declType *codegen.TypeDecl) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("Marshal%s implements %s.Marshaler.", strings.ToUpper(formatJSON), formatJSON)
		out.Printlnf("func (j *%s) Marshal%s() ([]byte, error) {", declType.Name, strings.ToUpper(formatJSON))
		out.Indent(1)
		out.Printlnf("return %s.Marshal(j.Value)", formatJSON)
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (jf *jsonFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	enumType codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Comment("UnmarshalJSON implements json.Unmarshaler.")
		out.Printlnf("func (j *%s) UnmarshalJSON(value []byte) error {", declType.Name)
		out.Indent(1)
		out.Printf("var v ")

		if err := enumType.Generate(out); err != nil {
			return fmt.Errorf("%w: %w", ErrCannotUnmarshalEnum, err)
		}

		out.Newline()

		varName := "v"
		if wrapInStruct {
			varName += ".Value"
		}

		out.Printlnf("if err := json.Unmarshal(value, &%s); err != nil { return err }", varName)
		out.Printlnf("var ok bool")
		out.Printlnf("for _, expected := range %s {", valueConstant.Name)
		out.Printlnf("if reflect.DeepEqual(%s, expected) { ok = true; break }", varName)
		out.Printlnf("}")
		out.Printlnf("if !ok {")
		out.Printlnf(`return fmt.Errorf("invalid value (expected one of %%#v): %%#v", %s, %s)`,
			valueConstant.Name, varName)
		out.Printlnf("}")
		out.Printlnf(`*j = %s(v)`, declType.Name)
		out.Printlnf(`return nil`)
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (jf *jsonFormatter) addImport(out *codegen.File, declType *codegen.TypeDecl) {
	out.Package.AddImport("encoding/json", "")

	if structType, ok := declType.Type.(*codegen.StructType); ok {
		for _, f := range structType.Fields {
			if f.Name == additionalProperties {
				out.Package.AddImport("reflect", "")
				out.Package.AddImport("strings", "")
				out.Package.AddImport("github.com/go-viper/mapstructure/v2", "")

				return
			}
		}
	}
}

func (yf *jsonFormatter) getName() string {
	return "json"
}
