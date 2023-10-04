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
}

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
}

type Generator struct {
	caser                 *text.Caser
	config                Config
	inScope               map[qualifiedDefinition]struct{}
	outputs               map[string]*output
	schemaCacheByFileName map[string]*schemas.Schema
	warner                func(string)
	formatters            []formatter
	fileLoader            schemas.Loader
}

const (
	varNamePlainStruct = "plain"
	varNameRawMap      = "raw"
	interfaceTypeName  = "interface{}"
)

var (
	errSchemaHasNoRoot                = errors.New("schema has no root")
	errArrayPropertyItems             = errors.New("array property must have 'items' set to a type")
	errEnumArrCannotBeEmpty           = errors.New("enum array cannot be empty")
	errEnumNonPrimitiveVal            = errors.New("enum has non-primitive value")
	errMapURIToPackageName            = errors.New("unable to map schema URI to Go package name")
	errExpectedNamedType              = errors.New("expected named type")
	errUnsupportedRefFormat           = errors.New("unsupported $ref format")
	errConflictSameFile               = errors.New("conflict: same file")
	errDefinitionDoesNotExistInSchema = errors.New("definition does not exist in schema")
	errCannotGenerateReferencedType   = errors.New("cannot generate referenced type")
)

func New(config Config) (*Generator, error) {
	formatters := []formatter{
		&jsonFormatter{},
	}
	if config.ExtraImports {
		formatters = append(formatters, &yamlFormatter{})
	}

	generator := &Generator{
		caser:                 text.NewCaser(config.Capitalizations, config.ResolveExtensions),
		config:                config,
		inScope:               map[qualifiedDefinition]struct{}{},
		outputs:               map[string]*output{},
		schemaCacheByFileName: map[string]*schemas.Schema{},
		warner:                config.Warner,
		formatters:            formatters,
	}

	generator.fileLoader = schemas.NewCachedLoader(
		schemas.NewFileLoader(config.ResolveExtensions, config.YAMLExtensions),
		generator.schemaCacheByFileName,
	)

	return generator, nil
}

func (g *Generator) Sources() map[string][]byte {
	var maxLineLength uint = 80

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
		schema, err = g.fileLoader.Load(fileName, "")
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

	return (&schemaGenerator{
		Generator:      g,
		schema:         schema,
		schemaFileName: fileName,
		output:         o,
	}).generateRootType()
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
	outputName, packageName string,
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
	if strings.ContainsAny(typeName[len(typeName)-1:], "0123456789") {
		return typeName + "_" + g.caser.Identifierize(value)
	}

	return typeName + g.caser.Identifierize(value)
}

type schemaGenerator struct {
	*Generator
	output         *output
	schema         *schemas.Schema
	schemaFileName string
}

func (g *schemaGenerator) generateRootType() error {
	if g.schema.ObjectAsType == nil {
		return errSchemaHasNoRoot
	}

	for _, name := range sortDefinitionsByName(g.schema.Definitions) {
		def := g.schema.Definitions[name]

		_, err := g.generateDeclaredType(def, newNameScope(g.caser.Identifierize(name)))
		if err != nil {
			return err
		}
	}

	if len(g.schema.ObjectAsType.Type) == 0 {
		return nil
	}

	rootTypeName := g.getRootTypeName(g.schema, g.schemaFileName)
	if _, ok := g.output.declsByName[rootTypeName]; ok {
		return nil
	}

	_, err := g.generateDeclaredType((*schemas.Type)(g.schema.ObjectAsType), newNameScope(rootTypeName))

	return err
}

