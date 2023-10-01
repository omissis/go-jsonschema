package generator

import (
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

const (
	formatJSON = "json"
)

type jsonFormatter struct{}

func (jf *jsonFormatter) generate(declType codegen.TypeDecl, validators []validator) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatJSON), formatJSON)
		out.Printlnf("func (j *%s) Unmarshal%s(b []byte) error {", declType.Name, strings.ToUpper(formatJSON))
		out.Indent(1)
		out.Printlnf("var %s map[string]interface{}", varNameRawMap)
		out.Printlnf("if err := %s.Unmarshal(b, &%s); err != nil { return err }",
			formatJSON, varNameRawMap)

		for _, v := range validators {
			if v.desc().beforeJSONUnmarshal {
				v.generate(out)
			}
		}

		out.Printlnf("type Plain %s", declType.Name)
		out.Printlnf("var %s Plain", varNamePlainStruct)
		out.Printlnf("if err := %s.Unmarshal(b, &%s); err != nil { return err }",
			formatJSON, varNamePlainStruct)

		for _, v := range validators {
			if !v.desc().beforeJSONUnmarshal {
				v.generate(out)
			}
		}

		out.Printlnf("*j = %s(%s)", declType.Name, varNamePlainStruct)
		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (jf *jsonFormatter) enumMarshal(declType codegen.TypeDecl) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		out.Commentf("Marshal%s implements %s.Marshaler.", strings.ToUpper(formatJSON), formatJSON)
		out.Printlnf("func (j *%s) Marshal%s() ([]byte, error) {", declType.Name, strings.ToUpper(formatJSON))
		out.Indent(1)
		out.Printlnf("return %s.Marshal(j.Value)", formatJSON)
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (jf *jsonFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	enumType codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		out.Comment("UnmarshalJSON implements json.Unmarshaler.")
		out.Printlnf("func (j *%s) UnmarshalJSON(b []byte) error {", declType.Name)
		out.Indent(1)
		out.Printf("var v ")
		enumType.Generate(out)
		out.Newline()

		varName := "v"
		if wrapInStruct {
			varName += ".Value"
		}

		out.Printlnf("if err := json.Unmarshal(b, &%s); err != nil { return err }", varName)
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
	}
}

func (jf *jsonFormatter) addImport(out *codegen.File) {
	out.Package.AddImport("encoding/json", "")
}
