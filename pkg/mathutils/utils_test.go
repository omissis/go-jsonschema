package mathutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atombender/go-jsonschema/pkg/mathutils"
)

func TestNormalizeBounds(t *testing.T) {
	anyMin := 100.0
	anyMax := 200.0
	var anySmallerMin any = 90.0
	var anyLargerMax any = 210.0
	var anyLargerMin any = 110.0
	var anySmallerMax any = 190.0
	var tr any = true
	var fa any = false
	t.Run("No exclusive bounds", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, nil, nil)
		assert.NotNil(t, nMin)
		assert.Equal(t, anyMin, *nMin)
		assert.NotNil(t, nMax)
		assert.Equal(t, anyMax, *nMax)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})

	t.Run("Less prohibitive exclusive bounds as numbers", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &anySmallerMin, &anyLargerMax)
		assert.NotNil(t, nMin)
		assert.Equal(t, anyMin, *nMin)
		assert.NotNil(t, nMax)
		assert.Equal(t, anyMax, *nMax)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})

	t.Run("More prohibitive exclusive bounds as numbers", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &anyLargerMin, &anySmallerMax)
		assert.NotNil(t, nMin)
		assert.Equal(t, anyLargerMin, *nMin)
		assert.NotNil(t, nMax)
		assert.Equal(t, anySmallerMax, *nMax)
		assert.True(t, nMinExclusive)
		assert.True(t, nMaxExclusive)
	})

	t.Run("Only exclusive bounds as numbers", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(nil, nil, &anyLargerMin, &anySmallerMax)
		assert.NotNil(t, nMin)
		assert.Equal(t, anyLargerMin, *nMin)
		assert.NotNil(t, nMax)
		assert.Equal(t, anySmallerMax, *nMax)
		assert.True(t, nMinExclusive)
		assert.True(t, nMaxExclusive)
	})

	t.Run("Exclusive bounds as bools", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &tr, &tr)
		assert.NotNil(t, nMin)
		assert.Equal(t, anyMin, *nMin)
		assert.NotNil(t, nMax)
		assert.Equal(t, anyMax, *nMax)
		assert.True(t, nMinExclusive)
		assert.True(t, nMaxExclusive)
	})

	t.Run("Exclusive bounds as false bools", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(&anyMin, &anyMax, &fa, &fa)
		assert.NotNil(t, nMin)
		assert.Equal(t, anyMin, *nMin)
		assert.NotNil(t, nMax)
		assert.Equal(t, anyMax, *nMax)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})

	t.Run("No bounds", func(t *testing.T) {
		nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(nil, nil, nil, nil)
		assert.Nil(t, nMin)
		assert.Nil(t, nMax)
		assert.False(t, nMinExclusive)
		assert.False(t, nMaxExclusive)
	})
}
