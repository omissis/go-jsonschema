package cmputil

import (
	"math"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Opts(t ...any) []cmp.Option {
	// Guard against potential overflow when computing the slice capacity.
	// If n is too large to safely multiply by 3 within the range of int,
	// fall back to a zero capacity so that the slice can grow dynamically.
	n := min(max(0, len(t)), math.MaxInt/3)

	opts := make([]cmp.Option, 0, 3*n)

	for _, v := range t {
		opts = append(opts, cmpopts.IgnoreUnexported(v), cmpopts.IgnoreFields(v, "Ref"), cmpopts.IgnoreFields(v, "AnyOf"))
	}

	return opts
}
