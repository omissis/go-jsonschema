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

func PrimitiveTypeFromJSONSchemaType(jsType string) (Type, error) {
	switch jsType {
	case schemas.TypeNameString:
		return PrimitiveType{"string"}, nil
	case schemas.TypeNameNumber:
		return PrimitiveType{"float64"}, nil
	case schemas.TypeNameInteger:
		return PrimitiveType{"int"}, nil
	case schemas.TypeNameBoolean:
		return PrimitiveType{"bool"}, nil
	case schemas.TypeNameNull:
		return NullType{}, nil
	case schemas.TypeNameObject, schemas.TypeNameArray:
		return nil, fmt.Errorf("unexpected type %q here", jsType)
	}
	return nil, fmt.Errorf("unknown JSON Schema type %q", jsType)
}
