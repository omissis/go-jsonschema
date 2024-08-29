package generator

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
}

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
}
