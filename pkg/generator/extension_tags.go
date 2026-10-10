package generator

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// extensionTags renders the configured `x-` extensions declared on prop as
// struct tags.
//
// A schema carries vendor data the generated Go type otherwise drops on the
// floor — `x-dimension: "Volume"` says what a number means, and a
// reflection-based consumer looks for exactly that on the struct field. Other
// generators reading the same file already emit it, so without this the same
// schema produces types that work with such a consumer through one generator
// and not through another.
//
// Only extensions named in Config.ExtensionTags are emitted, and only onto the
// field that declares them. The mapping is explicit because a struct tag key is
// part of a package's contract with whatever reflects over it: the consumer
// picks the name, not the schema author.
func (g *schemaGenerator) extensionTags(prop *schemas.Type, propName string) []string {
	if len(g.config.ExtensionTags) == 0 || prop == nil || len(prop.Extensions) == 0 {
		return nil
	}

	// Sorted so the emitted tag order does not depend on map iteration.
	extNames := make([]string, 0, len(g.config.ExtensionTags))
	for extName := range g.config.ExtensionTags {
		extNames = append(extNames, extName)
	}

	slices.Sort(extNames)

	// goJSONSchema.extraTags is the schema's own instruction for this one
	// field, so a key it already sets is left to it. Emitting the extension
	// as well would only add a second entry, which reflect never reads.
	var fieldTags map[string]string
	if prop.GoJSONSchemaExtension != nil {
		fieldTags = prop.GoJSONSchemaExtension.ExtraTags
	}

	tags := make([]string, 0, len(extNames))

	for _, extName := range extNames {
		raw, ok := prop.Extensions[extName]
		if !ok {
			continue
		}

		key := g.config.ExtensionTags[extName]

		if _, set := fieldTags[key]; set {
			g.warner(fmt.Sprintf(
				"Property %q declares %s, but its goJSONSchema.extraTags already sets "+
					"the %s tag; the field keeps that one and the extension is skipped",
				propName, extName, key,
			))

			continue
		}

		value, ok := extensionTagValue(raw)
		if !ok {
			g.warner(fmt.Sprintf(
				"Property %q declares %s with a %T; only strings, numbers and booleans "+
					"can be emitted as a struct tag, so it is skipped",
				propName, extName, raw,
			))

			continue
		}

		// The tag list is emitted inside a raw string literal, which a
		// backtick would terminate — that is a compile error in the
		// generated file, not a mangled tag, so there is nothing sensible
		// to escape it to.
		if strings.ContainsRune(value, '`') {
			g.warner(fmt.Sprintf(
				"Property %q declares %s with a value containing a backtick, which "+
					"cannot appear in a struct tag; it is skipped",
				propName, extName,
			))

			continue
		}

		// strconv.Quote rather than %q-into-a-hand-built-string: the value
		// is author-supplied, and an embedded quote would otherwise close
		// the tag early and silently change what reflection reads back.
		tags = append(tags, key+":"+strconv.Quote(value))
	}

	return tags
}

// extensionTagValue renders a scalar extension value as struct tag text.
//
// Composite values are refused rather than stringified: a struct tag is a flat
// string, so any rendering of an object or array would be this generator's
// invention rather than something the schema said, and a consumer reflecting
// over it would have to guess the encoding back.
func extensionTagValue(raw any) (string, bool) {
	switch v := raw.(type) {
	case string:
		return v, true

	case bool:
		return strconv.FormatBool(v), true

	case json.Number:
		// Schema-sourced numbers arrive undecoded, so the exact literal the
		// author wrote is what reaches the tag.
		return v.String(), true

	case float64:
		// Kept for extensions built programmatically rather than parsed.
		// Render integral values without a trailing ".0", which is what a
		// reader would expect to parse back.
		return strconv.FormatFloat(v, 'f', -1, 64), true

	default:
		return "", false
	}
}

// checkExtensionTagKeys refuses a mapping whose struct tag reflect would never
// read back. Such a tag still compiles, so without the check it would be
// silently invisible to the consumers --extension-tag exists for.
//
// A key must be one reflect can parse at all, and it must not repeat a key the
// tag already carries, from tags or from another mapping: Lookup returns the
// first entry for a key and never a second, so x-dimension=json would sit
// unread behind the field's own json tag.
func checkExtensionTagKeys(extensionTags map[string]string, tags []string) error {
	emittedBy := make(map[string]string, len(tags)+len(extensionTags))
	for _, tag := range tags {
		emittedBy[tag] = "--tags"
	}

	for _, ext := range sortedKeys(extensionTags) {
		key := extensionTags[ext]

		if !isStructTagKey(key) {
			return fmt.Errorf("%w: --extension-tag %s=%q", errInvalidExtensionTagKey, ext, key)
		}

		if owner, taken := emittedBy[key]; taken {
			return fmt.Errorf("%w: --extension-tag %s=%s repeats a key already emitted by %s",
				errDuplicateExtensionTagKey, ext, key, owner)
		}

		emittedBy[key] = "--extension-tag " + ext + "=" + key
	}

	return nil
}

// isStructTagKey reports whether key is a struct tag key reflect can read:
// non-empty, with no space, quote, colon or control character. A backtick is
// fine — tags carrying one are emitted as interpreted string literals.
func isStructTagKey(key string) bool {
	if key == "" {
		return false
	}

	for _, r := range key {
		if r <= ' ' || r == '"' || r == ':' || r == 0x7f {
			return false
		}
	}

	return true
}
