package generator

import (
	"sort"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func sortPropertiesByName(props map[string]*schemas.Type) []string {
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
