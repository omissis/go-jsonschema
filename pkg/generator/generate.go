package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
	"github.com/pkg/errors"
)

type Config struct {
	SchemaMappings     []SchemaMapping
	Capitalizations    []string
	ResolveExtensions  []string
	YAMLExtensions     []string
	DefaultPackageName string
	DefaultOutputName  string
	Warner             func(string)
}

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
}

type Generator struct {
	config                Config
	outputs               map[string]*output
	schemaCacheByFileName map[string]*schemas.Schema
	inScope               map[qualifiedDefinition]struct{}
	warner                func(string)
}

func New(config Config) (*Generator, error) {
	return &Generator{
		config:                config,
		outputs:               map[string]*output{},
		schemaCacheByFileName: map[string]*schemas.Schema{},
		inScope:               map[qualifiedDefinition]struct{}{},
		warner:                config.Warner,
	}, nil
}

func (g *Generator) Sources() map[string][]byte {
	sources := make(map[string]*strings.Builder, len(g.outputs))
	for _, output := range g.outputs {
		if output.file.FileName == "" {
			continue
		}
		emitter := codegen.NewEmitter(80)
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
			return errors.Wrap(err, "error parsing from standard input")
		}
	} else {
		schema, err = g.parseFile(fileName)
		if err != nil {
			return errors.Wrapf(err, "error parsing from file %s", fileName)
		}
	}
	return g.addFile(fileName, schema)
}

func (g *Generator) parseFile(fileName string) (*schemas.Schema, error) {
	// TODO: Refactor into some kind of loader
	isYAML := false
	for _, yamlExt := range g.config.YAMLExtensions {
		if strings.HasSuffix(fileName, yamlExt) {
			isYAML = true
			break
		}
	}
	if isYAML {
		return schemas.FromYAMLFile(fileName)
	} else {
		return schemas.FromJSONFile(fileName)
	}
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

func (g *Generator) loadSchemaFromFile(fileName, parentFileName string) (*schemas.Schema, error) {
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(filepath.Dir(parentFileName), fileName)
	}

	exts := append([]string{""}, g.config.ResolveExtensions...)
	for i, ext := range exts {
		qualified := fileName + ext

		// Poor man's resolving loop
		if i < len(exts)-1 && !fileExists(qualified) {
			continue
		}

		var err error
		qualified, err = filepath.EvalSymlinks(qualified)
		if err != nil {
			return nil, err
		}

		if schema, ok := g.schemaCacheByFileName[qualified]; ok {
			return schema, nil
		}

		schema, err := g.parseFile(qualified)
		if err != nil {
			return nil, err
		}
		g.schemaCacheByFileName[qualified] = schema

		if err = g.addFile(qualified, schema); err != nil {
			return nil, err
		}
		return schema, nil
	}
	return nil, fmt.Errorf("could not resolve schema %q", fileName)
}

