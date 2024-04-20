package generator

import (
	"fmt"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

type output struct {
	file          *codegen.File
	declsByName   map[string]*codegen.TypeDecl
	declsBySchema map[*schemas.Type]*codegen.TypeDecl
	warner        func(string)
}

func (o *output) uniqueTypeName(name string) string {
	v, ok := o.declsByName[name]

	if !ok || (ok && v.Type == nil) {
		return name
	}

	count := 1

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
