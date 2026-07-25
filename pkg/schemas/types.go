package schemas

import "strings"

const (
	TypeNameString  = "string"
	TypeNameArray   = "array"
	TypeNameNumber  = "number"
	TypeNameInteger = "integer"
	TypeNameObject  = "object"
	TypeNameBoolean = "boolean"
	TypeNameNull    = "null"
	PrefixEnumValue = "enumValues_"
)

func IsPrimitiveType(t string) bool {
	switch t {
	case TypeNameString, TypeNameNumber, TypeNameInteger, TypeNameBoolean, TypeNameNull:
		return true

	default:
		return false
	}
}

func CleanNameForSorting(name string) string {
	if cleanName, found := strings.CutPrefix(name, PrefixEnumValue); found {
		return cleanName + "_enumValues" // Append a string for sorting properly.
	}
	// Preserve existing sort order in tests
	name = strings.TrimSuffix(name, "_yaml")
	name = strings.TrimSuffix(name, "_json")

	return name
}

func isPrimitiveTypeList(types []*Type, baseType TypeList) bool {
	if len(baseType) > 0 && !IsPrimitiveType(baseType[0]) {
		return false
	}

	// Track whether any schema actually contributes a primitive type. A
	// subschema with no "type" (for example one that only carries keywords
	// go-jsonschema does not model, such as if/then/else) provides no type
	// information and must not, on its own, classify the list as primitive.
	// Otherwise an object schema whose allOf holds only such subschemas would
	// be merged as a primitive, discarding the properties of the base schema
	// and collapsing the whole type to interface{}.
	sawPrimitive := len(baseType) > 0

	for _, typ := range types {
		if len(typ.Type) == 0 {
			continue
		}

		if !IsPrimitiveType(typ.Type[0]) {
			return false
		}

		sawPrimitive = true
	}

	return sawPrimitive
}
