package generator

import (
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

const (
	formatYAML  = "yaml"
	YAMLPackage = "gopkg.in/yaml.v3"
)

type yamlFormatter struct{}

func (yf *yamlFormatter) generate(declType codegen.TypeDecl, validators []validator) func(*codegen.Emitter) {
	var beforeValidators []validator

	var afterValidators []validator

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

	return func(out *codegen.Emitter) {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatYAML), formatYAML)
		out.Printlnf("func (j *%s) Unmarshal%s(value *yaml.Node) error {", declType.Name,
			strings.ToUpper(formatYAML))
		out.Indent(1)

		if forceBefore || len(beforeValidators) != 0 {
			out.Printlnf("var %s map[string]interface{}", varNameRawMap)
			out.Printlnf("if err := value.Decode(&%s); err != nil { return err }", varNameRawMap)
		}

		for _, v := range beforeValidators {
			v.generate(out)
		}

		out.Printlnf("type Plain %s", declType.Name)
		out.Printlnf("var %s Plain", varNamePlainStruct)
		out.Printlnf("if err := value.Decode(&%s); err != nil { return err }", varNamePlainStruct)

		for _, v := range afterValidators {
			v.generate(out)
		}

		if structType, ok := declType.Type.(*codegen.StructType); ok {
			for _, f := range structType.Fields {
				if f.Name == "AdditionalProperties" {
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
	}
}

func (yf *yamlFormatter) enumMarshal(declType codegen.TypeDecl) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		out.Commentf("Marshal%s implements %s.Marshal.", strings.ToUpper(formatYAML), formatYAML)
		out.Printlnf("func (j *%s) Marshal%s() (interface{}, error) {", declType.Name,
			strings.ToUpper(formatYAML))
		out.Indent(1)
		out.Printlnf("return %s.Marshal(j.Value)", formatYAML)
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (yf *yamlFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	enumType codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatYAML), formatYAML)
		out.Printlnf("func (j *%s) Unmarshal%s(value *yaml.Node) error {", declType.Name,
			strings.ToUpper(formatYAML))
		out.Indent(1)
		out.Printf("var v ")
		enumType.Generate(out)
		out.Newline()

		varName := "v"
		if wrapInStruct {
			varName += ".Value"
		}

		out.Printlnf("if err := value.Decode(&%s); err != nil { return err }", varName)
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

func (yf *yamlFormatter) addImport(out *codegen.File) {
	out.Package.AddImport(YAMLPackage, "yaml")
}
