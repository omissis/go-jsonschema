package generator

import (
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type verifyFormatter struct{}

func (v verifyFormatter) addImport(_ *codegen.File) {}

func (v verifyFormatter) generate(declType codegen.TypeDecl, validators []validator) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		var prefix string
		switch declType.Type.(type) {
		// No need to dereference the struct just to verify it.
		case *codegen.StructType:
			prefix = "*"

		default:
			prefix = ""
		}

		out.Comment("Verify checks all fields on the struct match the schema.")
		out.Printlnf("func (%s %s%s) Verify() error {", varNamePlainStruct, prefix, declType.Name)
		out.Indent(1)

		for _, va := range validators {
			desc := va.desc()
			if desc.beforeJSONUnmarshal || desc.requiresRawAfter || !desc.hasError {
				continue
			}

			va.generate(out)
		}

		if stct, ok := declType.Type.(*codegen.StructType); ok {
			for _, field := range stct.Fields {
				name := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				if verifyEmit := v.verifyType(field.Type, name); verifyEmit != nil {
					out.Printlnf("%s := %s", name, getPlainName(field.Name))
					verifyEmit(out)
				}
			}
		}

		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (v verifyFormatter) verifyType(tpe codegen.Type, access string) func(*codegen.Emitter) {
	// For some types, pointers are sometimes used and sometime not.
	switch utpe := tpe.(type) {
	case *codegen.ArrayType:
		return v.verifyArray(*utpe, access)

	case codegen.ArrayType:
		return v.verifyArray(utpe, access)

	case codegen.CustomNameType, *codegen.CustomNameType, codegen.NamedType, *codegen.NamedType, *codegen.StructType:
		return func(out *codegen.Emitter) {
			out.Printlnf("if err := %s.Verify(); err != nil {", access)
			out.Indent(1)
			out.Printlnf("return err")
			out.Indent(-1)
			out.Printlnf("}")
		}

	case *codegen.MapType:
		return v.verifyMap(*utpe, access)

	case codegen.MapType:
		return v.verifyMap(utpe, access)

	case *codegen.PointerType:
		return v.verifyPointer(*utpe, access)

	case codegen.PointerType:
		return v.verifyPointer(utpe, access)

	default:
		return nil
	}
}

func (v verifyFormatter) enumMarshal(_ codegen.TypeDecl) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {}
}

func (v verifyFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	_ codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) {
	return func(out *codegen.Emitter) {
		varName := enumVarName(wrapInStruct)

		out.Comment("Verify checks all fields on the struct match the schema.")
		out.Printlnf("func (%s %s) Verify() error {", varNamePlainStruct, declType.Name)
		out.Indent(1)
		out.Printlnf("for _, expected := range %s {", valueConstant.Name)
		out.Indent(1)
		out.Printlnf("if reflect.DeepEqual(%s, expected) { return nil }", varName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf(`return fmt.Errorf("invalid value (expected one of %%#v): %%#v", %s, %s)`,
			valueConstant.Name, varName)
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (v verifyFormatter) verifyArray(tpe codegen.ArrayType, access string) func(*codegen.Emitter) {
	aaccess := "a" + access

	verifyFn := v.verifyType(tpe.Type, aaccess)
	if verifyFn == nil {
		return nil
	}

	return func(out *codegen.Emitter) {
		out.Printlnf("for _, %s := range %s {", aaccess, access)
		out.Indent(1)
		verifyFn(out)
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (v verifyFormatter) verifyMap(tpe codegen.MapType, access string) func(*codegen.Emitter) {
	keyAccess := "k" + access
	valueAccess := "v" + access
	verifyKeyFn := v.verifyType(tpe.KeyType, keyAccess)
	verifyValueFn := v.verifyType(tpe.ValueType, valueAccess)

	if verifyKeyFn == nil && verifyValueFn == nil {
		return nil
	}

	if verifyKeyFn == nil {
		keyAccess = "_"
	}

	if verifyValueFn == nil {
		valueAccess = "_"
	}

	return func(out *codegen.Emitter) {
		out.Printlnf("for %s, %s := range %s {", keyAccess, valueAccess, access)
		out.Indent(1)

		if verifyKeyFn != nil {
			verifyKeyFn(out)
		}

		if verifyValueFn != nil {
			verifyValueFn(out)
		}

		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (v verifyFormatter) verifyPointer(tpe codegen.PointerType, access string) func(*codegen.Emitter) {
	var prefix string
	switch tpe.Type.(type) {
	// Access the verify and fields without copying it.
	case codegen.CustomNameType, *codegen.CustomNameType, codegen.NamedType, *codegen.NamedType:
		prefix = ""

	default:
		prefix = "*"
	}

	paccess := "p" + access

	verifyFn := v.verifyType(tpe.Type, paccess)
	if verifyFn == nil {
		return nil
	}

	return func(out *codegen.Emitter) {
		out.Printlnf("if %s != nil {", access)
		out.Printlnf("%s := %s%s", paccess, prefix, access)
		out.Indent(1)
		verifyFn(out)
		out.Indent(-1)
		out.Printlnf("}")
	}
}
