package codegen

import (
	"errors"
	"fmt"
	"math"

	"github.com/atombender/go-jsonschema/pkg/mathutils"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

var (
	errUnexpectedType        = errors.New("unexpected type")
	errUnknownJSONSchemaType = errors.New("unknown JSON Schema type")
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

func PrimitiveTypeFromJSONSchemaType(
	jsType,
	format string,
	pointer,
	minIntSize bool,
	minimum *float64,
	maximum *float64,
	exclusiveMinimum *any,
	exclusiveMaximum *any,
) (Type, error) {
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
		if minIntSize {
			newType, removeBoudns := getMinIntType(minimum, maximum, exclusiveMinimum, exclusiveMaximum)
			t.Type = newType
			if removeBoudns {

			}
		}

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

// getMinIntType returns the smallest integer type that can represent the bounds, and if the bounds can be removed
func getMinIntType(minimum *float64, maximum *float64, exclusiveMinimum *any, exclusiveMaximum *any) (string, bool) {
	nMin, nMax, nExclusiveMin, nExclusiveMax := mathutils.NormalizeBounds(minimum, maximum, exclusiveMinimum, exclusiveMaximum)
	if nExclusiveMin && nMin != nil {
		*nMin = *nMin - 1.0
	}

	if nExclusiveMax && nMax != nil {
		*nMax = *nMax + 1.0
	}

	if nMin != nil && *nMin >= 0 {
		return adjustForUnsignedBounds(nMax)
	}

	return adjustForSignedBounds(nMin, nMax)
}

func adjustForSignedBounds(nMin *float64, nMax *float64) (string, bool) {
	if nMin == nil || nMax == nil {
		return "int64", true
	} else if *nMin < math.MinInt32 || *nMax > math.MaxInt32 {
		return "int64", *nMin == math.MinInt64 && *nMax == math.MaxInt64
	} else if *nMin < math.MinInt16 || *nMax > math.MaxInt16 {
		return "int32", *nMin == math.MinInt32 && *nMax == math.MaxInt32
	} else if *nMin < math.MinInt8 || *nMax > math.MaxInt8 {
		return "int16", *nMin == math.MinInt16 && *nMax == math.MaxInt16
	}

	return "int8", *nMin == math.MinInt8 && *nMax == math.MaxInt8
}

func adjustForUnsignedBounds(nMax *float64) (string, bool) {
	if nMax == nil {
		return "uint64", true
	} else if *nMax > math.MaxUint32 {
		return "uint64", *nMax == math.MaxUint64
	} else if *nMax > math.MaxUint16 {
		return "uint32", *nMax == math.MaxUint32
	} else if *nMax > math.MaxUint8 {
		return "uint16", *nMax == math.MaxUint16
	}

	return "uint8", *nMax == math.MaxUint8
}
