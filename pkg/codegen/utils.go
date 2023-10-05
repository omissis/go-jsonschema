package codegen

import (
	"fmt"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

var (
	errUnexpectedType        = fmt.Errorf("unexpected type")
	errUnknownJSONSchemaType = fmt.Errorf("unknown JSON Schema type")
)

func WrapTypeInPointer(t Type) Type {
	if isPointerType(t) {
		return t
	}

	return &PointerType{Type: t}
}

func isPointerType(t Type) bool {
	switch x := t.(type) {
	case *PointerType:
		return true

	case *NamedType:
		return isPointerType(x.Decl.Type)

	default:
		return false
	}
}

func PrimitiveTypeFromJSONSchemaType(jsType, format string, pointer bool) (Type, error) {
	var t Type

	switch jsType {
	case schemas.TypeNameString:
		switch format {
		case "ipv4", "ipv6":
			t = NamedType{
				Package: &Package{
					QualifiedName: "net/netip",
					Imports: []Import{
						{
							QualifiedName: "net/netip",
						},
					},
				},
				Decl: &TypeDecl{
					Name: "Addr",
				},
			}

		case "date-time":
			t = NamedType{
				Package: &Package{
					QualifiedName: "time",
					Imports: []Import{
						{
							QualifiedName: "time",
						},
					},
				},
				Decl: &TypeDecl{
					Name: "Time",
				},
			}

		case "date":
			t = NamedType{
				Package: &Package{
					QualifiedName: "types",
					Imports: []Import{
						{
							QualifiedName: "github.com/atombender/go-jsonschema/pkg/types",
						},
					},
				},
				Decl: &TypeDecl{
					Name: "SerializableDate",
				},
			}

		case "time":
			t = NamedType{
				Package: &Package{
					QualifiedName: "types",
					Imports: []Import{
						{
							QualifiedName: "github.com/atombender/go-jsonschema/pkg/types",
						},
					},
				},
				Decl: &TypeDecl{
					Name: "SerializableTime",
				},
			}

		default:
			t = PrimitiveType{"string"}
		}

		if pointer {
			return WrapTypeInPointer(t), nil
		}

		return t, nil

	case schemas.TypeNameNumber:
		t := PrimitiveType{"float64"}
		if pointer {
			return WrapTypeInPointer(t), nil
		}

		return t, nil

	case schemas.TypeNameInteger:
		t := PrimitiveType{"int"}
		if pointer {
			return WrapTypeInPointer(t), nil
		}

		return t, nil

	case schemas.TypeNameBoolean:
		t := PrimitiveType{"bool"}
		if pointer {
			return WrapTypeInPointer(t), nil
		}

		return t, nil

	case schemas.TypeNameNull:
		return NullType{}, nil

	case schemas.TypeNameObject, schemas.TypeNameArray:
		return nil, fmt.Errorf("%w %q here", errUnexpectedType, jsType)
	}

	return nil, fmt.Errorf("%w %q", errUnknownJSONSchemaType, jsType)
}
