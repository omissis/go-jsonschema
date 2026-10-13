package schemas_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func parseRoot(t *testing.T, doc string) *schemas.Schema {
	t.Helper()

	s, err := schemas.FromJSONReader(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	return s
}

// TestRootSchemaRoutesDualFormDependencies checks that the root goes through
// the same decoding as any subschema. Decoded straight into the embedded
// ObjectAsType, the root skipped the dual-form `dependencies` routing, so both
// forms were silently dropped there.
func TestRootSchemaRoutesDualFormDependencies(t *testing.T) {
	t.Parallel()

	s := parseRoot(t, `{
		"type": "object",
		"dependencies": {
			"a": ["b"],
			"c": {"required": ["d"]}
		}
	}`)

	if got := s.DependentRequired["a"]; !reflect.DeepEqual(got, []string{"b"}) {
		t.Errorf("DependentRequired[a] = %v, want [b]", got)
	}

	dep := s.DependentSchemas["c"]
	if dep == nil || !reflect.DeepEqual(dep.Required, []string{"d"}) {
		t.Errorf("DependentSchemas[c] = %+v, want a schema requiring d", dep)
	}
}

// TestRootSchemaKeepsIDsAndDefinitions pins what the root decoder still owns:
// the $id/id fallback, and definitions living on Schema rather than also on
// the embedded type.
func TestRootSchemaKeepsIDsAndDefinitions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		doc     string
		wantID  string
		wantDef string
	}{
		{
			name:    "$id and $defs",
			doc:     `{"$id": "https://example.com/a", "$defs": {"x": {"type": "string"}}}`,
			wantID:  "https://example.com/a",
			wantDef: "x",
		},
		{
			name:    "legacy id and definitions",
			doc:     `{"id": "https://example.com/b", "definitions": {"y": {"type": "string"}}}`,
			wantID:  "https://example.com/b",
			wantDef: "y",
		},
		{
			name:   "$id wins over id",
			doc:    `{"$id": "https://example.com/c", "id": "https://example.com/legacy"}`,
			wantID: "https://example.com/c",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := parseRoot(t, tc.doc)

			if s.ID != tc.wantID {
				t.Errorf("ID = %q, want %q", s.ID, tc.wantID)
			}

			if _, ok := s.Definitions[tc.wantDef]; tc.wantDef != "" && !ok {
				t.Errorf("Definitions = %v, want key %q", s.Definitions, tc.wantDef)
			}

			// A root with only ids and definitions still decodes to a type,
			// exactly as it does once it adds $schema or a title.
			if s.ObjectAsType == nil {
				t.Fatal("root decoded without a type")
			}

			if s.ObjectAsType.Definitions != nil {
				t.Errorf("root definitions duplicated onto the embedded type: %v", s.ObjectAsType.Definitions)
			}
		})
	}
}

// TestRootSchemaAcceptsTupleItems covers a draft-07 tuple at the root, which
// failed the whole parse while the root bypassed Type's decoding.
func TestRootSchemaAcceptsTupleItems(t *testing.T) {
	t.Parallel()

	s := parseRoot(t, `{"type": "array", "items": [{"type": "string"}, {"type": "integer"}], "additionalItems": false}`)

	if len(s.TupleItems) != 2 {
		t.Fatalf("TupleItems = %d entries, want 2", len(s.TupleItems))
	}

	if got := s.TupleItems[1].Type; !reflect.DeepEqual(got, schemas.TypeList{"integer"}) {
		t.Errorf("TupleItems[1].Type = %v, want [integer]", got)
	}
}

// TestRootSchemaRejectsNullItems covers `items: null` at the root. A null never
// reaches a decoder hook, so only the pre-processing Type does can reject it.
func TestRootSchemaRejectsNullItems(t *testing.T) {
	t.Parallel()

	_, err := schemas.FromJSONReader(strings.NewReader(`{"type": "array", "items": null}`))
	if !errors.Is(err, schemas.ErrNullNotASchema) {
		t.Fatalf("err = %v, want %v", err, schemas.ErrNullNotASchema)
	}
}
