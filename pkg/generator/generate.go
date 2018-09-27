package generator

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/sanity-io/litter"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

type SchemaMapping struct {
	SchemaID    string
	PackageName string
	RootType    string
	OutputName  string
}

type Generator struct {
	emitter            *codegen.Emitter
	defaultPackageName string
	defaultOutputName  string
	schemaMappings     []SchemaMapping
	warner             func(string)

	schemaFiles map[string]*codegen.File
	schemasRead map[string]*schemas.Schema
	enums       map[string]cachedEnum
	types       map[string]*codegen.TypeDecl
}

func New(
	schemaMappings []SchemaMapping,
	defaultPackageName string,
	defaultOutputName string,
	warner func(string)) (*Generator, error) {
	return &Generator{
		warner:             warner,
		schemaMappings:     schemaMappings,
		defaultPackageName: defaultPackageName,
		defaultOutputName:  defaultOutputName,
		schemaFiles:        map[string]*codegen.File{},
		types:              map[string]*codegen.TypeDecl{},
		schemasRead:        map[string]*schemas.Schema{},
		enums:              map[string]cachedEnum{},
	}, nil
}

func (g *Generator) Sources() map[string][]byte {
	sources := make(map[string]*strings.Builder, len(g.schemaFiles))
	for _, f := range g.schemaFiles {
		emitter := codegen.NewEmitter(80)
		f.Generate(emitter)

		sb, ok := sources[f.FileName]
		if !ok {
			sb = &strings.Builder{}
			sources[f.FileName] = sb
		}
		_, _ = sb.WriteString(emitter.String())
	}

	result := make(map[string][]byte, len(sources))
	for f, sb := range sources {
		result[f] = []byte(sb.String())
	}
	return result
}

func (g *Generator) AddFile(fileName string, schema *schemas.Schema) error {
	f, err := g.findOutputFileForSchemaID(schema.ID)
	if err != nil {
		return err
	}

	return (&schemaGenerator{
		Generator:      g,
		schema:         schema,
		schemaFileName: fileName,
		file:           f,
	}).generateRootType()
}