func (g *schemaGenerator) generateReferencedType(ref string) (codegen.Type, error) {
	refType, err := schemas.GetRefType(ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCannotGenerateReferencedType, err)
	}

	if refType != schemas.RefTypeFile {
		return nil, fmt.Errorf("%w: %w '%s'", errCannotGenerateReferencedType, errUnsupportedRefFormat, ref)
	}

	ref = strings.TrimPrefix(ref, "file://")

	fileName := ref

	var scope, defName string

	if i := strings.IndexRune(ref, '#'); i != -1 {
		var prefix string

		fileName, scope = ref[0:i], ref[i+1:]
		lowercaseScope := strings.ToLower(scope)

		for _, currentPrefix := range []string{
			"/$defs/",       // Draft-handrews-json-schema-validation-02.
			"/definitions/", // Legacy.
		} {
			if strings.HasPrefix(lowercaseScope, currentPrefix) {
				prefix = currentPrefix

				break
			}
		}

		if len(prefix) == 0 {
			return nil, fmt.Errorf(
				"%w: value must point to definition within file: '%s'",
				errCannotGenerateReferencedType,
				ref,
			)
		}

		defName = scope[len(prefix):]
	}

	schema := g.schema
	sg := g

	if fileName != "" {
		var serr error

		schema, serr = g.fileLoader.Load(fileName, g.schemaFileName)
		if serr != nil {
			return nil, fmt.Errorf("could not follow $ref %q to file %q: %w", ref, fileName, serr)
		}

		qualified, qerr := schemas.QualifiedFileName(fileName, g.schemaFileName, g.config.ResolveExtensions)
		if qerr != nil {
			return nil, fmt.Errorf("could not resolve qualified file name for %s: %w", fileName, qerr)
		}

		if ferr := g.addFile(qualified, schema); ferr != nil {
			return nil, ferr
		}

		output, oerr := g.findOutputFileForSchemaID(schema.ID)
		if oerr != nil {
			return nil, oerr
		}

		sg = &schemaGenerator{
			Generator:      g.Generator,
			schema:         schema,
			schemaFileName: fileName,
			output:         output,
		}
	}

	qual := qualifiedDefinition{
		schema: schema,
		name:   defName,
	}

	var def *schemas.Type

	if defName != "" {
		// TODO: Support nested definitions.
		var ok bool

		def, ok = schema.Definitions[defName]
		if !ok {
			return nil, fmt.Errorf("%w: %q (from ref %q)", errDefinitionDoesNotExistInSchema, defName, ref)
		}

		if len(def.Type) == 0 && len(def.Properties) == 0 {
			return &codegen.EmptyInterfaceType{}, nil
		}

		defName = g.caser.Identifierize(defName)
	} else {
		def = (*schemas.Type)(schema.ObjectAsType)
		defName = g.getRootTypeName(schema, fileName)
		if len(def.Type) == 0 {
			// Minor hack to make definitions default to being objects.
			def.Type = schemas.TypeList{schemas.TypeNameObject}
		}
	}

	_, isCycle := g.inScope[qual]
	if !isCycle {
		g.inScope[qual] = struct{}{}
		defer func() {
			delete(g.inScope, qual)
		}()
	}

	t, err := sg.generateDeclaredType(def, newNameScope(defName))
	if err != nil {
		return nil, err
	}

	nt, ok := t.(*codegen.NamedType)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", errExpectedNamedType, t)
	}

	if isCycle {
		g.warner(fmt.Sprintf("Cycle detected; must wrap type %s in pointer", nt.Decl.Name))

		t = codegen.WrapTypeInPointer(t)
	}

	if sg.output.file.Package.QualifiedName == g.output.file.Package.QualifiedName {
		return t, nil
	}

	var imp *codegen.Import

	for _, i := range g.output.file.Package.Imports {
		i := i
		if i.Name == sg.output.file.Package.Name() && i.QualifiedName == sg.output.file.Package.QualifiedName {
			imp = &i

			break
		}
	}

	if imp == nil {
		g.output.file.Package.AddImport(sg.output.file.Package.QualifiedName, sg.output.file.Package.Name())
	}

	return &codegen.NamedType{
		Package: &sg.output.file.Package,
		Decl:    nt.Decl,
	}, nil
}

