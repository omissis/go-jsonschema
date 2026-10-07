package generator

import (
	"errors"
	"fmt"
	"go/format"
	"go/token"
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
	errEnumArrCannotBeEmpty           = errors.New("enum array cannot be empty")
	errEnumNonPrimitiveVal            = errors.New("enum has non-primitive value")
	errRootConstNotInEnum             = errors.New("root const is not one of its enum values, so no value matches")
	errMapURIToPackageName            = errors.New("unable to map schema URI to Go package name")
	errExpectedNamedType              = errors.New("expected named type")
	errCannotResolveRef               = errors.New("cannot resolve reference")
	errConflictSameFile               = errors.New("conflict: same file")
	errDefinitionDoesNotExistInSchema = errors.New("definition does not exist in schema")
	errCannotGenerateReferencedType   = errors.New("cannot generate referenced type")
	errCannotGenerateSources          = errors.New("cannot generate sources")
	errInvalidExtensionTagKey         = errors.New("invalid struct tag key")
	errDuplicateExtensionTagKey       = errors.New("duplicate struct tag key")
)

type Generator struct {
	caser        *text.Caser
	config       Config
	inScope      map[qualifiedDefinition]struct{}
	outputs      map[string]*output
	warner       func(string)
	formatters   []formatter
	loader       schemas.Loader
	minimalNames bool
}

type qualifiedDefinition struct {
	schema     *schemas.Schema
	schemaType *schemas.Type
	filename   string
	name       string
}

func New(config Config) (*Generator, error) {
	if err := checkExtensionTagKeys(config.ExtensionTags, config.Tags); err != nil {
		return nil, err
	}

	if !config.StrictAdditionalProperties.IsValid() {
		return nil, fmt.Errorf("%w: got %q",
			ErrInvalidStrictAdditionalPropertiesMode, config.StrictAdditionalProperties)
	}

	// aliasByPackage tracks the alias each PackageName has been bound to
	// so far. Two SchemaMappings with the same PackageName but different
	// ImportAlias values would have resolveImportAlias silently picking
	// one and ignoring the other — reject at New() time so the
	// inconsistency surfaces loudly.
	aliasByPackage := make(map[string]string, len(config.SchemaMappings))

	for _, m := range config.SchemaMappings {
		if m.ImportAlias == "" {
			continue
		}

		if !IsValidImportAlias(m.ImportAlias) {
			return nil, fmt.Errorf("%w: schema %q -> %q",
				ErrInvalidImportAlias, m.SchemaID, m.ImportAlias)
		}

		if prev, exists := aliasByPackage[m.PackageName]; exists && prev != m.ImportAlias {
			return nil, fmt.Errorf("%w: package %q has conflicting aliases %q and %q",
				ErrConflictingImportAlias, m.PackageName, prev, m.ImportAlias)
		}

		aliasByPackage[m.PackageName] = m.ImportAlias
	}

	formatters := []formatter{
		&jsonFormatter{},
	}
	if config.ExtraImports {
		formatters = append(formatters, &yamlFormatter{})
	}

	generator := &Generator{
		caser:        text.NewCaser(config.Capitalizations, config.ResolveExtensions),
		config:       config,
		inScope:      map[qualifiedDefinition]struct{}{},
		outputs:      map[string]*output{},
		warner:       config.Warner,
		formatters:   formatters,
		loader:       config.Loader,
		minimalNames: config.MinimalNames,
	}

	if config.Loader == nil {
		// When the caller supplied Cache, build the default chain by hand and
		// hand the populated map to NewCachedLoader so cache hits short-circuit
		// before any FileLoader / HTTPLoader work.
		if config.Cache != nil {
			generator.loader = schemas.NewCachedLoader(
				schemas.NewDefaultMultiLoader(config.ResolveExtensions, config.YAMLExtensions),
				config.Cache,
			)
		} else {
			generator.loader = schemas.NewDefaultCacheLoader(config.ResolveExtensions, config.YAMLExtensions)
		}
	}

	return generator, nil
}

// resolveImportAlias returns the alias to use for an `import` statement that
// references the given Go package import path. When a SchemaMapping with a
// non-empty ImportAlias is present for the path, that alias wins; otherwise
// the historical last-path-segment derivation (codegen.Package.Name) is used.
//
// The same alias must be returned for the same qualifiedName at every call
// site (in particular, the duplicate-import check and the AddImport call in
// generateReferencedType) — otherwise the dup check mis-categorizes and two
// imports for the same package end up emitted.
func (g *Generator) resolveImportAlias(qualifiedName string) string {
	for _, m := range g.config.SchemaMappings {
		if m.PackageName == qualifiedName && m.ImportAlias != "" {
			return m.ImportAlias
		}
	}

	if i := strings.LastIndex(qualifiedName, "/"); i != -1 && i < len(qualifiedName)-1 {
		return qualifiedName[i+1:]
	}

	return qualifiedName
}

