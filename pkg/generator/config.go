package generator

import "github.com/atombender/go-jsonschema/pkg/schemas"

type Config struct {
	SchemaMappings      []SchemaMapping
	ExtraImports        bool
	Capitalizations     []string
	ResolveExtensions   []string
	YAMLExtensions      []string
	DefaultPackageName  string
	DefaultOutputName   string
	StructNameFromTitle bool
	Warner              func(string)
	Tags                []string
	OnlyModels          bool
	MinSizedInts        bool
	Loader              schemas.Loader
	// When DisableOmitempty is set to true,
	// an "omitempty" tag will never be present in generated struct fields.
	// When DisableOmitempty is set to false,
	// an "omitempty" tag will be present for all fields that are not required.
	DisableOmitempty bool
}

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
}
