package mathutils

// NormalizeBounds is a public function that normalizes the given bounds and exclusivity flags.
func NormalizeBounds(minimum, maximum *float64, exclusiveMinimum, exclusiveMaximum *any) (*float64, *float64, bool, bool) {
	var minBound, maxBound *float64
	var minExclusive, maxExclusive bool

	// Handle exclusiveMinimum
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

	// Handle minimum if exclusiveMinimum was not set
	if minimum != nil && minBound == nil {
		minBound = minimum
		minExclusive = false
	}

	// Handle exclusiveMaximum
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

	// Handle maximum if exclusiveMaximum was not set
	if maximum != nil && maxBound == nil {
		maxBound = maximum
		maxExclusive = false
	}

	return minBound, maxBound, minExclusive, maxExclusive
}
