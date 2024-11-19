package generator

import "github.com/atombender/go-jsonschema/pkg/schemas"

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
}

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
}
