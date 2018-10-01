package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

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
	emitter               *codegen.Emitter
	defaultPackageName    string
	defaultOutputName     string
	schemaMappings        []SchemaMapping
	warner                func(string)
	outputs               map[string]*output
	schemaCacheByFileName map[string]*schemas.Schema
}

func New(
	schemaMappings []SchemaMapping,
	defaultPackageName string,
	defaultOutputName string,
	warner func(string)) (*Generator, error) {
	return &Generator{
		warner:                warner,
		schemaMappings:        schemaMappings,
		defaultPackageName:    defaultPackageName,
		defaultOutputName:     defaultOutputName,
		outputs:               map[string]*output{},
		schemaCacheByFileName: map[string]*schemas.Schema{},
	}, nil
}

func (g *Generator) Sources() map[string][]byte {
	sources := make(map[string]*strings.Builder, len(g.outputs))
	for _, output := range g.outputs {
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
		result[f] = []byte(sb.String())
	}
	return result
}

func (g *Generator) DoFile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	schema, err := schemas.FromReader(f)
	if err != nil {
		return err
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

func (g *Generator) loadSchemaFromFile(fileName, parentFileName string) (*schemas.Schema, error) {
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(filepath.Dir(parentFileName), fileName)
	}

	fileName, err := filepath.EvalSymlinks(fileName)
	if err != nil {
		return nil, err
	}

	if schema, ok := g.schemaCacheByFileName[fileName]; ok {
		return schema, nil
	}

	schema, err := schemas.FromFile(fileName)
	if err != nil {
		return nil, err
	}
	g.schemaCacheByFileName[fileName] = schema

	if err = g.addFile(fileName, schema); err != nil {
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

func (g *Generator) findOutputFileForSchemaID(id string) (*output, error) {
	if o, ok := g.outputs[id]; ok {
		return o, nil
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
	outputName, packageName string) (*output, error) {
	if outputName == "" {
		return nil, fmt.Errorf("unable to map schema URI %q to a file name", id)
	}
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
		enums:         map[string]cachedEnum{},
	}
	g.outputs[id] = output
	return output, nil
}

type schemaGenerator struct {
	*Generator
	output         *output
	schema         *schemas.Schema
	schemaFileName string
}

func (g *schemaGenerator) generateRootType() error {
	if g.schema.Type == nil {
		return errors.New("schema has no root")
	}

	if g.schema.Type.Type == "" {
		for name, def := range g.schema.Definitions {
			_, err := g.generateDeclaredType(def, newNameScope(codegen.Identifierize(name)))
			if err != nil {
				return err
			}
		}
		return nil
	}

	rootTypeName := g.getRootTypeName(g.schema, g.schemaFileName)
	if _, ok := g.output.declsByName[rootTypeName]; ok {
		return nil
	}

	_, err := g.generateDeclaredType(g.schema.Type, newNameScope(rootTypeName))
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
		if def.Type == "" && len(def.Properties) == 0 {
			return nil, nil
		}
		// Minor hack to make definitions default to being objects
		def.Type = schemas.TypeNameObject
		defName = codegen.Identifierize(defName)
	} else {
		def = schema.Type
		defName = g.getRootTypeName(schema, fileName)
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
		Decl:    t.(*codegen.NamedType).Decl,
	}, nil
}

func (g *schemaGenerator) generateDeclaredType(
	t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if t, ok := g.output.declsBySchema[t]; ok {
		return &codegen.NamedType{Decl: t}, nil
	}

	decl := codegen.TypeDecl{
		Name:    g.output.uniqueTypeName(scope.string()),
		Comment: t.Description,
	}
	theType, err := g.generateType(t, scope)
	if err != nil {
		return nil, err
	}
	if d, ok := theType.(*codegen.NamedType); ok {
		return d, nil
	}
	decl.Type = theType

	g.output.declsBySchema[t] = &decl
	g.output.declsByName[decl.Name] = &decl
	g.output.file.Package.AddDecl(&decl)

	if structType, ok := theType.(*codegen.StructType); ok {
		g.output.file.Package.AddImport("encoding/json", "")
		g.output.file.Package.AddDecl(&codegen.Method{
			Impl: func(out *codegen.Emitter) {
				out.Comment("UnmarshalJSON implements json.Unmarshaler.")
				out.Println("func (j *%s) UnmarshalJSON(b []byte) error {", decl.Name)
				out.Indent(1)
				out.Println("var v struct {")
				out.Indent(1)
				fields := append([]codegen.StructField{}, structType.Fields...)
				for _, f := range structType.Fields {
					if f.Synthetic {
						f.Generate(out)
						out.Newline()
					}
				}
				out.Indent(-1)
				out.Println("}")
				out.Println("if err := json.Unmarshal(b, &v); err != nil { return err }")
				for _, f := range fields {
					if f.Synthetic {
						for _, r := range f.Rules {
							r.GenerateValidation(out, fmt.Sprintf("v.%s", f.Name),
								fmt.Sprintf("field %s", f.JSONName))
						}
					}
				}
				out.Println("type plain %s", decl.Name)
				out.Println("var p plain")
				out.Println("if err := json.Unmarshal(b, &p); err != nil { return err }")
				for _, f := range fields {
					if !f.Synthetic {
						for _, r := range f.Rules {
							r.GenerateValidation(out, fmt.Sprintf("p.%s", f.Name),
								fmt.Sprintf("field %s", f.JSONName))
						}
					}
				}
				out.Println("*j = %s(p)", decl.Name)
				out.Println("return nil")
				out.Indent(-1)
				out.Println("}")
			},
		})
	}

	return &codegen.NamedType{Decl: &decl}, nil
}

func (g *schemaGenerator) generateType(
	t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if t.Enum != nil {
		return g.generateEnumType(t, scope)
	}

	switch t.Type {
	case schemas.TypeNameArray:
		if t.Items == nil {
			return nil, errors.New("array property must have 'items' set to a type")
		}

		var elemType codegen.Type
		if schemas.IsPrimitiveType(t.Items.Type) {
			var err error
			elemType, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Items.Type)
			if err != nil {
				return nil, fmt.Errorf("cannot determine type of field: %s", err)
			}
		} else if t.Items.Type != "" {
			var err error
			elemType, err = g.generateDeclaredType(t.Items, scope.add("Elem"))
			if err != nil {
				return nil, err
			}
		} else if t.Items.Ref != "" {
			var err error
			elemType, err = g.generateReferencedType(t.Items.Ref)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("array property must have a type")
		}
		return codegen.ArrayType{elemType}, nil
	case schemas.TypeNameObject:
		return g.generateStructType(t, scope)
	case schemas.TypeNameNull:
		return codegen.EmptyInterfaceType{}, nil
	}

	if t.Ref != "" {
		return g.generateReferencedType(t.Ref)
	}
	return codegen.PrimitiveTypeFromJSONSchemaType(t.Type)
}

