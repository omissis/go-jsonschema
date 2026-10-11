package generator

import (
	"reflect"
	"sort"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

const additionalProperties = "AdditionalProperties"

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

// namedDecl returns the declaration t names, directly or through a pointer, or
// nil when t is not a named type.
func namedDecl(t codegen.Type) *codegen.TypeDecl {
	switch x := t.(type) {
	case *codegen.NamedType:
		return x.Decl

	case *codegen.PointerType:
		if nt, ok := x.Type.(*codegen.NamedType); ok {
			return nt.Decl
		}
	}

	return nil
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

// isAnnotationOnly reports whether t declares nothing but a title or a
// description.
func isAnnotationOnly(t *schemas.Type) bool {
	rest := *t
	rest.Title, rest.Description = "", ""

	return reflect.DeepEqual(rest, schemas.Type{})
}
