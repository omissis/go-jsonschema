package generator

import (
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
}
