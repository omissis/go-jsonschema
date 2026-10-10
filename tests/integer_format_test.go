package tests_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	testIntegerFormat "github.com/atombender/go-jsonschema/tests/data/core/integerFormat"
)

// TestIntegerFormatEnumValidation covers integer enums declaring a sized
// format. Each validator decodes into the sized type and compares every allowed
// value with reflect.DeepEqual; an untyped allowed-values list holds ints, so
// no value ever matched and every one was rejected, valid ones included.
func TestIntegerFormatEnumValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		payload string
		valid   bool
	}{
		{`{"code": 66}`, true},
		{`{"code": 67}`, true},
		{`{"code": 65}`, false},
		{`{"wide": 1643178300132}`, true},
		{`{"wide": 1643178300133}`, false},
		// A member the width cannot hold must not cost the members it can.
		{`{"mixed": 66}`, true},
		{`{"mixed": 2147483648}`, false},
	} {
		var v testIntegerFormat.IntegerFormatEnum

		err := json.Unmarshal([]byte(tc.payload), &v)
		if tc.valid {
			require.NoError(t, err, "payload %s", tc.payload)
		} else {
			require.Error(t, err, "payload %s", tc.payload)
		}
	}
}
