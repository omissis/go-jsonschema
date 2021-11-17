package codegen

import (
	"fmt"

	"github.com/atombender/go-jsonschema/pkg/schemas"
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

func PrimitiveTypeFromJSONSchemaType(jsType string, pointer bool) (Type, error) {
	switch jsType {
	case schemas.TypeNameString:
		t := PrimitiveType{"string"}
		if pointer == true {
			return WrapTypeInPointer(t), nil
		}
		return t, nil
	case schemas.TypeNameNumber:
		t := PrimitiveType{"float64"}
		if pointer == true {
			return WrapTypeInPointer(t), nil
		}
		return t, nil
	case schemas.TypeNameInteger:
		t := PrimitiveType{"int"}
		if pointer == true {
			return WrapTypeInPointer(t), nil
		}

		return t, nil
	case schemas.TypeNameBoolean:
		t := PrimitiveType{"bool"}
		if pointer == true {
			return WrapTypeInPointer(t), nil
		}

		return t, nil
	case schemas.TypeNameNull:
		return NullType{}, nil
	case schemas.TypeNameObject, schemas.TypeNameArray:
		return nil, fmt.Errorf("unexpected type %q here", jsType)
	}
	return nil, fmt.Errorf("unknown JSON Schema type %q", jsType)
}
