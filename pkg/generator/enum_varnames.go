package generator

import (
	"fmt"
	"unicode"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// enumVarnames resolves the `x-enum-varnames` extension into one constant name
// per enum value, or returns nil to fall back to the derived names.
//
// The extension exists because a schema shared between generators otherwise
// yields a different Go identifier per generator for the same enum member:
// oapi-codegen emits `Created` where this generator derives
// `JobStatusStatusCreated`, so a codebase reading both ends up with two names
// for one value.
//
// A varname is the COMPLETE constant name, not a suffix appended to the type
// name. That is what oapi-codegen does, and schemas in the wild rely on it —
// they hand-prefix entries (`JobRunningStatusRunning`) precisely to
// disambiguate members that would otherwise collide across sibling enums.
//
// The result is index-aligned with t.Enum. An empty entry means "derive this
// one", so a schema can name only the members that need it.
func (g *schemaGenerator) enumVarnames(t *schemas.Type, declName string) []string {
	if len(t.XEnumVarnames) == 0 {
		return nil
	}

	// A length mismatch cannot be resolved positionally without guessing
	// which values were meant to be skipped, and guessing here would
	// silently misname constants. Refuse the whole extension instead.
	if len(t.XEnumVarnames) != len(t.Enum) {
		g.warner(fmt.Sprintf(
			"Enum %s declares %d x-enum-varnames for %d enum values; ignoring the extension",
			declName, len(t.XEnumVarnames), len(t.Enum),
		))

		return nil
	}

	names := make([]string, len(t.XEnumVarnames))
	seen := make(map[string]int, len(t.XEnumVarnames))

	for i, raw := range t.XEnumVarnames {
		if raw == "" {
			continue
		}

		// Checked before identifierizing, not after: Identifierize falls
		// back to the literal "Undefined" for input it cannot use, which
		// would otherwise sail through as a real — and collision-prone —
		// constant name.
		if !hasIdentifierChar(raw) {
			g.warner(fmt.Sprintf(
				"Enum %s: x-enum-varnames[%d] %q yields no usable Go identifier; deriving that constant's name instead",
				declName, i, raw,
			))

			continue
		}

		// Identifierize rather than trusting the input: the extension is
		// author-supplied text, and a value that is not a legal Go
		// identifier would otherwise emit code that does not compile. It
		// is a no-op for names that are already identifiers.
		name := g.caser.Identifierize(raw)

		if prev, dup := seen[name]; dup {
			g.warner(fmt.Sprintf(
				"Enum %s: x-enum-varnames[%d] %q collides with entry %d; deriving that constant's name instead",
				declName, i, raw, prev,
			))

			continue
		}

		seen[name] = i
		names[i] = name
	}

	return names
}

// enumConstantName picks the constant name for the enum value at index i,
// preferring an x-enum-varname when one was supplied and usable.
func (g *schemaGenerator) enumConstantName(varnames []string, declName string, i int, value string) string {
	if i < len(varnames) && varnames[i] != "" {
		return varnames[i]
	}

	return g.makeEnumConstantName(declName, value)
}

// hasIdentifierChar reports whether s carries anything an identifier could be
// built from.
func hasIdentifierChar(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}

	return false
}
