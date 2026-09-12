package generator

import (
	"fmt"
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
	return generateUnmarshalBody(unmarshalContext{
		formatName:   formatJSON,
		methodSuffix: strings.ToUpper(formatJSON),
		paramDecl:    "value []byte",
		decodeCall: func(target string) string {
			return fmt.Sprintf("%s.Unmarshal(value, &%s)", formatJSON, target)
		},
	}, output, declType, validators)
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

	if hasAdditionalPropertiesField(declType) {
		out.Package.AddImport("reflect", "")
		out.Package.AddImport("strings", "")
		out.Package.AddImport("github.com/go-viper/mapstructure/v2", "")
	}
}

func (jf *jsonFormatter) getName() string {
	return formatJSON
}