func (g *schemaGenerator) generateDeclaredType(
	t *schemas.Type, scope nameScope,
) (codegen.Type, error) {
	if decl, ok := g.output.declsBySchema[t]; ok {
		return &codegen.NamedType{Decl: decl}, nil
	}

	if t.Enum != nil {
		return g.generateEnumType(t, scope)
	}

	decl := codegen.TypeDecl{
		Name:    g.output.uniqueTypeName(scope.string()),
		Comment: t.Description,
	}
	g.output.declsBySchema[t] = &decl
	g.output.declsByName[decl.Name] = &decl

	theType, err := g.generateType(t, scope)
	if err != nil {
		return nil, err
	}

	if isNamedType(theType) {
		// Don't declare named types under a new name.
		delete(g.output.declsBySchema, t)
		delete(g.output.declsByName, decl.Name)

		return theType, nil
	}

	decl.Type = theType

	g.output.file.Package.AddDecl(&decl)

	if g.config.OnlyModels {
		return &codegen.NamedType{Decl: &decl}, nil
	}

	if structType, ok := theType.(*codegen.StructType); ok {
		var validators []validator
		for _, f := range structType.RequiredJSONFields {
			validators = append(validators, &requiredValidator{f, decl.Name})
		}

		for _, f := range structType.Fields {
			if f.DefaultValue != nil {
				validators = append(validators, &defaultValidator{
					jsonName:         f.JSONName,
					fieldName:        f.Name,
					defaultValueType: f.Type,
					defaultValue:     f.DefaultValue,
				})
			}

			validators = g.structFieldValidators(validators, f, f.Type, false)
		}

		if len(validators) > 0 {
			for _, v := range validators {
				if v.desc().hasError {
					g.output.file.Package.AddImport("fmt", "")

					break
				}
			}

			for _, formatter := range g.formatters {
				formatter := formatter

				formatter.addImport(g.output.file)

				g.output.file.Package.AddDecl(&codegen.Method{
					Impl: formatter.generate(decl, validators),
				})
			}
		}
	}

	return &codegen.NamedType{Decl: &decl}, nil
}

func (g *schemaGenerator) structFieldValidators(
	validators []validator,
	f codegen.StructField,
	t codegen.Type,
	isNillable bool,
) []validator {
	switch v := t.(type) {
	case codegen.NullType:
		validators = append(validators, &nullTypeValidator{
			fieldName: f.Name,
			jsonName:  f.JSONName,
		})

	case *codegen.PointerType:
		validators = g.structFieldValidators(validators, f, v.Type, v.IsNillable())

	case codegen.PrimitiveType:
		if v.Type == schemas.TypeNameString {
			if f.SchemaType.MinLength != 0 || f.SchemaType.MaxLength != 0 {
				validators = append(validators, &stringValidator{
					jsonName:   f.JSONName,
					fieldName:  f.Name,
					minLength:  f.SchemaType.MinLength,
					maxLength:  f.SchemaType.MaxLength,
					isNillable: isNillable,
				})
			}
		}

	case *codegen.ArrayType:
		arrayDepth := 0
		for v, ok := t.(*codegen.ArrayType); ok; v, ok = t.(*codegen.ArrayType) {
			arrayDepth++
			if _, ok := v.Type.(codegen.NullType); ok {
				validators = append(validators, &nullTypeValidator{
					fieldName:  f.Name,
					jsonName:   f.JSONName,
					arrayDepth: arrayDepth,
				})

				break
			} else if f.SchemaType.MinItems != 0 || f.SchemaType.MaxItems != 0 {
				validators = append(validators, &arrayValidator{
					fieldName:  f.Name,
					jsonName:   f.JSONName,
					arrayDepth: arrayDepth,
					minItems:   f.SchemaType.MinItems,
					maxItems:   f.SchemaType.MaxItems,
				})
			}

			t = v.Type
		}
	}

	return validators
}

