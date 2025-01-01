package generator

import (
	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type formatter interface {
	addImport(out *codegen.File)

	generate(output *output, declType codegen.TypeDecl, validators []validator) func(*codegen.Emitter)
	enumMarshal(declType codegen.TypeDecl) func(*codegen.Emitter)
	enumUnmarshal(
		declType codegen.TypeDecl,
		enumType codegen.Type,
		valueConstant *codegen.Var,
		wrapInStruct bool,
	) func(*codegen.Emitter)
}
