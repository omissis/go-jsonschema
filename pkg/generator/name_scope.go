package generator

import (
	"strings"
)

type nameScope struct {
	stack []string

	// keepRoot suppresses --minimal-names shortening for this scope.
	//
	// Shortening works by dropping leading elements, so there is no way to
	// keep the root while dropping anything else — it is all or nothing. The
	// qualify collision strategy roots a definition's scope at its owning
	// schema's type name, and dropping that prefix is precisely what makes a
	// shared definition name bind positionally to whichever file happened to
	// be processed first. So a rooted scope is never shortened.
	keepRoot bool
}

func newNameScope(s string) nameScope {
	return nameScope{stack: []string{s}}
}

// newRootedNameScope starts a scope at root, which --minimal-names will not
// shorten away.
func newRootedNameScope(root, s string) nameScope {
	return nameScope{stack: []string{root, s}, keepRoot: true}
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
