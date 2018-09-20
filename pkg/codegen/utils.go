package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

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

func IdentifierFromFileName(fileName string) string {
	s := filepath.Base(fileName)
	return Identifierize(strings.TrimSuffix(strings.TrimSuffix(s, ".json"), ".schema"))
}

func Identifierize(s string) string {
	// FIXME: Better handling of non-identifier chars
	var sb strings.Builder
	seps := "_-. \t"
	for {
		i := strings.IndexAny(s, seps)
		if i == -1 {
			sb.WriteString(capitalize(s))
			break
		}
		sb.WriteString(capitalize(s[0:i]))
		for i < len(s) && strings.ContainsRune(seps, rune(s[i])) {
			i++
		}
		if i >= len(s) {
			break
		}
		s = s[i:]
	}
	return sb.String()
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}