func (g *Generator) Sources() (map[string][]byte, error) {
	var maxLineLength int32 = 80

	sources := make(map[string]*strings.Builder, len(g.outputs))

	for _, output := range g.outputs {
		if output.file.FileName == "" {
			continue
		}

		g.renameConstantsShadowingImports(output)

		if a, b, name, clash := importNameCollision(output.file.Package.Imports); clash {
			return nil, fmt.Errorf(
				"%w: %q and %q as %q in %s; an ImportAlias (--schema-package URL=PATH:ALIAS) tells schema packages apart",
				ErrImportAliasCollision, a, b, name, output.file.FileName,
			)
		}

		emitter := codegen.NewEmitter(maxLineLength)

		if err := output.file.Generate(emitter); err != nil {
			return nil, fmt.Errorf("%w: %w", errCannotGenerateSources, err)
		}

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

	return result, nil
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

	return g.AddFile(fileName, schema)
}

func (g *Generator) AddFile(fileName string, schema *schemas.Schema) error {
	o, err := g.findOutputFileForSchemaID(schema.ID)
	if err != nil {
		return err
	}

	if schema.ID != "" {
		if _, processed := o.processedSchemas[schema.ID]; processed {
			return nil
		}

		o.processedSchemas[schema.ID] = true
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
				errConflictSameFile, o.file.FileName, o.file.Package.QualifiedName, packageName, id,
			)
		}

		if o.file.FileName == outputName && o.file.Package.QualifiedName == packageName {
			return o, nil
		}
	}

	pkg := codegen.Package{
		QualifiedName: packageName,
	}

	output := &output{
		minimalNames: g.minimalNames,
		warner:       g.warner,
		file: &codegen.File{
			FileName: outputName,
			Package:  pkg,
		},
		declsBySchema:           map[*schemas.Type]*codegen.TypeDecl{},
		declsByName:             map[string]*codegen.TypeDecl{},
		unmarshallersByTypeDecl: map[*codegen.TypeDecl]bool{},
		processedSchemas:        map[string]bool{},
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

// renameConstantsShadowingImports renames a package-level constant that shares
// its name with an import, which Go rejects. enumVarnames refuses a varname
// naming a package already imported, but an import can arrive after the
// constant — a later type adding mapstructure, say — so the final set is
// checked here. Generated code never refers to an enum constant by name, so
// renaming one is safe; the warning says what the user-facing name became.
func (g *Generator) renameConstantsShadowingImports(o *output) {
	for _, decl := range o.file.Package.Decls {
		c, ok := decl.(*codegen.Constant)
		if !ok || !o.importsName(c.Name) {
			continue
		}

		for n := 1; ; n++ {
			candidate := fmt.Sprintf("%s_%d", c.Name, n)
			if o.importsName(candidate) || o.declaresName(candidate) {
				continue
			}

			g.config.Warner(fmt.Sprintf(
				"Constant %q shares its name with an imported package; declaring it as %q instead",
				c.Name, candidate,
			))

			c.Name = candidate

			break
		}
	}
}

// IsValidImportAlias reports whether alias can name an import in generated
// code: a Go identifier, other than a keyword, the blank identifier `_` (a
// blank import cannot be referred to) and `init` (which Go reserves for
// functions).
func IsValidImportAlias(alias string) bool {
	return token.IsIdentifier(alias) && !token.IsKeyword(alias) && alias != "_" && alias != "init"
}

// importNameCollision returns two import paths that Go would bind to the same
// name in one file, and that name. It runs once a file's imports are complete,
// since a clash can come from any of them: two schema packages, aliased or
// not, or a schema package aliased like a standard one such as `json`.
func importNameCollision(imports []codegen.Import) (string, string, string, bool) {
	seen := make(map[string]string, len(imports))

	for _, i := range imports {
		name := effectiveImportName(i)
		if name == "_" || name == "." {
			continue
		}

		if prev, ok := seen[name]; ok && prev != i.QualifiedName {
			return prev, i.QualifiedName, name, true
		}

		seen[name] = i.QualifiedName
	}

	return "", "", "", false
}

// effectiveImportName returns the name an import binds: its alias, or else the
// package name Go conventionally derives from its path, the last element without
// a major-version suffix (`.../mapstructure/v2` is `mapstructure`) or a gopkg.in
// `.vN` (`gopkg.in/yaml.v3` is `yaml`).
func effectiveImportName(i codegen.Import) string {
	if i.Name != "" {
		return i.Name
	}

	parts := strings.Split(i.QualifiedName, "/")
	last := parts[len(parts)-1]

	if len(parts) > 1 && isMajorVersion(last) {
		last = parts[len(parts)-2]
	}

	if j := strings.Index(last, ".v"); j > 0 && isMajorVersion(last[j+1:]) {
		last = last[:j]
	}

	return last
}

// isMajorVersion reports whether s is a module major-version element like `v2`.
func isMajorVersion(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}

	for _, r := range s[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
