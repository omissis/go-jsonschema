package generator

import (
	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type mapstructureFormatter struct{}

func (yf *mapstructureFormatter) generate(
	output *output,
	declType codegen.TypeDecl,
	validators []validator,
) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		out.Commentf("SetDefaults sets the fields of %s to their defaults.", declType.Name)
		out.Commentf("Fields which do not have a default value are left untouched.")
		out.Printlnf("func (%s *%s) SetDefaults() {", varNameStructPtr, declType.Name)
		out.Indent(1)

		for _, v := range validators {
			v.generateSetDefaults(out)
		}
		//TODO: Also generate a Validate() function?
		// It could check whether the properties comply with the schema.
		// E.g. whether a number is within a certain range.

		out.Indent(-1)
		out.Printlnf("}")
	}
}

// Noop.
func (yf *mapstructureFormatter) enumMarshal(declType codegen.TypeDecl) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
	}
}

// Noop.
func (yf *mapstructureFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	enumType codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
	}
}

// Noop.
func (yf *mapstructureFormatter) addImport(out *codegen.File) {}