func (g *Generator) loadSchemaFromFile(fileName, parentFileName string) (*schemas.Schema, error) {
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(filepath.Dir(parentFileName), fileName)
	}

	fileName, err := filepath.EvalSymlinks(fileName)
	if err != nil {
		return nil, err
	}

	if schema, ok := g.schemasRead[fileName]; ok {
		return schema, nil
	}

	schema, err := schemas.FromFile(fileName)
	if err != nil {
		return nil, err
	}
	g.schemasRead[fileName] = schema

	if err = g.AddFile(fileName, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (g *Generator) getRootTypeName(schema *schemas.Schema, fileName string) string {
	for _, m := range g.schemaMappings {
		if m.SchemaID == schema.ID && m.RootType != "" {
			return m.RootType
		}
	}
	return codegen.IdentifierFromFileName(fileName)
}

func (g *Generator) findOutputFileForSchemaID(id string) (*codegen.File, error) {
	if f, ok := g.schemaFiles[id]; ok {
		return f, nil
	}

	for _, m := range g.schemaMappings {
		if m.SchemaID == id {
			return g.beginOutput(id, m.OutputName, m.PackageName)
		}
	}
	return g.beginOutput(id, g.defaultOutputName, g.defaultPackageName)
}

func (g *Generator) beginOutput(
	id string,
	outputName, packageName string) (*codegen.File, error) {
	if outputName == "" {
		return nil, fmt.Errorf("unable to map schema URI %q to a file name", id)
	}
	if packageName == "" {
		return nil, fmt.Errorf("unable to map schema URI %q to a Go package name", id)
	}

	for _, f := range g.schemaFiles {
		if f.FileName == outputName && f.Package.QualifiedName != packageName {
			return nil, fmt.Errorf(
				"conflict: same file (%s) mapped to two different Go packages (%q and %q) for schema %q",
				f.FileName, f.Package.QualifiedName, packageName, id)
		}
		if f.FileName == outputName && f.Package.QualifiedName == packageName {
			return f, nil
		}
	}

	pkg := codegen.Package{
		QualifiedName: packageName,
	}
	pkg.AddDecl(codegen.Fragment(func(out *codegen.Emitter) {
		out.Println(`func validateEnum(value interface{}, expected ...interface{}) error {
  for _, v := range expected {
		if reflect.DeepEqual(value, v) {
			return true
		}
	}
	return fmt.Errorf("invalid value: %%#v", value)
}`)
	}))

	f := codegen.File{
		FileName: outputName,
		Package:  pkg,
	}
	g.schemaFiles[id] = &f
	return &f, nil
}

type schemaGenerator struct {
	*Generator
	file           *codegen.File
	schema         *schemas.Schema
	schemaFileName string
}

func (g *schemaGenerator) generateRootType() error {
	if g.schema.Type == nil {
		return errors.New("schema has no root type")
	}
	if g.schema.Type.Type != schemas.TypeNameObject {
		return fmt.Errorf("type of root must be object; found %q", g.schema.Type.Type)
	}

	rootTypeName := g.getRootTypeName(g.schema, g.schemaFileName)
	if _, ok := g.types[rootTypeName]; ok {
		return nil
	}

	_, err := g.generateStructType(rootTypeName, "", g.schema.Type, g.schema.Definitions, true)
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

	var def *schemas.Type
	if defName != "" {
		// TODO: Support nested definitions
		var ok bool
		def, ok = schema.Definitions[defName]
		if !ok {
			return nil, fmt.Errorf("definition %q (from ref %q) does not exist in schema", defName, ref)
		}
	} else {
		def = schema.Type
		defName = g.getRootTypeName(schema, fileName)
	}

	var sg *schemaGenerator
	if fileName != "" {
		f, err := g.findOutputFileForSchemaID(schema.ID)
		if err != nil {
			return nil, err
		}

		sg = &schemaGenerator{
			Generator:      g.Generator,
			schema:         schema,
			schemaFileName: fileName,
			file:           f,
		}
	} else {
		sg = g
	}

	t, err := sg.generateStructType(codegen.Identifierize(defName), defName, def, schema.Definitions, false)
	if err != nil {
		return nil, err
	}

	namedType, ok := t.(*codegen.NamedType)
	if !ok {
		panic(fmt.Sprintf("expected *NamedType, got %T", t))
	}

	if sg.file.Package.QualifiedName == g.file.Package.QualifiedName {
		return namedType, nil
	}

	var imp *codegen.Import
	for _, i := range g.file.Package.Imports {
		if i.Name == sg.file.Package.Name() && i.QualifiedName == sg.file.Package.QualifiedName {
			imp = &i
			break
		}
	}
	if imp == nil {
		g.file.Package.AddImport(codegen.Import{
			Name:          sg.file.Package.Name(),
			QualifiedName: sg.file.Package.QualifiedName,
		})
	}

	// TODO: Use better result here that doesn't need string concatenation
	return codegen.PrimitiveType{sg.file.Package.Name() + "." + namedType.Decl.Name}, nil
}

func (g *schemaGenerator) generateStructType(
	typeName, origName string,
	t *schemas.Type,
	defs schemas.Definitions,
	isRoot bool) (codegen.Type, error) {
	if typeName == "" {
		return nil, errors.New("empty type name")
	}

	if s, ok := g.types[typeName]; ok {
		return &codegen.NamedType{Decl: s}, nil
	}

	structDecl := codegen.TypeDecl{
		Name:    typeName,
		Comment: t.Description,
	}
	g.types[typeName] = &structDecl

	if structDecl.Comment == "" {
		if origName != "" {
			structDecl.Comment = fmt.Sprintf("%s corresponds to the JSON schema type %q.",
				typeName, origName)
		} else if isRoot {
			structDecl.Comment = fmt.Sprintf("%s corresponds to the root of the JSON schema %q.",
				typeName, filepath.Base(g.schemaFileName))
		}
	}

	propNames := make([]string, 0, len(t.Properties))
	for name := range t.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	requiredNames := make(map[string]bool, len(t.Properties))
	for _, r := range t.Required {
		requiredNames[r] = true
	}

	uniqueNames := make(map[string]int, len(propNames))

	var structType codegen.StructType
	structDecl.Type = &structType

	for _, name := range propNames {
		prop := t.Properties[name]

		fieldName := codegen.Identifierize(name)
		if count, ok := uniqueNames[fieldName]; ok {
			uniqueNames[fieldName] = count + 1
			fieldName = fmt.Sprintf("%s_%d", fieldName, count+1)
			g.warner(fmt.Sprintf("field %q maps to a field by the same name declared "+
				"in the struct %s; it will be declared as %s", name, structDecl.Name, fieldName))
		} else {
			uniqueNames[fieldName] = 1
		}

		structField := codegen.StructField{
			Name:       fieldName,
			Comment:    prop.Description,
			JSONName:   name,
			IsRequired: requiredNames[name],
		}

		if structField.IsRequired {
			structField.Tags = fmt.Sprintf(`json:"%s"`, name)
		} else {
			structField.Tags = fmt.Sprintf(`json:"%s,omitempty"`, name)
		}

		if structField.Comment == "" {
			structField.Comment = fmt.Sprintf("%s corresponds to the JSON schema field %q.",
				structField.Name, name)
		}

		if err := g.generateTypeForStructField(
			name, prop, &structDecl, &structType, &structField); err != nil {
			return nil, err
		}

		structType.AddField(structField)
	}

	g.file.Package.AddDecl(&structDecl)
	g.file.Package.AddDecl(&codegen.Method{
		Impl: func(out *codegen.Emitter) {
			out.Comment("UnmarshalJSON implements json.Unmarshaler.")
			out.Println("func (j *%s) UnmarshalJSON(b []byte) error {", structDecl.Name)
			out.Indent(1)
			out.Println("var v struct {")
			out.Indent(1)
			out.Println("%s", structDecl.Name)

			fields := append([]codegen.StructField{}, structType.Fields...)
			for _, f := range structType.Fields {
				if f.Synthetic {
					f.Generate(out)
				}
			}

			out.Indent(-1)
			out.Println("}")
			out.Println("if err := json.Unmarshal(b, &v); err != nil {")
			out.Indent(1)
			out.Println("return err")
			out.Indent(-1)
			out.Println("}")
			for _, f := range fields {
				for _, r := range f.Rules {
					r.GenerateValidation(out, fmt.Sprintf("v.%s", f.Name),
						fmt.Sprintf("field %s", f.JSONName))
				}
			}
			// for _, f := range fields {
			// 	f.generateValidation(out)
			// }
			out.Println("*j = v.%s", structDecl.Name)
			out.Println("return nil")
			out.Indent(-1)
			out.Println("}")
		},
	})

	return &codegen.NamedType{Decl: &structDecl}, nil
}

func (g *schemaGenerator) generateTypeForStructField(
	name string,
	t *schemas.Type,
	parentStructDecl *codegen.TypeDecl,
	parentStructType *codegen.StructType,
	structField *codegen.StructField) error {
	if t.Enum != nil {
		return g.generateTypeForStructFieldEnum(name, t, parentStructDecl, parentStructType, structField)
	}

	switch t.Type {
	case schemas.TypeNameArray:
		if t.Items == nil {
			return fmt.Errorf("array property %q must have 'items' set to a type", name)
		}

		var elemType codegen.Type
		if schemas.IsPrimitiveType(t.Items.Type) {
			var err error
			elemType, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Items.Type)
			if err != nil {
				return fmt.Errorf("cannot determine type of field %q: %s", name, err)
			}
		} else if t.Items.Type != "" {
			var err error
			elemType, err = g.generateAnonymousType(t.Items, name, parentStructDecl)
			if err != nil {
				return fmt.Errorf("cannot determine type of array field %q: %s", name, err)
			}
		} else if t.Items.Ref != "" {
			var err error
			elemType, err = g.generateReferencedType(t.Items.Ref)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("array property %q must have a type", name)
		}

		structField.Type = codegen.ArrayType{elemType}
		if structField.IsRequired {
			structField.AddRule(codegen.ArrayNotEmpty{})
		}
	case schemas.TypeNameObject:
		var err error
		structField.Type, err = g.generateAnonymousType(t, name, parentStructDecl)
		if err != nil {
			return fmt.Errorf("cannot determine type of array field %q: %s", name, err)
		}
	case schemas.TypeNameNull:
		structField.Type = codegen.EmptyInterfaceType{}
	default:
		if t.Ref != "" {
			var err error
			structField.Type, err = g.generateReferencedType(t.Ref)
			if err != nil {
				return err
			}
		} else {
			var err error
			structField.Type, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Type)
			if err != nil {
				return fmt.Errorf("cannot determine type of field %q: %s", name, err)
			}
		}
	}

	if structField.IsRequired && !structField.Type.IsNillable() {
		syntheticField := *structField
		syntheticField.Comment = ""
		syntheticField.Synthetic = true
		syntheticField.Name = "__synthetic_" + syntheticField.Name
		syntheticField.Type = codegen.PointerType{Type: syntheticField.Type}
		syntheticField.AddRule(codegen.NilStructFieldRequired{})
		parentStructType.AddField(syntheticField)
	}
	return nil
}

