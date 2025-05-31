package generator

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

const additionalProperties = "AdditionalProperties"

var (
	alnumOnlyRegex               = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	errCannotConvertToStructName = errors.New("cannot convert to struct name")
)

func sortedKeys[T any](props map[string]T) []string {
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func sortDefinitionsByName(defs schemas.Definitions) []string {
	names := make([]string, 0, len(defs))

	for name := range defs {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func isNamedType(t codegen.Type) bool {
	switch x := t.(type) {
	case *codegen.NamedType:
		return true

	case *codegen.PointerType:
		if _, ok := x.Type.(*codegen.NamedType); ok {
			return true
		}
	}

	return false
}

func isMapType(t codegen.Type) bool {
	_, isMapType := t.(*codegen.MapType)

	return isMapType
}

// toStructName converts a string to a valid Go struct name.
func toStructName(s string) (string, error) {
	cleaned := alnumOnlyRegex.ReplaceAllString(s, " ")
	words := strings.Fields(cleaned)

	caser := cases.Title(language.English)
	for i, w := range words {
		words[i] = caser.String(w)
	}

	result := strings.Join(words, "")

	if result == "" || !unicode.IsLetter(rune(result[0])) {
		return "", errCannotConvertToStructName
	}

	if isGoKeyword(strings.ToLower(result)) {
		return "", errCannotConvertToStructName
	}

	return result, nil
}

func isGoKeyword(s string) bool {
	switch s {
	case "break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var":
		return true
	default:
		return false
	}
}
