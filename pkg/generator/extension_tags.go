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
// floor — `x-measurement: "Volume_Flowrate"` says what a number means, and a
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

	tags := make([]string, 0, len(extNames))

	for _, extName := range extNames {
		raw, ok := prop.Extensions[extName]
		if !ok {
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
		tags = append(tags, g.config.ExtensionTags[extName]+":"+strconv.Quote(value))
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
