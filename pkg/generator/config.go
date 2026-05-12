package generator

import (
	"errors"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/schemas"
)

type Config struct {
	SchemaMappings []SchemaMapping
	// ExtraImports allows the generator to pull imports from outside the standard library.
	ExtraImports bool
	// Capitalizations configures capitalized forms for identifiers which take precedence over the default algorithm.
	Capitalizations []string
	// ResolveExtensions configures file extensions to use when resolving schema names.
	ResolveExtensions []string
	// YAMLExtensions configures the file extensions that are recognized as YAML files.
	YAMLExtensions []string
	// DefaultPackageName configures the package to declare files under.
	DefaultPackageName string
	// DefaultOutputName configures the file to write.
	DefaultOutputName string
	// StructNameFromTitle configures the generator to use the schema title as the generated struct name.
	StructNameFromTitle bool
	// Warner provides a handler for warning messages.
	Warner func(string)
	// Tags specifies which struct tags should be generated.
	Tags []string
	// OnlyModels configures the generator to omit unmarshal methods, validations, anything but models.
	OnlyModels bool
	// MinSizedInts configures the generator to use the smallest int and uint types based on schema maximum values.
	MinSizedInts bool
	// MinimalNames configures the generator to use the shortest identifier names possible.
	MinimalNames bool
	// Loader provides a schema loader for the generator.
	Loader schemas.Loader
	// Cache pre-populates the default loader's URL → *schemas.Schema cache.
	// When non-nil and Loader is nil, New() builds the default cached loader
	// using this map so that any `$ref` to a URL listed here resolves to the
	// pre-loaded schema without an HTTP fetch (or a file lookup, depending on
	// the URL scheme). Schemas inserted this way are loaded for resolution
	// only — no Go code is emitted for them. Ignored when Loader is non-nil
	// (build your own pre-populated CachedLoader and pass it via Loader if
	// you need full control over the loader chain).
	Cache map[string]*schemas.Schema
	// When DisableOmitempty is set to true,
	// an "omitempty" tag will never be present in generated struct fields.
	// When DisableOmitempty is set to false,
	// an "omitempty" tag will be present for all fields that are not required.
	DisableOmitEmpty bool
	// When DisableOmitZero is set to true,
	// an "omitzero" tag will never be present in generated struct fields.
	// When DisableOmitZero is set to false,
	// an "omitzero" tag will be present for all fields that are not required.
	DisableOmitZero bool
	// DisableReadOnlyValidation configures the generator to omit validation for read-only fields.
	DisableReadOnlyValidation bool
	// DisableCustomTypesForMaps configures the generator to avoid creating a custom type for maps,
	// and to use the map type directly.
	DisableCustomTypesForMaps bool
	// AliasSingleAllOfAnyOfRefs will convert types with a single nested anyOf or allOf ref type into a type alias.
	AliasSingleAllOfAnyOfRefs bool
	// FormatValidation controls runtime validation of JSON Schema `format`
	// keywords (uuid, email, uri, uri-reference, hostname, regex). The zero
	// value disables all format validation, preserving previous behavior.
	FormatValidation FormatValidationConfig
	// StrictAdditionalProperties controls runtime rejection of unknown
	// fields during both `UnmarshalJSON` and `UnmarshalYAML`. The zero value
	// (off) preserves the previous behavior where `additionalProperties:
	// false` is silently ignored at unmarshal time.
	StrictAdditionalProperties StrictAdditionalPropertiesMode
}

// StrictAdditionalPropertiesMode selects how the generator enforces JSON Schema's
// `additionalProperties: false` keyword at unmarshal time. Enforcement is
// applied uniformly to JSON and YAML inputs, since both go through the same
// generated body.
type StrictAdditionalPropertiesMode string

const (
	// StrictAdditionalPropertiesOff silently drops unknown fields from JSON
	// and YAML input. This is the historical (and zero-value) behavior.
	StrictAdditionalPropertiesOff StrictAdditionalPropertiesMode = ""
	// StrictAdditionalPropertiesRespectSchema rejects unknown fields (in
	// JSON or YAML input) for objects whose schema declares
	// `additionalProperties: false`. Other objects continue to accept and
	// silently drop unknown fields.
	StrictAdditionalPropertiesRespectSchema StrictAdditionalPropertiesMode = "respect-schema"
	// StrictAdditionalPropertiesStrict rejects unknown fields (in JSON or
	// YAML input) for every generated object type, regardless of whether
	// the schema declared `additionalProperties: false`. Skipped when the
	// schema specifies a typed additionalProperties (a catch-all field is
	// generated instead).
	StrictAdditionalPropertiesStrict StrictAdditionalPropertiesMode = "strict"
)

