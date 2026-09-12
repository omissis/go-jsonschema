package schemas_test

import (
	"errors"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// recordingLoader is a stub Loader that records every Load call so tests can
// assert that the wrapped loader is (or isn't) invoked.
type recordingLoader struct {
	calls []string
	ret   *schemas.Schema
	err   error
}

func (r *recordingLoader) Load(uri, _ string) (*schemas.Schema, error) {
	r.calls = append(r.calls, uri)

	return r.ret, r.err
}

func TestCachedLoader_PrePopulatedHitSkipsWrapped(t *testing.T) {
	t.Parallel()

	const url = "https://example.com/known/v1/known.schema.json"

	wantSchema := &schemas.Schema{}
	wrapped := &recordingLoader{}

	loader := schemas.NewCachedLoader(wrapped, map[string]*schemas.Schema{url: wantSchema})

	got, err := loader.Load(url, "")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}

	if got != wantSchema {
		t.Errorf("Load: got %p, want pre-populated %p", got, wantSchema)
	}

	if n := len(wrapped.calls); n != 0 {
		t.Errorf("wrapped loader was invoked %d time(s) for cached URL: %v", n, wrapped.calls)
	}
}

func TestCachedLoader_MissDelegatesAndCaches(t *testing.T) {
	t.Parallel()

	const url = "https://example.com/uncached/v1/uncached.schema.json"

	wantSchema := &schemas.Schema{}
	wrapped := &recordingLoader{ret: wantSchema}

	loader := schemas.NewCachedLoader(wrapped, map[string]*schemas.Schema{})

	// First call: cache miss, wrapped loader is invoked.
	got, err := loader.Load(url, "")
	if err != nil {
		t.Fatalf("first Load: unexpected error: %v", err)
	}

	if got != wantSchema {
		t.Errorf("first Load: got %p, want %p", got, wantSchema)
	}

	if n := len(wrapped.calls); n != 1 {
		t.Fatalf("first Load: wrapped invoked %d times, want 1", n)
	}

	// Second call: should hit the populated cache, no further wrapped invocation.
	got, err = loader.Load(url, "")
	if err != nil {
		t.Fatalf("second Load: unexpected error: %v", err)
	}

	if got != wantSchema {
		t.Errorf("second Load: got %p, want %p", got, wantSchema)
	}

	if n := len(wrapped.calls); n != 1 {
		t.Errorf("second Load: wrapped invoked %d times after cache populated, want still 1", n)
	}
}

func TestCachedLoader_WrappedErrorIsWrapped(t *testing.T) {
	t.Parallel()

	const url = "https://example.com/missing/v1/missing.schema.json"

	wrappedErr := errors.New("synthetic load failure")
	wrapped := &recordingLoader{err: wrappedErr}
	loader := schemas.NewCachedLoader(wrapped, map[string]*schemas.Schema{})

	_, err := loader.Load(url, "")
	if err == nil {
		t.Fatal("expected error from wrapped loader, got nil")
	}

	if !errors.Is(err, schemas.ErrCannotLoadSchema) {
		t.Errorf("expected ErrCannotLoadSchema in chain, got %v", err)
	}

	if !errors.Is(err, wrappedErr) {
		t.Errorf("expected wrapped error in chain, got %v", err)
	}
}
