package generator

import (
	"fmt"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

const (
	formatYAML  = "yaml"
	YAMLPackage = "gopkg.in/yaml.v3"
)

type yamlFormatter struct{}

func (yf *yamlFormatter) generate(
	output *output,
	declType *codegen.TypeDecl,
	validators []validator,
) func(*codegen.Emitter) error {
	return generateUnmarshalBody(unmarshalContext{
		formatName:   formatYAML,
		methodSuffix: strings.ToUpper(formatYAML),
		paramDecl:    "value *yaml.Node",
		decodeCall: func(target string) string {
			return fmt.Sprintf("value.Decode(&%s)", target)
		},
	}, output, declType, validators)
}

func (yf *yamlFormatter) enumMarshal(declType *codegen.TypeDecl) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("Marshal%s implements %s.Marshal.", strings.ToUpper(formatYAML), formatYAML)
		out.Printlnf("func (j *%s) Marshal%s() (interface{}, error) {", declType.Name, strings.ToUpper(formatYAML))
		out.Indent(1)
		out.Printlnf("return %s.Marshal(j.Value)", formatYAML)
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (yf *yamlFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	enumType codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatYAML), formatYAML)
		out.Printlnf("func (j *%s) Unmarshal%s(value *yaml.Node) error {", declType.Name, strings.ToUpper(formatYAML))
		out.Indent(1)
		out.Printf("var v ")

		if err := enumType.Generate(out); err != nil {
			return fmt.Errorf("cannot unmarshal enum content: %w", err)
		}

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
		out.Printlnf(`return fmt.Errorf("invalid value (expected one of %%#v): %%#v", %s, %s)`, valueConstant.Name, varName)
		out.Printlnf("}")
		out.Printlnf(`*j = %s(v)`, declType.Name)
		out.Printlnf(`return nil`)
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (yf *yamlFormatter) addImport(out *codegen.File, declType *codegen.TypeDecl) {
	out.Package.AddImport(YAMLPackage, "yaml")

	if hasAdditionalPropertiesField(declType) {
		out.Package.AddImport("reflect", "")
		out.Package.AddImport("strings", "")
		out.Package.AddImport("github.com/go-viper/mapstructure/v2", "")
	}
}

func (yf *yamlFormatter) getName() string {
	return formatYAML
}
