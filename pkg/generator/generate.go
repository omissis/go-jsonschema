package generator

import (
	"errors"
	"fmt"
	"go/format"
	"os"
	"strings"

	"github.com/atombender/go-jsonschema/internal/x/text"
	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

const (
	varNamePlainStruct = "plain"
	varNameRawMap      = "raw"
	interfaceTypeName  = "interface{}"
	typePlain          = "Plain"
)

var (
	errSchemaHasNoRoot                = errors.New("schema has no root")
	errArrayPropertyItems             = errors.New("array property must have 'items' set to a type")
	errEnumArrCannotBeEmpty           = errors.New("enum array cannot be empty")
	errEnumNonPrimitiveVal            = errors.New("enum has non-primitive value")
	errMapURIToPackageName            = errors.New("unable to map schema URI to Go package name")
	errExpectedNamedType              = errors.New("expected named type")
	errCannotResolveRef               = errors.New("cannot resolve reference")
	errConflictSameFile               = errors.New("conflict: same file")
	errDefinitionDoesNotExistInSchema = errors.New("definition does not exist in schema")
	errCannotGenerateReferencedType   = errors.New("cannot generate referenced type")
)

type Generator struct {
	caser      *text.Caser
	config     Config
	inScope    map[qualifiedDefinition]struct{}
	outputs    map[string]*output
	warner     func(string)
	formatters []formatter
	loader     schemas.Loader
}

type qualifiedDefinition struct {
	schema     *schemas.Schema
	schemaType *schemas.Type
	filename   string
	name       string
}

func New(config Config) (*Generator, error) {
	formatters := []formatter{
		&jsonFormatter{},
	}
	if config.ExtraImports {
		formatters = append(formatters, &yamlFormatter{})
	}

	generator := &Generator{
		caser:      text.NewCaser(config.Capitalizations, config.ResolveExtensions),
		config:     config,
		inScope:    map[qualifiedDefinition]struct{}{},
		outputs:    map[string]*output{},
		warner:     config.Warner,
		formatters: formatters,
		loader:     config.Loader,
	}

	if config.Loader == nil {
		generator.loader = schemas.NewDefaultCacheLoader(config.ResolveExtensions, config.YAMLExtensions)
	}

	return generator, nil
}

func (g *Generator) Sources() map[string][]byte {
	var maxLineLength int32 = 80

	sources := make(map[string]*strings.Builder, len(g.outputs))

	for _, output := range g.outputs {
		if output.file.FileName == "" {
			continue
		}

		emitter := codegen.NewEmitter(maxLineLength)
		output.file.Generate(emitter)

		sb, ok := sources[output.file.FileName]
		if !ok {
			sb = &strings.Builder{}
			sources[output.file.FileName] = sb
		}

		_, _ = sb.WriteString(emitter.String())
	}

	result := make(map[string][]byte, len(sources))

	for f, sb := range sources {
		source := []byte(sb.String())

		src, err := format.Source(source)
		if err != nil {
			g.config.Warner(fmt.Sprintf("The generated code could not be formatted automatically; "+
				"falling back to unformatted: %s", err))

			src = source
		}

		result[f] = src
	}

	return result
}

func (g *Generator) DoFile(fileName string) error {
	var err error

	var schema *schemas.Schema

	if fileName == "-" {
		schema, err = schemas.FromJSONReader(os.Stdin)
		if err != nil {
			return fmt.Errorf("error parsing from standard input: %w", err)
		}
	} else {
		schema, err = g.loader.Load(fileName, "")
		if err != nil {
			return fmt.Errorf("error parsing from file %s: %w", fileName, err)
		}
	}

	return g.addFile(fileName, schema)
}

func (g *Generator) addFile(fileName string, schema *schemas.Schema) error {
	o, err := g.findOutputFileForSchemaID(schema.ID)
	if err != nil {
		return err
	}

	return newSchemaGenerator(g, schema, fileName, o).generateRootType()
}

func (g *Generator) getRootTypeName(schema *schemas.Schema, fileName string) string {
	for _, m := range g.config.SchemaMappings {
		if m.SchemaID == schema.ID && m.RootType != "" {
			return m.RootType
		}
	}

	if g.config.StructNameFromTitle && schema.Title != "" {
		return g.caser.Identifierize(schema.Title)
	}

	return g.caser.IdentifierFromFileName(fileName)
}

func (g *Generator) findOutputFileForSchemaID(id string) (*output, error) {
	if o, ok := g.outputs[id]; ok {
		return o, nil
	}

	for _, m := range g.config.SchemaMappings {
		if m.SchemaID == id {
			return g.beginOutput(id, m.OutputName, m.PackageName)
		}
	}

	return g.beginOutput(id, g.config.DefaultOutputName, g.config.DefaultPackageName)
}

func (g *Generator) beginOutput(
	id string,
	outputName,
	packageName string,
) (*output, error) {
	if packageName == "" {
		return nil, fmt.Errorf("%w: %q", errMapURIToPackageName, id)
	}

	for _, o := range g.outputs {
		if o.file.FileName == outputName && o.file.Package.QualifiedName != packageName {
			return nil, fmt.Errorf(
				"%w (%s) mapped to two different Go packages (%q and %q) for schema %q",
				errConflictSameFile, o.file.FileName, o.file.Package.QualifiedName, packageName, id)
		}

		if o.file.FileName == outputName && o.file.Package.QualifiedName == packageName {
			return o, nil
		}
	}

	pkg := codegen.Package{
		QualifiedName: packageName,
	}

	output := &output{
		warner: g.warner,
		file: &codegen.File{
			FileName: outputName,
			Package:  pkg,
		},
		declsBySchema: map[*schemas.Type]*codegen.TypeDecl{},
		declsByName:   map[string]*codegen.TypeDecl{},
	}
	g.outputs[id] = output

	return output, nil
}

func (g *Generator) makeEnumConstantName(typeName, value string) string {
	idv := g.caser.Identifierize(value)

	if len(typeName) == 0 {
		return "Enum" + idv
	}

	if strings.ContainsAny(typeName[len(typeName)-1:], "0123456789") {
		return typeName + "_" + idv
	}

	return typeName + idv
}

func (g *Generator) DisableOmitempty() bool {
	return g.config.DisableOmitempty
}
