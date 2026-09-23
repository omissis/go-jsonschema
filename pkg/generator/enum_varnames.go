package generator

import (
	"fmt"
	"go/token"
	"unicode"

	"github.com/atombender/go-jsonschema/pkg/codegen"
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

		// Only repair what needs repairing. Identifierize also *reshapes*
		// names it considers badly cased — `HTTP_Status` becomes
		// `HTTPStatus` — and doing that to a name the schema supplied
		// defeats the point of the extension, which is that the schema
		// decides the identifier. A name that is already a legal Go
		// identifier is therefore taken verbatim.
		name := raw
		if !token.IsIdentifier(name) {
			name = g.caser.Identifierize(raw)
		}

		// A name the package already declares would be dropped in silence:
		// Package.AddDecl dedupes by name, so a varname colliding with its
		// own enum's type (`JobStatus` on enum `JobStatus`) emits nothing
		// at all rather than failing.
		if g.nameIsDeclared(name) {
			g.warner(fmt.Sprintf(
				"Enum %s: x-enum-varnames[%d] %q is already declared in this package; "+
					"deriving that constant's name instead",
				declName, i, raw,
			))

			continue
		}

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

// nameIsDeclared reports whether the output package already declares name.
//
// Package.AddDecl dedupes by name and keeps the first declaration, so a second
// one is discarded without a word. That is tolerable for identical types but
// not for a constant the schema explicitly asked for.
func (g *schemaGenerator) nameIsDeclared(name string) bool {
	for _, decl := range g.output.file.Package.Decls {
		if named, ok := decl.(codegen.Named); ok && named.GetName() == name {
			return true
		}
	}

	return false
}