func (g *schemaGenerator) generateTypeForStructFieldEnum(
	name string,
	t *schemas.Type,
	parentStructDecl *codegen.TypeDecl,
	parentStructType *codegen.StructType,
	structField *codegen.StructField) error {
	if len(t.Enum) == 0 {
		return fmt.Errorf("property %q enum array cannot be empty", name)
	}

	if enumDecl, ok := g.findEnum(t.Enum); ok {
		structField.Type = enumDecl
		return nil
	}

	var propType codegen.Type
	if t.Type != "" {
		var err error
		if propType, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Type); err != nil {
			return err
		}
	} else {
		var enumPrimitiveType string
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
					return fmt.Errorf("property %q enum has non-primitive value %v", name, v)
				}
			}
			if enumPrimitiveType == "" {
				enumPrimitiveType = valueType
			} else if enumPrimitiveType != valueType {
				enumPrimitiveType = "interface{}"
				break
			}
		}
		propType = codegen.PrimitiveType{enumPrimitiveType}
	}

	enumDecl := codegen.TypeDecl{
		Name:    codegen.Identifierize(name) + "Enum",
		Type:    propType,
		Comment: t.Description,
	}
	g.file.Package.AddDecl(&enumDecl)
	g.file.Package.AddDecl(&codegen.Method{
		Impl: func(out *codegen.Emitter) {
			out.Comment("UnmarshalJSON implements json.Unmarshaler.")
			out.Println("func (j *%s) UnmarshalJSON(b []byte) error {", enumDecl.Name)
			out.Indent(1)
			out.Print("var v ")
			propType.Generate(out)
			out.Newline()
			out.Println("if err := json.Unmarshal(b, &v); err != nil { return err }")
			out.Print("if err := validateEnum(v, ")
			for i, v := range t.Enum {
				if i > 0 {
					out.Print(", ")
				}
				out.Print("%s", litter.Sdump(v))
			}
			out.Println("); err != nil { return err }")
			out.Println("*j = %s(v)", enumDecl.Name)
			out.Println("return nil")
			out.Indent(-1)
			out.Println("}")
		},
	})

	// TODO: May be aliased string type
	if prim, ok := propType.(codegen.PrimitiveType); ok && prim.Type == "string" {
		for _, v := range t.Enum {
			if s, ok := v.(string); ok {
				// TODO: Make sure the name is unique across scope
				g.file.Package.AddDecl(&codegen.Constant{
					Name:  enumDecl.Name + codegen.Identifierize(s),
					Type:  &codegen.NamedType{Decl: &enumDecl},
					Value: s,
				})
			}
		}
	}

	structField.Type = &codegen.NamedType{Decl: &enumDecl}

	g.enums[hashArrayOfValues(t.Enum)] = cachedEnum{
		enum:   &enumDecl,
		values: t.Enum,
	}
	return nil
}