// ErrInvalidStrictAdditionalPropertiesMode is returned when Config holds a
// StrictAdditionalProperties value outside the documented set. Without this
// check a typo (e.g. "rstrict") would silently fall through the per-type
// switches and emit strict validators in modes the documented values would
// not select.
var ErrInvalidStrictAdditionalPropertiesMode = errors.New(
	"invalid StrictAdditionalProperties mode (expected one of: \"\", \"respect-schema\", \"strict\")",
)

// ErrInvalidImportAlias is returned when a SchemaMapping carries an ImportAlias
// that isn't a valid (non-keyword) Go identifier. Caught at startup so a typo
// (e.g. "v 1") doesn't silently propagate into broken `import` statements.
var ErrInvalidImportAlias = errors.New(
	"invalid ImportAlias on SchemaMapping (must be a non-keyword Go identifier)",
)

// ErrConflictingImportAlias is returned when two SchemaMappings declare the
// same PackageName but different ImportAlias values. resolveImportAlias picks
// whichever it iterates over first, so the inconsistency would silently
// degrade to one alias winning and the other being ignored. Reject at New()
// so the conflict surfaces at the configuration boundary.
var ErrConflictingImportAlias = errors.New(
	"conflicting ImportAlias values for the same PackageName on SchemaMappings",
)

// IsValid reports whether the mode is one of the documented values.
func (m StrictAdditionalPropertiesMode) IsValid() bool {
	switch m {
	case StrictAdditionalPropertiesOff,
		StrictAdditionalPropertiesRespectSchema,
		StrictAdditionalPropertiesStrict:
		return true
	}

	return false
}

// FormatValidationConfig controls runtime validation of JSON Schema `format` keywords.
type FormatValidationConfig struct {
	// Enabled turns on format validation. When false, no validation is generated
	// regardless of AllowList.
	Enabled bool
	// AllowList restricts which formats receive runtime validation. When nil and
	// Enabled is true, every supported format keyword is validated. When non-nil,
	// only the listed format names are validated; unlisted formats fall back to
	// plain `string` with no checks.
	AllowList []string
}

// SupportedFormats returns the JSON Schema `format` keyword names that have a
// built-in runtime validator. The returned slice is a fresh copy; callers may
// sort or modify it without affecting subsequent calls.
func SupportedFormats() []string {
	return []string{
		formatKeywordUUID,
		formatKeywordEmail,
		formatKeywordURI,
		formatKeywordURIReference,
		formatKeywordHostname,
		formatKeywordRegex,
	}
}

// IsSupportedFormat reports whether the named JSON Schema `format` keyword has
// a built-in runtime validator. The check normalizes whitespace and case so
// that callers passing user input (e.g. CLI flag values) match the canonical
// keyword names.
func IsSupportedFormat(format string) bool {
	return isKnownFormatKeyword(strings.ToLower(strings.TrimSpace(format)))
}

// shouldValidate reports whether the named JSON Schema format keyword should
// receive a generated runtime validator under this configuration.
//
// AllowList entries are normalized (trimmed and lowercased) so that
// configuration variations like `"UUID"` or `" email "` match the canonical
// keyword names rather than silently disabling validation.
//
// Formats with no built-in validator return false even when AllowList is nil,
// so callers don't have to layer their own IsSupportedFormat check on top.
func (c FormatValidationConfig) shouldValidate(format string) bool {
	if !c.Enabled {
		return false
	}

	if !IsSupportedFormat(format) {
		return false
	}

	if c.AllowList == nil {
		return true
	}

	target := strings.ToLower(strings.TrimSpace(format))
	for _, entry := range c.AllowList {
		if strings.ToLower(strings.TrimSpace(entry)) == target {
			return true
		}
	}

	return false
}

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
	// ImportAlias overrides the import alias used in generated code for this
	// mapping's PackageName. When empty, the alias is derived from the last
	// path segment via Package.Name() (the historical behavior). Set this when
	// two mapped packages share a last path segment (e.g. both end in "/v1")
	// to avoid an `import v1 "..." / import v1 "..."` collision.
	ImportAlias string
}
