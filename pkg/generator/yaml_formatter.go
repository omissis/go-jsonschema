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
	return func(out *codegen.Emitter) {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatYAML), formatYAML)
		out.Printlnf("func (j *%s) Unmarshal%s(value *yaml.Node) error {", declType.Name,
			strings.ToUpper(formatYAML))
		out.Indent(1)
		out.Printlnf("var %s map[string]interface{}", varNameRawMap)
		out.Printlnf("if err := value.Decode(&%s); err != nil { return err }", varNameRawMap)

		for _, v := range validators {
			if v.desc().beforeJSONUnmarshal {
				v.generate(out)
			}
		}

		out.Printlnf("type Plain %s", declType.Name)
		out.Printlnf("var %s Plain", varNamePlainStruct)
		out.Printlnf("if err := value.Decode(&%s); err != nil { return err }", varNamePlainStruct)

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
