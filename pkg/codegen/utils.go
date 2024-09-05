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
	minimum **float64,
	maximum **float64,
	exclusiveMinimum **any,
	exclusiveMaximum **any,
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
			newType, removeMin, removeMax := getMinIntType(*minimum, *maximum, *exclusiveMinimum, *exclusiveMaximum)
			t.Type = newType

			if removeMin {
				*minimum = nil
				*exclusiveMaximum = nil
			}

			if removeMax {
				*maximum = nil
				*exclusiveMinimum = nil
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

// getMinIntType returns the smallest integer type that can represent the bounds, and if the bounds can be removed.
func getMinIntType(
	minimum, maximum *float64, exclusiveMinimum, exclusiveMaximum *any,
) (string, bool, bool) {
	nMin, nMax, nExclusiveMin, nExclusiveMax := mathutils.NormalizeBounds(
		minimum, maximum, exclusiveMinimum, exclusiveMaximum,
	)

	if nExclusiveMin && nMin != nil {
		*nMin += 1.0
	}

	if nExclusiveMax && nMax != nil {
		*nMax -= 1.0
	}

	if nMin != nil && *nMin >= 0 {
		return adjustForUnsignedBounds(nMin, nMax)
	}

	return adjustForSignedBounds(nMin, nMax)
}

const i64 = "int64"

func adjustForSignedBounds(nMin, nMax *float64) (string, bool, bool) {
	switch {
	case nMin == nil && nMax == nil:
		return i64, false, false

	case nMin == nil:
		return i64, false, int64(*nMax) == math.MaxInt64

	case nMax == nil:
		return i64, int64(*nMin) == math.MinInt64, false

	case *nMin < math.MinInt32 || *nMax > math.MaxInt32:
		return i64, int64(*nMin) == math.MinInt64, int64(*nMax) == math.MaxInt64

	case *nMin < math.MinInt16 || *nMax > math.MaxInt16:
		return "int32", int32(*nMin) == math.MinInt32, int32(*nMax) == math.MaxInt32

	case *nMin < math.MinInt8 || *nMax > math.MaxInt8:
		return "int16", int16(*nMin) == math.MinInt16, int16(*nMax) == math.MaxInt16

	default:
		return "int8", int8(*nMin) == math.MinInt8, int8(*nMax) == math.MaxInt8
	}
}

func adjustForUnsignedBounds(nMin, nMax *float64) (string, bool, bool) {
	removeMin := nMin != nil && *nMin == 0.0

	switch {
	case nMax == nil:
		return "uint64", removeMin, false

	case *nMax > math.MaxUint32:
		return "uint64", removeMin, uint64(*nMax) == math.MaxUint64

	case *nMax > math.MaxUint16:
		return "uint32", removeMin, uint32(*nMax) == math.MaxUint32

	case *nMax > math.MaxUint8:
		return "uint16", removeMin, uint16(*nMax) == math.MaxUint16

	default:
		return "uint8", removeMin, uint8(*nMax) == math.MaxUint8
	}
}
