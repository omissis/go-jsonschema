package generator

import (
	"fmt"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

type output struct {
	minimalNames  bool
	file          *codegen.File
	declsByName   map[string]*codegen.TypeDecl
	declsBySchema map[*schemas.Type]*codegen.TypeDecl
	warner        func(string)
}

// uniqueTypeName finds the shortest identifier in a name scope that yields a unique type name.
// If a given suffix on the name scope is not unique, more context from the scope is added. If the
// entire context does not yield a unique name, a numeric suffix is used.
// TODO: we should check for schema equality on name collisions here to deduplicate identifiers.
func (o *output) uniqueTypeName(scope nameScope) string {
	if o.minimalNames {
		for i := len(scope) - 1; i >= 0; i-- {
			name := scope[i:].string()

			v, ok := o.declsByName[name]
			if !ok || (ok && v.Type == nil) {
				// An identifier using the current amount of name context is unique, use it.
				return name
			}
		}
	}

	// If we can't make a unique name with the entire context, attempt a numeric suffix.
	count := 1
	name := scope.string()

	v, ok := o.declsByName[name]
	if !ok || (ok && v.Type == nil) {
		return name
	}

	for {
		suffixed := fmt.Sprintf("%s_%d", name, count)
		if _, ok := o.declsByName[suffixed]; !ok {
			o.warner(fmt.Sprintf(
				"Multiple types map to the name %q; declaring duplicate as %q instead", name, suffixed))

			return suffixed
		}

		count++
	}
}
