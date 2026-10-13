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

		case "duration":
			t = DurationType{}

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

		// An explicit `format` fixes the width the value is carried at.
		// Without this, `{"type":"integer","format":"int64"}` emits a bare
		// `int`, which is 32 bits on 32-bit platforms — so a value the schema
		// explicitly declares as 64-bit fails to decode there at all, while
		// the identical code and input succeed on a 64-bit build.
		//
		// `--min-sized-ints` still wins where it applies: it derives a width
		// from the declared bounds, which is strictly more information than
		// the format hint, and leaving it on top keeps this change invisible
		// to anyone already using that flag.
		if sized, ok := intTypeFromFormat(format); ok {
			t.Type = sized
		}

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

// intTypeFromFormat maps the OpenAPI integer formats onto Go's sized integer
// types. Only `int32` and `int64` are recognised: they are the two the OpenAPI
// Format Registry defines for `type: integer`, and they are what generators on
// the other side of these schemas emit. Anything else falls through to the
// caller's default so an unknown format never silently changes the type.
func intTypeFromFormat(format string) (string, bool) {
	switch format {
	case "int32":
		return "int32", true

	case i64:
		return i64, true

	default:
		return "", false
	}
}

func adjustForSignedBounds(nMin, nMax *float64) (string, bool, bool) {
	var minRounded, maxRounded float64

	if nMin != nil {
		minRounded = math.Round(*nMin)
	}

	if nMax != nil {
		maxRounded = math.Round(*nMax)
	}

	switch {
	case nMin == nil && nMax == nil:
		return i64, false, false

	case nMin == nil:
		return i64, false, maxRounded == float64(math.MaxInt64)

	case nMax == nil:
		return i64, minRounded == float64(math.MinInt64), false

	case minRounded < float64(math.MinInt32) || maxRounded > float64(math.MaxInt32):
		return i64, minRounded == float64(math.MinInt64), maxRounded == float64(math.MaxInt64)

	case minRounded < float64(math.MinInt16) || maxRounded > float64(math.MaxInt16):
		return "int32", minRounded == float64(math.MinInt32), maxRounded == float64(math.MaxInt32)

	case minRounded < float64(math.MinInt8) || maxRounded > float64(math.MaxInt8):
		return "int16", minRounded == float64(math.MinInt16), maxRounded == float64(math.MaxInt16)

	default:
		return "int8", minRounded == float64(math.MinInt8), maxRounded == float64(math.MaxInt8)
	}
}

func adjustForUnsignedBounds(nMin, nMax *float64) (string, bool, bool) {
	removeMin := nMin != nil && *nMin == 0.0

	var maxRounded float64

	if nMax != nil {
		maxRounded = math.Round(*nMax)
	}

	switch {
	case nMax == nil:
		return "uint64", removeMin, false

	case maxRounded > float64(math.MaxUint32):
		return "uint64", removeMin, maxRounded == float64(math.MaxUint64)

	case maxRounded > float64(math.MaxUint16):
		return "uint32", removeMin, maxRounded == float64(math.MaxUint32)

	case maxRounded > float64(math.MaxUint8):
		return "uint16", removeMin, maxRounded == float64(math.MaxUint16)

	default:
		return "uint8", removeMin, maxRounded == float64(math.MaxUint8)
	}
}
