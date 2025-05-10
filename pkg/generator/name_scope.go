package generator

import (
	"strings"
)

type nameScope struct {
	stack []string
}

func newNameScope(s string) nameScope {
	return nameScope{stack: []string{s}}
}

func (ns nameScope) string() string {
	return strings.Join(ns.stack, "")
}

func (ns nameScope) stringFrom(start int) string {
	if start >= len(ns.stack) {
		return ""
	}

	return strings.Join(ns.stack[start:], "")
}

func (ns nameScope) add(s string) nameScope {
	result := make([]string, len(ns.stack)+1)
	copy(result, ns.stack)
	result[len(result)-1] = s

	ns.stack = result

	return ns
}

func (ns nameScope) last() (string, bool) {
	if len(ns.stack) == 0 {
		return "", false
	}

	return ns.stack[len(ns.stack)-1], true
}

func (ns nameScope) len() int {
	return len(ns.stack)
}
