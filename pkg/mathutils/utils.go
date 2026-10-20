package mathutils

import "math"

// SaturatingInt64 converts f to int64, clamping values outside int64's range
// to its limits. A plain int64(f) is implementation-defined there: a JSON
// 9223372036854775807 decodes to float64 2^63, which saturates to
// math.MaxInt64 on arm64 but becomes math.MinInt64 on amd64.
func SaturatingInt64(f float64) int64 {
	switch {
	case f >= math.MaxInt64:
		return math.MaxInt64
	case f <= math.MinInt64:
		return math.MinInt64
	}

	return int64(f)
}

// NormalizeBounds is a public function that normalizes the given bounds and exclusivity flags.
func NormalizeBounds(
	minimum, maximum *float64, exclusiveMinimum, exclusiveMaximum *any,
) (*float64, *float64, bool, bool) {
	var minBound, maxBound *float64

	var minExclusive, maxExclusive bool

	if exclusiveMinimum != nil {
		switch v := (*exclusiveMinimum).(type) {
		case bool:
			minExclusive = v
			minBound = minimum

		case float64:
			if minimum == nil || v > *minimum {
				minBound = &v
				minExclusive = true
			} else {
				minBound = minimum
				minExclusive = false
			}
		}
	} else {
		minBound = minimum
		minExclusive = false
	}

	if minimum != nil && minBound == nil {
		minBound = minimum
		minExclusive = false
	}

	if exclusiveMaximum != nil {
		switch v := (*exclusiveMaximum).(type) {
		case bool:
			maxExclusive = v
			maxBound = maximum

		case float64:
			if maximum == nil || v < *maximum {
				maxBound = &v
				maxExclusive = true
			} else {
				maxBound = maximum
				maxExclusive = false
			}
		}
	} else {
		maxBound = maximum
		maxExclusive = false
	}

	if maximum != nil && maxBound == nil {
		maxBound = maximum
		maxExclusive = false
	}

	return minBound, maxBound, minExclusive, maxExclusive
}
