package codegen

import (
	"fmt"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func PrimitiveTypeFromJSONSchemaType(jsType string) (Type, error) {
	switch jsType {
	case schemas.TypeNameString:
		return PrimitiveType{"string"}, nil
	case schemas.TypeNameNumber:
		return PrimitiveType{"float64"}, nil
	case schemas.TypeNameBoolean:
		return PrimitiveType{"bool"}, nil
	case schemas.TypeNameNull:
		return EmptyInterfaceType{}, nil
	case schemas.TypeNameObject, schemas.TypeNameArray:
		return nil, fmt.Errorf("unexpected type %q here", jsType)
	}
	return nil, fmt.Errorf("unknown JSON Schema type %q", jsType)
}