func (g *schemaGenerator) generateType(
	t *schemas.Type, scope nameScope,
) (codegen.Type, error) {
	typeIndex := 0

	var typeShouldBePointer bool

	two := 2

	if ext := t.GoJSONSchemaExtension; ext != nil {
		for _, pkg := range ext.Imports {
			g.output.file.Package.AddImport(pkg, "")
		}

		if ext.Type != nil {
			return &codegen.CustomNameType{Type: *ext.Type}, nil
		}
	}

	if t.Enum != nil {
		return g.generateEnumType(t, scope)
	}

	if t.Ref != "" {
		return g.generateReferencedType(t.Ref)
	}

	if len(t.Type) == 0 {
		return codegen.EmptyInterfaceType{}, nil
	}

	if len(t.Type) == two {
		for i, t := range t.Type {
			if t == "null" {
				typeShouldBePointer = true

				continue
			}

			typeIndex = i
		}
	} else if len(t.Type) != 1 {
		// TODO: Support validation for properties with multiple types.
		g.warner("Property has multiple types; will be represented as interface{} with no validation")

		return codegen.EmptyInterfaceType{}, nil
	}

	switch t.Type[typeIndex] {
	case schemas.TypeNameArray:
		if t.Items == nil {
			return nil, errArrayPropertyItems
		}

		elemType, err := g.generateType(t.Items, scope.add("Elem"))
		if err != nil {
			return nil, err
		}

		return codegen.ArrayType{Type: elemType}, nil

	case schemas.TypeNameObject:
		return g.generateStructType(t, scope)

	case schemas.TypeNameNull:
		return codegen.EmptyInterfaceType{}, nil

	default:
		cg, err := codegen.PrimitiveTypeFromJSONSchemaType(t.Type[typeIndex], t.Format, typeShouldBePointer)
		if err != nil {
			return nil, fmt.Errorf("invalid type %q: %w", t.Type[typeIndex], err)
		}

		if ncg, ok := cg.(codegen.NamedType); ok {
			for _, imprt := range ncg.Package.Imports {
				g.output.file.Package.AddImport(imprt.QualifiedName, "")
			}

			return ncg, nil
		}

		return cg, nil
	}
}

func (g *schemaGenerator) generateStructType(
	t *schemas.Type,
	scope nameScope,
) (codegen.Type, error) {
	if len(t.Properties) == 0 {
		if len(t.Required) > 0 {
			g.warner("Object type with no properties has required fields; " +
				"skipping validation code for them since we don't know their types")
		}

		valueType := codegen.Type(codegen.EmptyInterfaceType{})

		var err error

		if t.AdditionalProperties != nil {
			if valueType, err = g.generateType(t.AdditionalProperties, nil); err != nil {
				return nil, err
			}
		}

		return &codegen.MapType{
			KeyType:   codegen.PrimitiveType{Type: schemas.TypeNameString},
			ValueType: valueType,
		}, nil
	}

	requiredNames := make(map[string]bool, len(t.Properties))
	for _, r := range t.Required {
		requiredNames[r] = true
	}

	uniqueNames := make(map[string]int, len(t.Properties))

	var structType codegen.StructType

	for _, name := range sortPropertiesByName(t.Properties) {
		prop := t.Properties[name]
		isRequired := requiredNames[name]

		fieldName := g.caser.Identifierize(name)

		if ext := prop.GoJSONSchemaExtension; ext != nil {
			for _, pkg := range ext.Imports {
				g.output.file.Package.AddImport(pkg, "")
			}

			if ext.Identifier != nil {
				fieldName = *ext.Identifier
			}
		}

		if count, ok := uniqueNames[fieldName]; ok {
			uniqueNames[fieldName] = count + 1
			fieldName = fmt.Sprintf("%s_%d", fieldName, count+1)
			g.warner(fmt.Sprintf("Field %q maps to a field by the same name declared "+
				"in the same struct; it will be declared as %s", name, fieldName))
		} else {
			uniqueNames[fieldName] = 1
		}

		structField := codegen.StructField{
			Name:       fieldName,
			Comment:    prop.Description,
			JSONName:   name,
			SchemaType: prop,
		}

		tags := ""

		if isRequired {
			for _, tag := range g.config.Tags {
				tags += fmt.Sprintf(`%s:"%s" `, tag, name)
			}
		} else {
			for _, tag := range g.config.Tags {
				tags += fmt.Sprintf(`%s:"%s,omitempty" `, tag, name)
			}
		}

		structField.Tags = strings.TrimSpace(tags)

		if structField.Comment == "" {
			structField.Comment = fmt.Sprintf("%s corresponds to the JSON schema field %q.",
				structField.Name, name)
		}

		var err error

		structField.Type, err = g.generateTypeInline(prop, scope.add(structField.Name))
		if err != nil {
			return nil, fmt.Errorf("could not generate type for field %q: %w", name, err)
		}

		switch {
		case prop.Default != nil:
			structField.DefaultValue = g.defaultPropertyValue(prop)

		default:
			if isRequired {
				structType.RequiredJSONFields = append(structType.RequiredJSONFields, structField.JSONName)
			} else if !structField.Type.IsNillable() {
				structField.Type = codegen.WrapTypeInPointer(structField.Type)
			}
		}

		structType.AddField(structField)
	}

	return &structType, nil
}

