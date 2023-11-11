package generator

import (
	"strings"
)

type nameScope []string

func newNameScope(s string) nameScope {
	return nameScope{s}
}

func (ns nameScope) string() string {
	return strings.Join(ns, "")
}

func (ns nameScope) add(s string) nameScope {
	result := make(nameScope, len(ns)+1)
	copy(result, ns)
	result[len(result)-1] = s

	return result
}