func (g *schemaGenerator) generateAnonymousType(
	t *schemas.Type,
	fieldName string,
	parentType *codegen.TypeDecl) (codegen.Type, error) {
	if t.Type == schemas.TypeNameObject {
		if len(t.Properties) == 0 {
			return codegen.MapType{
				KeyType:   codegen.PrimitiveType{"string"},
				ValueType: codegen.EmptyInterfaceType{},
			}, nil
		}

		typeName := g.uniqueTypeName(fmt.Sprintf("%s%s", parentType.Name, codegen.Identifierize(fieldName)))
		return g.generateStructType(typeName, "", t, g.schema.Definitions, false)
	}
	return nil, fmt.Errorf("unexpected type %q", t.Type)
}

func (g *schemaGenerator) uniqueTypeName(name string) string {
	if _, ok := g.types[name]; !ok {
		return name
	}
	count := 1
	for {
		suffixed := fmt.Sprintf("%s_%d", name, count)
		if _, ok := g.types[suffixed]; !ok {
			return suffixed
		}
		count++
	}
}

func (g *schemaGenerator) findEnum(values []interface{}) (codegen.Type, bool) {
	key := hashArrayOfValues(values)
	if t, ok := g.enums[key]; ok && reflect.DeepEqual(values, t.values) {
		return &codegen.NamedType{Decl: t.enum}, true
	}
	return nil, false
}

type cachedEnum struct {
	values []interface{}
	enum   *codegen.TypeDecl
}