func (g *Generator) getRootTypeName(schema *schemas.Schema, fileName string) string {
	for _, m := range g.config.SchemaMappings {
		if m.SchemaID == schema.ID && m.RootType != "" {
			return m.RootType
		}
	}
	return g.identifierFromFileName(fileName)
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
	outputName, packageName string) (*output, error) {
	if packageName == "" {
		return nil, fmt.Errorf("unable to map schema URI %q to a Go package name", id)
	}

	for _, o := range g.outputs {
		if o.file.FileName == outputName && o.file.Package.QualifiedName != packageName {
			return nil, fmt.Errorf(
				"conflict: same file (%s) mapped to two different Go packages (%q and %q) for schema %q",
				o.file.FileName, o.file.Package.QualifiedName, packageName, id)
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
		return typeName + "_" + g.identifierize(value)
	}
	return typeName + g.identifierize(value)
}

func (g *Generator) identifierFromFileName(fileName string) string {
	s := filepath.Base(fileName)
	for _, ext := range g.config.ResolveExtensions {
		trimmed := strings.TrimSuffix(s, ext)
		if trimmed == s {
			break
		}
		s = trimmed
	}
	return g.identifierize(s)
}

func (g *Generator) identifierize(s string) string {
	if s == "" {
		return "Blank"
	}

	// FIXME: Better handling of non-identifier chars
	var sb strings.Builder
	for _, part := range splitIdentifierByCaseAndSeparators(s) {
		_, _ = sb.WriteString(g.capitalize(part))
	}
	ident := sb.String()
	if !unicode.IsLetter(rune(ident[0])) {
		ident = "A" + ident
	}
	return ident
}

func (g *Generator) capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	for _, c := range g.config.Capitalizations {
		if strings.ToLower(c) == strings.ToLower(s) {
			return c
		}
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}

type schemaGenerator struct {
	*Generator
	output         *output
	schema         *schemas.Schema
	schemaFileName string
}

func (g *schemaGenerator) generateRootType() error {
	if g.schema.ObjectAsType == nil {
		return errors.New("schema has no root")
	}

	for _, name := range sortDefinitionsByName(g.schema.Definitions) {
		def := g.schema.Definitions[name]
		_, err := g.generateDeclaredType(def, newNameScope(g.identifierize(name)))
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
	var fileName, scope, defName string
	if i := strings.IndexRune(ref, '#'); i == -1 {
		fileName = ref
	} else {
		fileName, scope = ref[0:i], ref[i+1:]
		if !strings.HasPrefix(strings.ToLower(scope), "/definitions/") {
			return nil, fmt.Errorf("unsupported $ref format; must point to definition within file: %q", ref)
		}
		defName = scope[13:]
	}

	var schema *schemas.Schema
	if fileName != "" {
		var err error
		schema, err = g.loadSchemaFromFile(fileName, g.schemaFileName)
		if err != nil {
			return nil, fmt.Errorf("could not follow $ref %q to file %q: %s", ref, fileName, err)
		}
	} else {
		schema = g.schema
	}

	qual := qualifiedDefinition{
		schema: schema,
		name:   defName,
	}

	var def *schemas.Type
	if defName != "" {
		// TODO: Support nested definitions
		var ok bool
		def, ok = schema.Definitions[defName]
		if !ok {
			return nil, fmt.Errorf("definition %q (from ref %q) does not exist in schema", defName, ref)
		}
		if len(def.Type) == 0 && len(def.Properties) == 0 {
			return &codegen.EmptyInterfaceType{}, nil
		}
		defName = g.identifierize(defName)
	} else {
		def = (*schemas.Type)(schema.ObjectAsType)
		defName = g.getRootTypeName(schema, fileName)
		if len(def.Type) == 0 {
			// Minor hack to make definitions default to being objects
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

	var sg *schemaGenerator
	if fileName != "" {
		output, err := g.findOutputFileForSchemaID(schema.ID)
		if err != nil {
			return nil, err
		}

		sg = &schemaGenerator{
			Generator:      g.Generator,
			schema:         schema,
			schemaFileName: fileName,
			output:         output,
		}
	} else {
		sg = g
	}

	t, err := sg.generateDeclaredType(def, newNameScope(defName))
	if err != nil {
		return nil, err
	}

	nt := t.(*codegen.NamedType)

	if isCycle {
		g.warner(fmt.Sprintf("Cycle detected; must wrap type %s in pointer", nt.Decl.Name))
		t = codegen.WrapTypeInPointer(t)
	}

	if sg.output.file.Package.QualifiedName == g.output.file.Package.QualifiedName {
		return t, nil
	}

	var imp *codegen.Import
	for _, i := range g.output.file.Package.Imports {
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
	t *schemas.Type, scope nameScope) (codegen.Type, error) {
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
		// Don't declare named types under a new name
		delete(g.output.declsBySchema, t)
		delete(g.output.declsByName, decl.Name)
		return theType, nil
	}
	decl.Type = theType

	g.output.file.Package.AddDecl(&decl)

	if structType, ok := theType.(*codegen.StructType); ok {
		var validators []validator
		for _, f := range structType.RequiredJSONFields {
			validators = append(validators, &requiredValidator{f})
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
			if _, ok := f.Type.(codegen.NullType); ok {
				validators = append(validators, &nullTypeValidator{
					fieldName: f.Name,
					jsonName:  f.JSONName,
				})
			} else {
				t, arrayDepth := f.Type, 0
				for v, ok := t.(*codegen.ArrayType); ok; v, ok = t.(*codegen.ArrayType) {
					arrayDepth++
					if _, ok := v.Type.(codegen.NullType); ok {
						validators = append(validators, &nullTypeValidator{
							fieldName:  f.Name,
							jsonName:   f.JSONName,
							arrayDepth: arrayDepth,
						})
						break
					} else {
						if f.SchemaType.MinItems != 0 || f.SchemaType.MaxItems != 0 {
							validators = append(validators, &arrayValidator{
								fieldName:  f.Name,
								jsonName:   f.JSONName,
								arrayDepth: arrayDepth,
								minItems:   f.SchemaType.MinItems,
								maxItems:   f.SchemaType.MaxItems,
							})
						}
					}

					t = v.Type
				}
			}
		}

		if len(validators) > 0 {
			for _, v := range validators {
				if v.desc().hasError {
					g.output.file.Package.AddImport("fmt", "")
					break
				}
			}

			g.output.file.Package.AddImport("encoding/json", "")
			g.output.file.Package.AddDecl(&codegen.Method{
				Impl: func(out *codegen.Emitter) {
					out.Comment("UnmarshalJSON implements json.Unmarshaler.")
					out.Println("func (j *%s) UnmarshalJSON(b []byte) error {", decl.Name)
					out.Indent(1)
					out.Println("var %s map[string]interface{}", varNameRawMap)
					out.Println("if err := json.Unmarshal(b, &%s); err != nil { return err }",
						varNameRawMap)
					for _, v := range validators {
						if v.desc().beforeJSONUnmarshal {
							v.generate(out)
						}
					}

					out.Println("type Plain %s", decl.Name)
					out.Println("var %s Plain", varNamePlainStruct)
					out.Println("if err := json.Unmarshal(b, &%s); err != nil { return err }",
						varNamePlainStruct)

					for _, v := range validators {
						if !v.desc().beforeJSONUnmarshal {
							v.generate(out)
						}
					}

					out.Println("*j = %s(%s)", decl.Name, varNamePlainStruct)
					out.Println("return nil")
					out.Indent(-1)
					out.Println("}")
				},
			})
		}
	}

	return &codegen.NamedType{Decl: &decl}, nil
}

func (g *schemaGenerator) generateType(
	t *schemas.Type, scope nameScope) (codegen.Type, error) {
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
	if len(t.Type) != 1 {
		// TODO: Support validation for properties with multiple types
		g.warner("Property has multiple types; will be represented as interface{} with no validation")
		return codegen.EmptyInterfaceType{}, nil
	}

	switch t.Type[0] {
	case schemas.TypeNameArray:
		if t.Items == nil {
			return nil, errors.New("array property must have 'items' set to a type")
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
		return codegen.PrimitiveTypeFromJSONSchemaType(t.Type[0])
	}
}

func (g *schemaGenerator) generateStructType(
	t *schemas.Type,
	scope nameScope) (codegen.Type, error) {
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
			KeyType:   codegen.PrimitiveType{Type: "string"},
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

		fieldName := g.identifierize(name)
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

		if isRequired {
			structField.Tags = fmt.Sprintf(`json:"%s"`, name)
		} else {
			structField.Tags = fmt.Sprintf(`json:"%s,omitempty"`, name)
		}

		if structField.Comment == "" {
			structField.Comment = fmt.Sprintf("%s corresponds to the JSON schema field %q.",
				structField.Name, name)
		}

		var err error
		structField.Type, err = g.generateTypeInline(prop, scope.add(structField.Name))
		if err != nil {
			return nil, fmt.Errorf("could not generate type for field %q: %s", name, err)
		}

		if prop.Default != nil {
			structField.DefaultValue = prop.Default
		} else if isRequired {
			structType.RequiredJSONFields = append(structType.RequiredJSONFields, structField.JSONName)
		} else {
			// Optional, so must be pointer
			if !structField.Type.IsNillable() {
				structField.Type = codegen.WrapTypeInPointer(structField.Type)
			}
		}

		structType.AddField(structField)
	}
	return &structType, nil
}

func (g *schemaGenerator) generateTypeInline(
	t *schemas.Type,
	scope nameScope) (codegen.Type, error) {
	if t.Enum == nil && t.Ref == "" {
		if ext := t.GoJSONSchemaExtension; ext != nil {
			for _, pkg := range ext.Imports {
				g.output.file.Package.AddImport(pkg, "")
			}
			if ext.Type != nil {
				return &codegen.CustomNameType{Type: *ext.Type}, nil
			}
		}

		if len(t.Type) > 1 {
			g.warner("Property has multiple types; will be represented as interface{} with no validation")
			return codegen.EmptyInterfaceType{}, nil
		}
		if len(t.Type) == 0 {
			return codegen.EmptyInterfaceType{}, nil
		}

		if schemas.IsPrimitiveType(t.Type[0]) {
			return codegen.PrimitiveTypeFromJSONSchemaType(t.Type[0])
		}

		if t.Type[0] == schemas.TypeNameArray {
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
	t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if len(t.Enum) == 0 {
		return nil, errors.New("enum array cannot be empty")
	}

	var wrapInStruct bool
	var enumType codegen.Type
	if len(t.Type) == 1 {
		var err error
		if enumType, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Type[0]); err != nil {
			return nil, err
		}
		wrapInStruct = t.Type[0] == schemas.TypeNameNull // Null uses interface{}, which cannot have methods
	} else {
		if len(t.Type) > 1 {
			// TODO: Support multiple types
			g.warner("Enum defined with multiple types; ignoring it and using enum values instead")
		}

		var primitiveType string
		for _, v := range t.Enum {
			var valueType string
			if v == nil {
				valueType = "interface{}"
			} else {
				switch v.(type) {
				case string:
					valueType = "string"
				case float64:
					valueType = "float64"
				case bool:
					valueType = "bool"
				default:
					return nil, fmt.Errorf("enum has non-primitive value %v", v)
				}
			}
			if primitiveType == "" {
				primitiveType = valueType
			} else if primitiveType != valueType {
				primitiveType = "interface{}"
				break
			}
		}
		if primitiveType == "interface{}" {
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

	valueConstant := &codegen.Var{
		Name:  "enumValues_" + enumDecl.Name,
		Value: t.Enum,
	}
	g.output.file.Package.AddDecl(valueConstant)

	if wrapInStruct {
		g.output.file.Package.AddImport("encoding/json", "")
		g.output.file.Package.AddDecl(&codegen.Method{
			Impl: func(out *codegen.Emitter) {
				out.Comment("MarshalJSON implements json.Marshaler.")
				out.Println("func (j *%s) MarshalJSON() ([]byte, error) {", enumDecl.Name)
				out.Indent(1)
				out.Println("return json.Marshal(j.Value)")
				out.Indent(-1)
				out.Println("}")
			},
		})
	}

	g.output.file.Package.AddImport("fmt", "")
	g.output.file.Package.AddImport("reflect", "")
	g.output.file.Package.AddImport("encoding/json", "")
	g.output.file.Package.AddDecl(&codegen.Method{
		Impl: func(out *codegen.Emitter) {
			out.Comment("UnmarshalJSON implements json.Unmarshaler.")
			out.Println("func (j *%s) UnmarshalJSON(b []byte) error {", enumDecl.Name)
			out.Indent(1)
			out.Print("var v ")
			enumType.Generate(out)
			out.Newline()
			varName := "v"
			if wrapInStruct {
				varName += ".Value"
			}
			out.Println("if err := json.Unmarshal(b, &%s); err != nil { return err }", varName)
			out.Println("var ok bool")
			out.Println("for _, expected := range %s {", valueConstant.Name)
			out.Println("if reflect.DeepEqual(%s, expected) { ok = true; break }", varName)
			out.Println("}")
			out.Println("if !ok {")
			out.Println(`return fmt.Errorf("invalid value (expected one of %%#v): %%#v", %s, %s)`,
				valueConstant.Name, varName)
			out.Println("}")
			out.Println(`*j = %s(v)`, enumDecl.Name)
			out.Println(`return nil`)
			out.Indent(-1)
			out.Println("}")
		},
	})

	// TODO: May be aliased string type
	if prim, ok := enumType.(codegen.PrimitiveType); ok && prim.Type == "string" {
		for _, v := range t.Enum {
			if s, ok := v.(string); ok {
				// TODO: Make sure the name is unique across scope
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
	if _, ok := o.declsByName[name]; !ok {
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

type cachedEnum struct {
	values []interface{}
	enum   *codegen.TypeDecl
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

var (
	varNamePlainStruct = "plain"
	varNameRawMap      = "raw"
)

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil || !os.IsNotExist(err)
}