func (g *schemaGenerator) defaultPropertyValue(prop *schemas.Type) any {
	if prop.AdditionalProperties != nil {
		if len(prop.AdditionalProperties.Type) == 0 {
			return map[string]any{}
		}

		if len(prop.AdditionalProperties.Type) != 1 {
			g.warner("Additional property has multiple types; will be represented as an empty interface with no validation")

			return map[string]any{}
		}

		switch prop.AdditionalProperties.Type[0] {
		case schemas.TypeNameString:
			return map[string]string{}

		case schemas.TypeNameArray:
			return map[string][]any{}

		case schemas.TypeNameNumber:
			return map[string]float64{}

		case schemas.TypeNameInteger:
			return map[string]int{}

		case schemas.TypeNameBoolean:
			return map[string]bool{}

		default:
			return map[string]any{}
		}
	}

	return prop.Default
}

func (g *schemaGenerator) generateTypeInline(
	t *schemas.Type,
	scope nameScope,
) (codegen.Type, error) {
	two := 2

	if t.Enum == nil && t.Ref == "" {
		if ext := t.GoJSONSchemaExtension; ext != nil {
			for _, pkg := range ext.Imports {
				g.output.file.Package.AddImport(pkg, "")
			}

			if ext.Type != nil {
				return &codegen.CustomNameType{Type: *ext.Type}, nil
			}
		}

		typeIndex := 0

		var typeShouldBePointer bool

		if len(t.Type) == two {
			for i, t := range t.Type {
				if t == "null" {
					typeShouldBePointer = true

					continue
				}

				typeIndex = i
			}
		} else if len(t.Type) > 1 {
			g.warner("Property has multiple types; will be represented as interface{} with no validation")

			return codegen.EmptyInterfaceType{}, nil
		}

		if len(t.Type) == 0 {
			return codegen.EmptyInterfaceType{}, nil
		}

		if schemas.IsPrimitiveType(t.Type[typeIndex]) {
			cg, err := codegen.PrimitiveTypeFromJSONSchemaType(t.Type[typeIndex], t.Format, typeShouldBePointer)
			if err != nil {
				return nil, fmt.Errorf("invalid type %q: %w", t.Type[typeIndex], err)
			}

			if ncg, ok := cg.(codegen.NamedType); ok {
				for _, imprt := range ncg.Package.Imports {
					g.output.file.Package.AddImport(imprt.QualifiedName, "")
				}

				return ncg, nil
			}

			return cg, nil
		}

		if t.Type[typeIndex] == schemas.TypeNameArray {
			var theType codegen.Type

			if t.Items == nil {
				theType = codegen.EmptyInterfaceType{}
			} else {
				var err error

				theType, err = g.generateTypeInline(t.Items, scope.add("Elem"))
				if err != nil {
					return nil, err
				}
			}

			return &codegen.ArrayType{Type: theType}, nil
		}
	}

	return g.generateDeclaredType(t, scope)
}

