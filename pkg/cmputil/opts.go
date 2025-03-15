package cmputil

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Opts(t ...any) []cmp.Option {
	opts := make([]cmp.Option, 0)

	for _, v := range t {
		opts = append(opts, cmpopts.IgnoreUnexported(v), cmpopts.IgnoreFields(v, "Ref"), cmpopts.IgnoreFields(v, "AnyOf"))
	}

	return opts
}
