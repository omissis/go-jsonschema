package generator

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

func hashArrayOfValues(values []interface{}) string {
	sorted := make([]interface{}, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return fmt.Sprintf("%#v", sorted[i]) < fmt.Sprintf("%#v", sorted[j])
	})

	h := sha256.New()
	for _, v := range sorted {
		h.Write([]byte(fmt.Sprintf("%#v", v)))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func makeEnumConstantName(typeName, value string) string {
	if strings.ContainsAny(typeName[len(typeName)-1:], "0123456789") {
		return typeName + "_" + codegen.Identifierize(value)
	}
	return typeName + codegen.Identifierize(value)
}
