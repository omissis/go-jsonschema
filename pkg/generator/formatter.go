package generator

import (
	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type formatter interface {
	addImport(out *codegen.File, declType codegen.TypeDecl)

	generate(output *output, declType codegen.TypeDecl, validators []validator) func(*codegen.Emitter) error
	enumMarshal(declType codegen.TypeDecl) func(*codegen.Emitter) error
	enumUnmarshal(
		declType codegen.TypeDecl,
		enumType codegen.Type,
		valueConstant *codegen.Var,
		wrapInStruct bool,
	) func(*codegen.Emitter) error
}
