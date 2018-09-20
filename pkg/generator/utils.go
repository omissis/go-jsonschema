package generator

import (
	"crypto/sha256"
	"fmt"
	"sort"
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