func (g *schemaGenerator) generateEnumType(
	t *schemas.Type, scope nameScope,
) (codegen.Type, error) {
	if len(t.Enum) == 0 {
		return nil, errEnumArrCannotBeEmpty
	}

	var wrapInStruct bool

	var enumType codegen.Type

	if len(t.Type) == 1 {
		var err error
		if enumType, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Type[0], t.Format, false); err != nil {
			return nil, fmt.Errorf("invalid type %q: %w", t.Type[0], err)
		}

		// Enforce integer type for enum values.
		if t.Type[0] == "integer" {
			for i, v := range t.Enum {
				switch v := v.(type) {
				case float64:
					t.Enum[i] = int(v)

				default:
					return nil, fmt.Errorf("%w %v", errEnumNonPrimitiveVal, v)
				}
			}
		}

		wrapInStruct = t.Type[0] == schemas.TypeNameNull // Null uses interface{}, which cannot have methods.
	} else {
		if len(t.Type) > 1 {
			// TODO: Support multiple types.
			g.warner("Enum defined with multiple types; ignoring it and using enum values instead")
		}

		var primitiveType string
		for _, v := range t.Enum {
			var valueType string
			if v == nil {
				valueType = interfaceTypeName
			} else {
				switch v.(type) {
				case string:
					valueType = "string"
				case float64:
					valueType = "float64"
				case bool:
					valueType = "bool"
				default:
					return nil, fmt.Errorf("%w %v", errEnumNonPrimitiveVal, v)
				}
			}
			if primitiveType == "" {
				primitiveType = valueType
			} else if primitiveType != valueType {
				primitiveType = interfaceTypeName

				break
			}
		}
		if primitiveType == interfaceTypeName {
			wrapInStruct = true
		}
		enumType = codegen.PrimitiveType{Type: primitiveType}
	}

	if wrapInStruct {
		g.warner("Enum field wrapped in struct in order to store values of multiple types")

		enumType = &codegen.StructType{
			Fields: []codegen.StructField{
				{
					Name: "Value",
					Type: enumType,
				},
			},
		}
	}

	enumDecl := codegen.TypeDecl{
		Name: g.output.uniqueTypeName(scope.string()),
		Type: enumType,
	}
	g.output.file.Package.AddDecl(&enumDecl)

	g.output.declsByName[enumDecl.Name] = &enumDecl
	g.output.declsBySchema[t] = &enumDecl

	if !g.config.OnlyModels {
		valueConstant := &codegen.Var{
			Name:  "enumValues_" + enumDecl.Name,
			Value: t.Enum,
		}
		g.output.file.Package.AddDecl(valueConstant)

		g.output.file.Package.AddImport("fmt", "")
		g.output.file.Package.AddImport("reflect", "")

		for _, formatter := range g.formatters {
			formatter := formatter

			formatter.addImport(g.output.file)

			if wrapInStruct {
				g.output.file.Package.AddDecl(&codegen.Method{
					Impl: formatter.enumMarshal(enumDecl),
				})
			}

			g.output.file.Package.AddDecl(&codegen.Method{
				Impl: formatter.enumUnmarshal(enumDecl, enumType, valueConstant, wrapInStruct),
			})
		}
	}

	// TODO: May be aliased string type.
	if prim, ok := enumType.(codegen.PrimitiveType); ok && prim.Type == "string" {
		for _, v := range t.Enum {
			if s, ok := v.(string); ok {
				// TODO: Make sure the name is unique across scope.
				g.output.file.Package.AddDecl(&codegen.Constant{
					Name:  g.makeEnumConstantName(enumDecl.Name, s),
					Type:  &codegen.NamedType{Decl: &enumDecl},
					Value: s,
				})
			}
		}
	}

	return &codegen.NamedType{Decl: &enumDecl}, nil
}

type output struct {
	file          *codegen.File
	declsByName   map[string]*codegen.TypeDecl
	declsBySchema map[*schemas.Type]*codegen.TypeDecl
	warner        func(string)
}

func (o *output) uniqueTypeName(name string) string {
	v, ok := o.declsByName[name]

	if !ok || (ok && v.Type == nil) {
		return name
	}

	count := 1

	for {
		suffixed := fmt.Sprintf("%s_%d", name, count)
		if _, ok := o.declsByName[suffixed]; !ok {
			o.warner(fmt.Sprintf(
				"Multiple types map to the name %q; declaring duplicate as %q instead", name, suffixed))

			return suffixed
		}
		count++
	}
}

type qualifiedDefinition struct {
	schema *schemas.Schema
	name   string
}

type nameScope []string

func newNameScope(s string) nameScope {
	return nameScope{s}
}

func (ns nameScope) string() string {
	return strings.Join(ns, "")
}

func (ns nameScope) add(s string) nameScope {
	result := make(nameScope, len(ns)+1)
	copy(result, ns)
	result[len(result)-1] = s

	return result
}
