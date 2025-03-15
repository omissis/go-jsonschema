package generator

import (
	"strings"
)

type nameScope struct {
	stack     []string
	subschema bool
}

func newNameScope(s string) nameScope {
	return nameScope{stack: []string{s}}
}

func (ns nameScope) string() string {
	return strings.Join(ns.stack, "")
}

func (ns nameScope) add(s string) nameScope {
	result := make([]string, len(ns.stack)+1)
	copy(result, ns.stack)
	result[len(result)-1] = s

	return ns
}

func (ns nameScope) enterSubschema() nameScope {
	ns.subschema = true

	return ns
}

func (ns nameScope) exitSubschema() nameScope {
	ns.subschema = false

	return ns
}
