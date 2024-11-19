package generator

import (
	"fmt"

	"github.com/google/go-cmp/cmp"

	"github.com/atombender/go-jsonschema/pkg/cmputil"
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

func (o *output) getDeclByEqualSchema(name string, t *schemas.Type) *codegen.TypeDecl {
	v, ok := o.declsByName[name]
	if !ok {
		o.warner(fmt.Sprintf("Name not found: %s", name))

		return nil
	}

	if cmp.Equal(v.SchemaType, t, cmputil.Opts(*v.SchemaType, *t)...) {
		return v
	}

	for count := 1; ; count++ {
		suffixed := fmt.Sprintf("%s_%d", name, count)

		sv, ok := o.declsByName[suffixed]
		if !ok {
			return nil
		}

		if cmp.Equal(sv.SchemaType, t, cmputil.Opts(*sv.SchemaType, *t)...) {
			return sv
		}
	}
}

func (o *output) isUniqueTypeName(name string) bool {
	v, ok := o.declsByName[name]

	return !ok || (ok && v.Type == nil)
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
