package generator

import (
	"strings"
)

const maxUint = ^uint(0)

type nameScope struct {
	stack []string
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

	ns.stack = result

	return ns
}
