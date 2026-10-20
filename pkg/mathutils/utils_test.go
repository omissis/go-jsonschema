package mathutils_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atombender/go-jsonschema/pkg/mathutils"
)

func TestNormalizeBounds(t *testing.T) {
	t.Parallel()

	anyMin := 100.0
	anyMax := 200.0

	var anySmallerMin any = 90.0

	var anyLargerMax any = 210.0

	var anyLargerMin any = 110.0

	var anySmallerMax any = 190.0

	var tr any = true

	var fa any = false

	t.Run("No exclusive bounds", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, nil, nil)
		assert.NotNil(t, nMin)
		assert.InEpsilon(t, anyMin, *nMin, 0.000001)
		assert.NotNil(t, nMax)
		assert.InEpsilon(t, anyMax, *nMax, 0.000001)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})

	t.Run("Less prohibitive exclusive bounds as numbers", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &anySmallerMin, &anyLargerMax)
		assert.NotNil(t, nMin)
		assert.InEpsilon(t, anyMin, *nMin, 0.000001)
		assert.NotNil(t, nMax)
		assert.InEpsilon(t, anyMax, *nMax, 0.000001)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})

	t.Run("More prohibitive exclusive bounds as numbers", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &anyLargerMin, &anySmallerMax)
		assert.NotNil(t, nMin)
		assert.InEpsilon(t, anyLargerMin, *nMin, 0.000001)
		assert.NotNil(t, nMax)
		assert.InEpsilon(t, anySmallerMax, *nMax, 0.000001)
		assert.True(t, nMinExclusive)
		assert.True(t, nMaxExclusive)
	})

	t.Run("Only exclusive bounds as numbers", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(nil, nil, &anyLargerMin, &anySmallerMax)
		assert.NotNil(t, nMin)
		assert.InEpsilon(t, anyLargerMin, *nMin, 0.000001)
		assert.NotNil(t, nMax)
		assert.InEpsilon(t, anySmallerMax, *nMax, 0.000001)
		assert.True(t, nMinExclusive)
		assert.True(t, nMaxExclusive)
	})

	t.Run("Exclusive bounds as bools", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &tr, &tr)
		assert.NotNil(t, nMin)
		assert.InEpsilon(t, anyMin, *nMin, 0.000001)
		assert.NotNil(t, nMax)
		assert.InEpsilon(t, anyMax, *nMax, 0.000001)
		assert.True(t, nMinExclusive)
		assert.True(t, nMaxExclusive)
	})

	t.Run("Exclusive bounds as false bools", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &fa, &fa)
		assert.NotNil(t, nMin)
		assert.InEpsilon(t, anyMin, *nMin, 0.000001)
		assert.NotNil(t, nMax)
		assert.InEpsilon(t, anyMax, *nMax, 0.000001)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})

	t.Run("No bounds", func(t *testing.T) {
		t.Parallel()

		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(nil, nil, nil, nil)
		assert.Nil(t, nMin)
		assert.Nil(t, nMax)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})
}

func TestSaturatingInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		in   float64
		want int64
	}{
		{desc: "zero", in: 0, want: 0},
		{desc: "truncates toward zero", in: 2.9, want: 2},
		{desc: "truncates negative toward zero", in: -2.9, want: -2},
		{desc: "int64 max as decoded from JSON", in: math.MaxInt64, want: math.MaxInt64},
		{desc: "int64 min", in: math.MinInt64, want: math.MinInt64},
		{desc: "above int64 range", in: 1e19, want: math.MaxInt64},
		{desc: "below int64 range", in: -1e19, want: math.MinInt64},
		{desc: "positive infinity", in: math.Inf(1), want: math.MaxInt64},
		{desc: "negative infinity", in: math.Inf(-1), want: math.MinInt64},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tC.want, mathutils.SaturatingInt64(tC.in))
		})
	}
}
