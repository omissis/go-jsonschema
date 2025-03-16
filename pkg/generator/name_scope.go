package generator

import (
	"fmt"
	"strings"
)

const maxUint = ^uint(0)

var (
	ErrCannotEnterSubschema          = fmt.Errorf("cannot enter subschema")
	ErrMaxNumberOfSubschemasExceeded = fmt.Errorf("exceeded max number of supported subschema")
	ErrCannotExitSubschema           = fmt.Errorf("cannot exit subschema")
	ErrNoSubschemasLeft              = fmt.Errorf("no subschemas left")
)

type nameScope struct {
	stack     []string
	subschema uint
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

func (ns nameScope) enterSubschema() (nameScope, error) {
	if ns.subschema < maxUint {
		ns.subschema++

		return ns, nil
	}

	return ns, fmt.Errorf("%w: %w", ErrCannotEnterSubschema, ErrMaxNumberOfSubschemasExceeded)
}

func (ns nameScope) exitSubschema() (nameScope, error) {
	if ns.subschema > 0 {
		ns.subschema--

		return ns, nil
	}

	return ns, fmt.Errorf("%w: %w", ErrCannotExitSubschema, ErrNoSubschemasLeft)
}