func (g *schemaGenerator) generateStructType(
	t *schemas.Type,
	scope nameScope) (codegen.Type, error) {
	if len(t.Properties) == 0 {
		if len(t.Required) > 0 {
			g.warner("object type with no properties has required fields; " +
				"skipping validation code for them since we don't know their types")
		}
		return &codegen.MapType{
			KeyType:   codegen.PrimitiveType{"string"},
			ValueType: codegen.EmptyInterfaceType{},
		}, nil
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
	for _, name := range propNames {
		prop := t.Properties[name]
		isRequired := requiredNames[name]

		fieldName := codegen.Identifierize(name)
		if count, ok := uniqueNames[fieldName]; ok {
			uniqueNames[fieldName] = count + 1
			fieldName = fmt.Sprintf("%s_%d", fieldName, count+1)
			g.warner(fmt.Sprintf("field %q maps to a field by the same name declared "+
				"in the same struct; it will be declared as %s", name, fieldName))
		} else {
			uniqueNames[fieldName] = 1
		}

		structField := codegen.StructField{
			Name:     fieldName,
			Comment:  prop.Description,
			JSONName: name,
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

		if isRequired {
			if rt, ok := structField.Type.(codegen.RuledType); ok {
				g.output.file.Package.AddImport("fmt", "") // All rules need fmt
				for _, r := range rt.GetRequiredRules() {
					structField.AddRule(r)
				}
			}
			if !structField.Type.IsNillable() {
				g.output.file.Package.AddImport("fmt", "") // All rules need fmt
				syntheticField := structField
				syntheticField.Comment = ""
				syntheticField.Synthetic = true
				syntheticField.Type = codegen.PointerType{Type: syntheticField.Type}
				syntheticField.AddRule(codegen.NilStructFieldRequired{})
				structType.AddField(syntheticField)
			}
		} else if !structField.Type.IsNillable() {
			structField.Type = codegen.PointerType{Type: structField.Type}
		}

		structType.AddField(structField)
	}
	return &structType, nil
}

func (g *schemaGenerator) generateTypeInline(
	t *schemas.Type,
	scope nameScope) (codegen.Type, error) {
	if schemas.IsPrimitiveType(t.Type) && t.Enum == nil && t.Ref == "" {
		return codegen.PrimitiveTypeFromJSONSchemaType(t.Type)
	}

	if t.Type == schemas.TypeNameArray {
		theType, err := g.generateTypeInline(t.Items, scope.add("Elem"))
		if err != nil {
			return nil, err
		}
		return &codegen.ArrayType{Type: theType}, nil
	}

	return g.generateDeclaredType(t, scope)
}

func (g *schemaGenerator) generateEnumType(
	t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if len(t.Enum) == 0 {
		return nil, errors.New("enum array cannot be empty")
	}

	if decl, ok := g.output.findEnum(t.Enum); ok {
		return decl, nil
	}

	var wrapInStruct bool
	var enumType codegen.Type
	if t.Type != "" {
		var err error
		if enumType, err = codegen.PrimitiveTypeFromJSONSchemaType(t.Type); err != nil {
			return nil, err
		}
		wrapInStruct = t.Type == schemas.TypeNameNull // Null uses interface{}, which cannot have methods
	} else {
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
		enumType = codegen.PrimitiveType{primitiveType}
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

	if enumDecl, ok := enumType.(*codegen.NamedType); ok {
		return enumDecl, nil
	}

	enumDecl := codegen.TypeDecl{
		Name: g.output.uniqueTypeName(scope.add("Enum").string()),
		Type: enumType,
	}
	g.output.file.Package.AddDecl(&enumDecl)

	g.output.declsByName[enumDecl.Name] = &enumDecl
	g.output.enums[hashArrayOfValues(t.Enum)] = cachedEnum{
		enum:   &enumDecl,
		values: t.Enum,
	}

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
					Name:  makeEnumConstantName(enumDecl.Name, s),
					Type:  &codegen.NamedType{Decl: &enumDecl},
					Value: s,
				})
			}
		}
	}

	return &codegen.NamedType{Decl: &enumDecl}, nil
}

// func (g *schemaGenerator) generateAnonymousType(
// 	t *schemas.Type, name string) (codegen.Type, error) {
// 	if t.Type == schemas.TypeNameObject {
// 		if len(t.Properties) == 0 {
// 			return codegen.MapType{
// 				KeyType:   codegen.PrimitiveType{"string"},
// 				ValueType: codegen.EmptyInterfaceType{},
// 			}, nil
// 		}
// 		s, err := g.generateStructDecl(t, name, "", false)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return &codegen.NamedType{Decl: s}, nil
// 	}
// 	return nil, fmt.Errorf("unexpected type %q", t.Type)
// }

type output struct {
	file          *codegen.File
	enums         map[string]cachedEnum
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
				"multiple types map to the name %q; declaring duplicate as %q instead", name, suffixed))
			return suffixed
		}
		count++
	}
}

func (o *output) findEnum(values []interface{}) (codegen.Type, bool) {
	key := hashArrayOfValues(values)
	if t, ok := o.enums[key]; ok && reflect.DeepEqual(values, t.values) {
		return &codegen.NamedType{Decl: t.enum}, true
	}
	return nil, false
}

type cachedEnum struct {
	values []interface{}
	enum   *codegen.TypeDecl
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
