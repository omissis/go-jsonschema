package generator

import (
	"slices"
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

// resolvesToRef reports whether generateType renders t as the type its $ref
// points to, rather than as an enum or a goJSONSchema type override.
func resolvesToRef(t *schemas.Type) bool {
	if t.Ref == "" || t.Enum != nil {
		return false
	}

	ext := t.GoJSONSchemaExtension

	return ext == nil || ext.Type == nil
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

// mergesIntoStruct reports whether merging parent's anyOf branches yields an
// object with properties, which becomes a struct rather than a map. The merge
// keeps the first type list it meets, the parent's before the branches', and
// collects the properties of all of them.
func mergesIntoStruct(parent *schemas.Type, branches []*schemas.Type) bool {
	types := parent.Type

	for _, b := range branches {
		if len(types) > 0 {
			break
		}

		types = b.Type
	}

	if !slices.Contains(types, schemas.TypeNameObject) {
		return false
	}

	if len(parent.Properties) > 0 {
		return true
	}

	for _, b := range branches {
		if len(b.Properties) > 0 {
			return true
		}
	}

	return false
}

// mixesObjectAndNonObject reports whether some of parent's anyOf branches
// admit objects while others admit a primitive or an array.
func (g *schemaGenerator) mixesObjectAndNonObject(parent *schemas.Type, branches []*schemas.Type) bool {
	var object, other bool

	seen := map[*schemas.Type]bool{}

	for _, b := range branches {
		o, n := g.branchAdmits(parent.Type, b, seen)
		object = object || o
		other = other || n
	}

	return object && other
}

// branchAdmits reports whether an anyOf branch admits objects and whether it
// admits anything besides objects and null. A branch without a type takes the
// enclosing one, as generateAnyOfType gives it. Without either, object
// keywords make it an object; an anyOf or oneOf admits what its members do;
// and the empty schema, a bare not, and an enum or const of non-objects admit
// other values too. Anything else, such as a branch listing only required,
// counts as neither.
func (g *schemaGenerator) branchAdmits(
	enclosing schemas.TypeList,
	b *schemas.Type,
	seen map[*schemas.Type]bool,
) (bool, bool) {
	if b.Ref != "" {
		if b = g.lookupRef(b); b == nil {
			return false, false
		}
	}

	if seen[b] {
		return false, false
	}

	seen[b] = true

	types := b.Type
	if len(types) == 0 {
		types = enclosing
	}

	if len(types) > 0 {
		var object, other bool

		for _, name := range types {
			switch name {
			case schemas.TypeNameObject:
				object = true

			case schemas.TypeNameNull:

			default:
				other = true
			}
		}

		return object, other
	}

	if len(b.Properties) > 0 || len(b.PatternProperties) > 0 || b.AdditionalProperties != nil {
		return true, false
	}

	if members := slices.Concat(b.AnyOf, b.OneOf); len(members) > 0 {
		var object, other bool

		for _, m := range members {
			o, n := g.branchAdmits(nil, m, seen)
			object = object || o
			other = other || n
		}

		return object, other
	}

	rest := *b
	rest.Not = nil

	if isAnnotationOnly(&rest) {
		return false, true
	}

	values := make([]any, 0, len(b.Enum)+1)
	values = append(values, b.Enum...)

	if b.Const != nil {
		values = append(values, b.Const)
	}

	for _, v := range values {
		if _, isObject := v.(map[string]any); !isObject && v != nil {
			return false, true
		}
	}

	return false, false
}
