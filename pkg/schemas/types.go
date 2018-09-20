package schemas

const (
	TypeNameString  = "string"
	TypeNameArray   = "array"
	TypeNameNumber  = "number"
	TypeNameObject  = "object"
	TypeNameBoolean = "boolean"
	TypeNameNull    = "null"
)

func IsPrimitiveType(t string) bool {
	switch t {
	case TypeNameString, TypeNameNumber, TypeNameBoolean, TypeNameNull:
		return true
	default:
		return false
	}
}
