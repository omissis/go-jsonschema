package schemas_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func TestResolveRef(t *testing.T) {
	t.Parallel()

	t.Run("direct Loader implements RefResolver", func(t *testing.T) {
		t.Parallel()

		qualified, err := schemas.ResolveRef(prefixResolverLoader{}, "child.json", "parent.json", nil)

		// We should get the prefix from the loader's RefResolver.
		require.NoError(t, err)
		assert.Equal(t, "resolved:parent.json:child.json", qualified)
	})

	t.Run("wrapped Loader implements RefResolver", func(t *testing.T) {
		t.Parallel()

		// Wrap the prefixResolverLoader in a CachedLoader
		loader := schemas.NewCachedLoader(prefixResolverLoader{}, map[string]*schemas.Schema{})
		qualified, err := schemas.ResolveRef(loader, "child.json", "parent.json", nil)

		// We should still get the expected resolved: prefix from the wrapped
		// loader's RefResolver.
		require.NoError(t, err)
		assert.Equal(t, "resolved:parent.json:child.json", qualified)
	})

	t.Run("multiply-wrapped Loader implements RefResolver", func(t *testing.T) {
		t.Parallel()

		// Wrap the prefixResolverLoader in a CachedLoader, and that in a
		// passthroughLoader. Neither wrapper implements RefResolver.
		loader := passthroughLoader{
			wrapped: schemas.NewCachedLoader(prefixResolverLoader{}, map[string]*schemas.Schema{}),
		}
		qualified, err := schemas.ResolveRef(loader, "child.json", "parent.json", nil)

		// ResolveRef should follow Unwrap through both layers to find the
		// prefixResolverLoader's RefResolver.
		require.NoError(t, err)
		assert.Equal(t, "resolved:parent.json:child.json", qualified)
	})

	t.Run("no RefResolver falls back to the native filesystem", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		childPath := filepath.Join(dir, "child.json")
		require.NoError(t, os.WriteFile(childPath, []byte("{}"), 0o600))

		expected, err := filepath.EvalSymlinks(childPath)
		require.NoError(t, err)

		// Construct a FileLoader that doesn't implement RefResolver
		loader := schemas.NewFileLoader([]string{".json"}, nil)

		// We should get the expected filesystem qualified ref.
		qualified, err := schemas.ResolveRef(loader, "child", filepath.Join(dir, "parent.json"), []string{".json"})
		require.NoError(t, err)
		assert.Equal(t, expected, qualified)
	})
}

// passthroughLoader wraps another Loader, implementing Unwrap but not
// RefResolver.
type passthroughLoader struct {
	wrapped schemas.Loader
}

func (l passthroughLoader) Load(fileName, parentFileName string) (*schemas.Schema, error) {
	schema, err := l.wrapped.Load(fileName, parentFileName)
	if err != nil {
		return nil, fmt.Errorf("passthrough load of %s: %w", fileName, err)
	}

	return schema, nil
}

func (l passthroughLoader) Unwrap() schemas.Loader {
	return l.wrapped
}

// prefixResolverLoader is an impl. of RefResolver that adds a 'resolved:'
// prefix to the resolved name.
type prefixResolverLoader struct{}

func (prefixResolverLoader) Load(_, _ string) (*schemas.Schema, error) {
	return &schemas.Schema{}, nil
}

func (prefixResolverLoader) ResolveRef(fileName, parentFileName string) (string, error) {
	return "resolved:" + parentFileName + ":" + fileName, nil
}
