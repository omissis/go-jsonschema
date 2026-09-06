package generator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

var (
	errRefPathEmpty        = errors.New("reference path is empty")
	errRefPathUnsupported  = errors.New("unsupported keyword in reference path")
	errRefPathBadEscape    = errors.New("invalid JSON Pointer escape: ~ must be followed by 0 or 1")
	errRefPathNeedsSegment = errors.New("reference path ends on a keyword expecting a name or index")
)

// resolveRefPath walks a `$ref` path relative to a schema's definitions.
//
// The common case is a single segment naming a definition — `#/definitions/Foo`
// arrives here as "Foo". JSON Pointer allows descending further, though, and
// real schemas do: github-workflow refs
// `#/definitions/workflowDispatchInput/properties/options`, addressing a
// property subschema rather than a definition. Each subsequent step is a
// keyword naming the container to descend into, followed by a name or index
// where the keyword holds a collection.
//
// Only keywords whose values are schemas are traversable; anything else is
// rejected rather than silently resolving to the wrong node.
func resolveRefPath(defs schemas.Definitions, path string) (*schemas.Type, error) {
	segments, err := splitRefPath(path)
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		return nil, errRefPathEmpty
	}

	cur, ok := defs[segments[0]]
	if !ok || cur == nil {
		return nil, errDefinitionDoesNotExistInSchema
	}

	for i := 1; i < len(segments); i++ {
		next, consumed, err := stepRefPath(cur, segments[i:])
		if err != nil {
			return nil, err
		}

		cur = next
		i += consumed - 1
	}

	return cur, nil
}

// stepRefPath resolves one keyword step, returning the schema it names and how
// many segments were consumed (one for a direct keyword, two where the keyword
// holds a map or list and a key or index follows).
func stepRefPath(cur *schemas.Type, rest []string) (*schemas.Type, int, error) {
	keyword := rest[0]

	// Keywords holding a single schema.
	switch keyword {
	case "additionalProperties":
		return derefRefPath(cur.AdditionalProperties, 1)
	case "additionalItems":
		return derefRefPath(cur.AdditionalItems, 1)
	case "not":
		return derefRefPath(cur.Not, 1)
	}

	// `items` is dual-form: a tuple indexes positionally and consumes the
	// index, while the single-schema form takes no key and may end the path.
	if keyword == "items" {
		if len(rest) > 1 {
			if idx, err := strconv.Atoi(rest[1]); err == nil {
				return derefRefPathIndex(cur.TupleItems, idx, 2)
			}
		}

		return derefRefPath(cur.Items, 1)
	}

	// Keywords holding a collection, so a name or index must follow.
	if len(rest) < 2 {
		return nil, 0, fmt.Errorf("%w: %q", errRefPathNeedsSegment, keyword)
	}

	key := rest[1]

	switch keyword {
	case "properties":
		return derefRefPath(cur.Properties[key], 2)
	case "patternProperties":
		return derefRefPath(cur.PatternProperties[key], 2)
	case "definitions", "$defs":
		return derefRefPath(cur.Definitions[key], 2)
	case "allOf":
		return derefRefPathList(cur.AllOf, key)
	case "anyOf":
		return derefRefPathList(cur.AnyOf, key)
	case "oneOf":
		return derefRefPathList(cur.OneOf, key)
	}

	return nil, 0, fmt.Errorf("%w: %q", errRefPathUnsupported, keyword)
}

func derefRefPath(t *schemas.Type, consumed int) (*schemas.Type, int, error) {
	if t == nil {
		return nil, 0, errDefinitionDoesNotExistInSchema
	}

	return t, consumed, nil
}

func derefRefPathIndex(list []*schemas.Type, idx, consumed int) (*schemas.Type, int, error) {
	if idx < 0 || idx >= len(list) {
		return nil, 0, errDefinitionDoesNotExistInSchema
	}

	return derefRefPath(list[idx], consumed)
}

func derefRefPathList(list []*schemas.Type, key string) (*schemas.Type, int, error) {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %q is not an index", errDefinitionDoesNotExistInSchema, key)
	}

	return derefRefPathIndex(list, idx, 2)
}

// splitRefPath splits a JSON Pointer tail into segments, undoing the pointer
// escapes: `~1` for a literal `/` and `~0` for a literal `~`. Order matters —
// `~0` must be decoded last so that `~01` yields `~1` rather than `/`.
func splitRefPath(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}

	raw := strings.Split(path, "/")
	out := make([]string, 0, len(raw))

	for _, seg := range raw {
		if err := checkPointerEscapes(seg); err != nil {
			return nil, err
		}

		seg = strings.ReplaceAll(seg, "~1", "/")
		seg = strings.ReplaceAll(seg, "~0", "~")
		out = append(out, seg)
	}

	return out, nil
}

// checkPointerEscapes rejects tokens where `~` is not part of a valid escape.
// RFC 6901 defines only `~0` and `~1`, so anything else — `~2`, or a trailing
// `~` — is a malformed pointer. Without this the sequence would survive
// decoding untouched and silently resolve against a member whose name happens
// to contain it.
func checkPointerEscapes(segment string) error {
	for i := 0; i < len(segment); i++ {
		if segment[i] != '~' {
			continue
		}

		if i+1 >= len(segment) || (segment[i+1] != '0' && segment[i+1] != '1') {
			return fmt.Errorf("%w: %q", errRefPathBadEscape, segment)
		}

		i++
	}

	return nil
}

// refPathTypeName builds a Go-facing name for a nested reference by dropping
// the structural keywords and keeping the parts that carry meaning, so
// `workflowDispatchInput/properties/options` names a type after the definition
// and the property rather than after the plumbing between them.
func refPathTypeName(path string) string {
	structural := map[string]bool{
		"properties": true, "patternProperties": true,
		"definitions": true, "$defs": true,
		"items": true, "additionalProperties": true, "additionalItems": true,
		"not":   true,
		"allOf": true, "anyOf": true, "oneOf": true,
	}

	// Validation already happened in resolveRefPath; a malformed path cannot
	// reach here, and falling back to the raw path is harmless if it did.
	segments, err := splitRefPath(path)
	if err != nil {
		return path
	}

	kept := make([]string, 0, len(segments))

	for _, seg := range segments {
		if structural[seg] {
			continue
		}

		kept = append(kept, seg)
	}

	if len(kept) == 0 {
		return path
	}

	return strings.Join(kept, "_")
}
