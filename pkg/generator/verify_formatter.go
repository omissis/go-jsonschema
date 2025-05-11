package generator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

var (
	ErrCannotGeneratePointerVerificationCode = errors.New("cannot generate pointer verification code")
	ErrCannotGenerateMapVerificationCode     = errors.New("cannot generate map verification code")
	ErrCannotGenerateSelfVerificationCode    = errors.New("cannot generate self-verification code")
	ErrCannotGenerateArrayVerificationCode   = errors.New("cannot generate array verification code")
)

type verifyFormatter struct{}

func (v verifyFormatter) addImport(_ *codegen.File, _ codegen.TypeDecl) {}

func (v verifyFormatter) generate(
	_ *output,
	declType codegen.TypeDecl,
	validators []validator,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
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

			if err := va.generate(out, ""); err != nil {
				return fmt.Errorf("%w: %w", ErrCannotGenerateSelfVerificationCode, err)
			}
		}

		if stct, ok := declType.Type.(*codegen.StructType); ok {
			for _, field := range stct.Fields {
				name := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				if verifyEmit := v.verifyType(field.Type, name); verifyEmit != nil {
					out.Printlnf("%s := %s", name, getPlainName(field.Name))

					if err := verifyEmit(out); err != nil {
						return fmt.Errorf("%w: %w", ErrCannotGenerateSelfVerificationCode, err)
					}
				}
			}
		}

		out.Printlnf("return nil")
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (v verifyFormatter) verifyType(tpe codegen.Type, access string) func(*codegen.Emitter) error {
	// For some types, pointers are sometimes used and sometime not.
	switch utpe := tpe.(type) {
	case *codegen.ArrayType:
		return v.verifyArray(*utpe, access)

	case codegen.ArrayType:
		return v.verifyArray(utpe, access)

	case codegen.CustomNameType, *codegen.CustomNameType, codegen.NamedType, *codegen.NamedType, *codegen.StructType:
		return func(out *codegen.Emitter) error {
			out.Printlnf("if err := %s.Verify(); err != nil {", access)
			out.Indent(1)
			out.Printlnf("return err")
			out.Indent(-1)
			out.Printlnf("}")

			return nil
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

func (v verifyFormatter) enumMarshal(_ codegen.TypeDecl) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		return nil
	}
}

func (v verifyFormatter) enumUnmarshal(
	declType codegen.TypeDecl,
	_ codegen.Type,
	valueConstant *codegen.Var,
	wrapInStruct bool,
) func(*codegen.Emitter) error {
	return func(out *codegen.Emitter) error {
		varName := enumVarName(wrapInStruct)

		out.Comment("Verify checks all fields on the struct match the schema.")
		out.Printlnf("func (%s %s) Verify() error {", varNamePlainStruct, declType.Name)
		out.Indent(1)
		out.Printlnf("for _, expected := range %s {", valueConstant.Name)
		out.Indent(1)
		out.Printlnf("if reflect.DeepEqual(%s, expected) { return nil }", varName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf(
			`return fmt.Errorf("invalid value (expected one of %%#v): %%#v", %s, %s)`,
			valueConstant.Name,
			varName,
		)
		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (v verifyFormatter) verifyArray(tpe codegen.ArrayType, access string) func(*codegen.Emitter) error {
	aaccess := "a" + access

	verifyFn := v.verifyType(tpe.Type, aaccess)
	if verifyFn == nil {
		return nil
	}

	return func(out *codegen.Emitter) error {
		out.Printlnf("for _, %s := range %s {", aaccess, access)
		out.Indent(1)

		if err := verifyFn(out); err != nil {
			return fmt.Errorf("%w: %w", ErrCannotGenerateArrayVerificationCode, err)
		}

		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (v verifyFormatter) verifyMap(tpe codegen.MapType, access string) func(*codegen.Emitter) error {
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

	return func(out *codegen.Emitter) error {
		out.Printlnf("for %s, %s := range %s {", keyAccess, valueAccess, access)
		out.Indent(1)

		if verifyKeyFn != nil {
			if err := verifyKeyFn(out); err != nil {
				return fmt.Errorf("%w: %w", ErrCannotGenerateMapVerificationCode, err)
			}
		}

		if verifyValueFn != nil {
			if err := verifyValueFn(out); err != nil {
				return fmt.Errorf("%w: %w", ErrCannotGenerateMapVerificationCode, err)
			}
		}

		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}

func (v verifyFormatter) verifyPointer(tpe codegen.PointerType, access string) func(*codegen.Emitter) error {
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

	return func(out *codegen.Emitter) error {
		out.Printlnf("if %s != nil {", access)
		out.Printlnf("%s := %s%s", paccess, prefix, access)
		out.Indent(1)

		if err := verifyFn(out); err != nil {
			return fmt.Errorf("%w: %w", ErrCannotGeneratePointerVerificationCode, err)
		}

		out.Indent(-1)
		out.Printlnf("}")

		return nil
	}
}
