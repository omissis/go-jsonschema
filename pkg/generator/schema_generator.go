package generator

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/tuotuoxp/go-jsonschema/pkg/cmputil"
	"github.com/tuotuoxp/go-jsonschema/pkg/codegen"
	"github.com/tuotuoxp/go-jsonschema/pkg/schemas"
)

var (
	//nolint: gochecknoglobals // global to avoid duplication
	boolTypeVal = codegen.PrimitiveType{Type: "bool"}
	//nolint: gochecknoglobals // global to avoid duplication
	emptyInterfaceTypeVal = codegen.EmptyInterfaceType{}
	//nolint: gochecknoglobals // global to avoid duplication
	intTypeVal = codegen.PrimitiveType{Type: "int"}
	//nolint: gochecknoglobals // global to avoid duplication
	stringTypeVal = codegen.PrimitiveType{Type: "string"}
	//nolint: gochecknoglobals // global to avoid duplication
	arrayTypeVal = codegen.ArrayType{Type: emptyInterfaceTypeVal}

	errEmptyInAnyOf = errors.New("cannot have empty anyOf array")

	//nolint:gochecknoglobals // compiled once for schema extension validation
	goIdentifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

const float64Type = "float64"

func newSchemaGenerator(
	g *Generator,
	schema *schemas.Schema,
	fileName string,
	output *output,
) *schemaGenerator {
	return &schemaGenerator{
		Generator:        g,
		schema:           schema,
		schemaFileName:   fileName,
		output:           output,
		schemaTypesByRef: make(map[string]*schemas.Type),
	}
}

type schemaGenerator struct {
	*Generator

	output           *output
	schema           *schemas.Schema
	schemaFileName   string
	schemaTypesByRef map[string]*schemas.Type
}

func (g *schemaGenerator) generateRootType() error {
	rootSchemaTarget := g.isRootSchemaTarget(g.schema, g.schemaFileName)

	// Generate $defs from the outer Schema.Definitions field.  This is safe even
	// when ObjectAsType is nil (defs-only schemas such as shared-definitions files
	// that contain only "$defs" with no root type).
	for _, name := range sortDefinitionsByName(g.schema.Definitions) {
		def := g.schema.Definitions[name]
		if !rootSchemaTarget {
			ref := fmt.Sprintf("#/$defs/%s", name)
			_, _, _, hasRefMapping, err := g.resolveReferencedXGoRefMapping(def, g.caser.Identifierize(name), ref)
			if err != nil {
				return err
			}

			if hasRefMapping {
				continue
			}
		}

		_, err := g.generateDeclaredType(def, newNameScope(g.caser.Identifierize(name)))
		if err != nil {
			return err
		}
	}

	// Defs-only schema: ObjectAsType is nil but definitions were present.
	// This is valid (a shared-definitions file with no root type).
	if g.schema.ObjectAsType == nil {
		if len(g.schema.Definitions) == 0 {
			return errSchemaHasNoRoot
		}

		return nil
	}

	// Schema with an empty type list: nothing to generate for the root type.
	if len(g.schema.Type) == 0 {
		return nil
	}

	rootTypeName := g.getRootTypeName(g.schema, g.schemaFileName)
	if _, ok := g.output.declsByName[rootTypeName]; ok {
		return nil
	}

	_, err := g.generateDeclaredType((*schemas.Type)(g.schema.ObjectAsType), newNameScope(rootTypeName))

	return err
}

func (g *schemaGenerator) generateReferencedType(t *schemas.Type) (codegen.Type, error) {
	defName, fileName, err := g.extractRefNames(t)
	if err != nil {
		return nil, err
	}

	mappedType, importPath, importAlias, hasRefMapping, mappingErr := g.resolveReferencedXGoRefMappingForRef(t)
	if mappingErr != nil {
		return nil, mappingErr
	}

	if hasRefMapping {
		g.output.file.Package.AddImport(importPath, importAlias)

		return &codegen.CustomNameType{Type: mappedType}, nil
	}

	if fileName == "" {
		defLookupName := defName
		if schemaDef, ok := g.schema.Definitions[defName]; ok && schemaDef.XGoRef != nil {
			defLookupName, err = g.resolveReferencedDefinitionTypeName(
				schemaDef,
				g.caser.Identifierize(defName),
			)
			if err != nil {
				return nil, err
			}
		}

		if schemaOutput, ok := g.outputs[g.schema.ID]; ok {
			if decl, ok := schemaOutput.declsByName[defLookupName]; ok {
				if decl != nil {
					return &codegen.NamedType{Decl: decl}, nil
				}
			}
		}
	}

	if t.Ref == "#" {
		if schemaOutput, ok := g.outputs[g.schema.ID]; ok {
			if decl, ok := schemaOutput.declsBySchema[t]; ok {
				if decl != nil {
					return decl.Type, nil
				}
			}
		}

		return emptyInterfaceTypeVal, nil
	}

	schema := g.schema
	sg := g

	if fileName != "" {
		var serr error

		schema, serr = g.loader.Load(fileName, g.schemaFileName)
		if serr != nil {
			return nil, fmt.Errorf("could not follow $ref %q to file %q: %w", t.Ref, fileName, serr)
		}

		qualified, qerr := schemas.QualifiedFileName(fileName, g.schemaFileName, g.config.ResolveExtensions)
		if qerr != nil {
			return nil, fmt.Errorf("could not resolve qualified file name for %s: %w", fileName, qerr)
		}

		if ferr := g.AddFile(qualified, schema); ferr != nil {
			return nil, ferr
		}

		output, oerr := g.findOutputFileForSchemaID(schema.ID)
		if oerr != nil {
			return nil, oerr
		}

		sg = newSchemaGenerator(g.Generator, schema, qualified, output)
	}

	var def *schemas.Type

	if defName != "" {
		// TODO: Support nested definitions.
		var ok bool

		def, ok = schema.Definitions[defName]
		if !ok {
			return nil, fmt.Errorf("%w: %q (from ref %q)", errDefinitionDoesNotExistInSchema, defName, t.Ref)
		}

		defName, err = g.resolveReferencedDefinitionTypeName(def, g.caser.Identifierize(defName))
		if err != nil {
			return nil, err
		}
	} else {
		def = (*schemas.Type)(schema.ObjectAsType)
		defName = g.getRootTypeName(schema, fileName)

		if len(def.Type) == 0 {
			// Minor hack to make definitions default to being objects.
			def.Type = schemas.TypeList{schemas.TypeNameObject}
		}
	}

	isCycle, cleanupCycle, cycleErr := g.detectCycle(t)
	if cycleErr != nil {
		return nil, cycleErr
	}

	defer cleanupCycle()

	dt, err := sg.generateDeclaredType(def, newNameScope(defName))
	if err != nil {
		return nil, err
	}

	// We need this in order to handle cases when
	// generateDeclaredType outputs a MapType and not a NamedType.
	if isMapType(dt) && g.config.DisableCustomTypesForMaps {
		return dt, nil
	}

	nt, ok := dt.(*codegen.NamedType)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", errExpectedNamedType, t)
	}

	if isCycle {
		g.warner(fmt.Sprintf("Cycle detected; must wrap type %s in pointer", nt.Decl.Name))

		dt = codegen.WrapTypeInPointer(dt)
	}

	if sg.output.file.Package.QualifiedName == g.output.file.Package.QualifiedName {
		return dt, nil
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

func (g *schemaGenerator) extractRefNames(t *schemas.Type) (string, string, error) {
	scope := ""
	defName := ""
	fileName := t.Ref

	if i := strings.IndexRune(t.Ref, '#'); i != -1 {
		var prefix string

		fileName, scope = t.Ref[0:i], t.Ref[i+1:]
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
			return "", "", fmt.Errorf(
				"%w: value must point to definition within file: '%s'",
				errCannotGenerateReferencedType,
				t.Ref,
			)
		}

		defName = scope[len(prefix):]
	}

	return defName, fileName, nil
}

//nolint:gocyclo // todo: reduce cyclomatic complexity
func (g *schemaGenerator) generateDeclaredType(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if decl, ok := g.output.declsBySchema[t]; ok {
		if t.Dereferenced {
			if decl.Name != scope.string() {
				if existingDecl, exists := g.output.declsByName[scope.string()]; exists &&
					existingDecl != decl && existingDecl.Type != nil {
					return &codegen.NamedType{Decl: decl}, nil
				}

				declAlias := &codegen.AliasType{
					Alias: scope.string(),
					Name:  decl.Name,
				}

				g.generateUnmarshaler(decl, nil)

				g.output.file.Package.AddDecl(declAlias)
			}
		}

		return &codegen.NamedType{Decl: decl}, nil
	}

	if !g.output.isUniqueTypeName(scope.string()) {
		odecl := g.output.declsByName[scope.string()]

		if cmp.Equal(odecl.SchemaType, t, cmputil.Opts(*odecl.SchemaType, *t)...) {
			return &codegen.NamedType{Decl: odecl}, nil
		}

		if odecl := g.output.declsBySchema[t]; odecl != nil {
			return &codegen.NamedType{Decl: odecl}, nil
		}

		if odecl := g.output.getDeclByEqualSchema(scope.string(), t); odecl != nil {
			return &codegen.NamedType{Decl: odecl}, nil
		}
	}

	if t.Enum != nil {
		return g.generateEnumType(t, scope)
	}

	name := g.output.uniqueTypeName(scope)

	if g.config.StructNameFromTitle && t.Title != "" {
		name = g.caser.Identifierize(t.Title)
	}

	decl := codegen.TypeDecl{
		Name:       name,
		Comment:    t.Description,
		SchemaType: t,
	}
	g.output.declsBySchema[t] = &decl
	g.output.declsByName[decl.Name] = &decl

	theType, err := g.generateType(t, scope)
	if err != nil {
		return nil, err
	}

	if isNamedType(theType) || (isMapType(theType) && g.config.DisableCustomTypesForMaps) {
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

	var validators []validator

	switch tt := theType.(type) {
	case *codegen.StructType:
		if t.GetSubSchemaType() == schemas.SubSchemaTypeAnyOf {
			validators = append(validators, &anyOfValidator{decl.Name, t.GetSubSchemasCount()})
			g.generateUnmarshaler(&decl, validators)

			return &codegen.NamedType{Decl: &decl}, nil
		}

		// Detect oneOf-envelope pattern: one or more struct properties carry
		// x-go-oneof-envelope.
		envFields := findOneOfEnvelopeFields(t)
		if len(envFields) > 0 {
			for _, envField := range envFields {
				ext := envField.prop.GoOneOfEnvelope
				if len(envField.prop.OneOf) == 0 || ext == nil || ext.Discriminator == "" || len(ext.Mapping) == 0 {
					return nil, fmt.Errorf(
						"x-go-oneof-envelope on field %q requires oneOf, discriminator, and mapping",
						envField.jsonName,
					)
				}
			}
		}

		for _, f := range tt.RequiredJSONFields {
			validators = append(validators, &requiredValidator{f, decl.Name})
		}

		for _, f := range tt.Fields {
			if f.DefaultValue != nil {
				if f.Name == additionalProperties {
					g.output.file.Package.AddImport("reflect", "")
					g.output.file.Package.AddImport("strings", "")
					g.output.file.Package.AddImport("github.com/go-viper/mapstructure/v2", "")
				}

				dvType := f.Type
				dvIsPointer := false

				if pt, ok := f.Type.(*codegen.PointerType); ok {
					dvType = pt.Type
					dvIsPointer = true
				}

				validators = append(validators, &defaultValidator{
					jsonName:         f.JSONName,
					fieldName:        f.Name,
					defaultValueType: dvType,
					defaultValue:     f.DefaultValue,
					isPointer:        dvIsPointer,
				})
			}

			if !g.config.DisableReadOnlyValidation && f.SchemaType.ReadOnly {
				validators = append(validators, &readOnlyValidator{
					jsonName: f.JSONName,
					declName: decl.Name,
				})
			}

			validators = g.structFieldValidators(validators, f, f.Type, false)
		}

		if len(envFields) > 0 {
			g.addValidatorImports(validators)
			g.generateEnvelopeOuterUnmarshal(&decl, t, envFields, validators)

			if g.config.ExtraImports {
				g.generateEnvelopeOuterUnmarshalYAML(&decl)
			}
		} else if t.IsSubSchemaTypeElem() || len(validators) > 0 {
			g.generateUnmarshaler(&decl, validators)
		}

	case codegen.PrimitiveType, *codegen.PrimitiveType:
		validators = g.structFieldValidators(nil, codegen.StructField{
			Type:       tt,
			SchemaType: t,
		}, tt, false)

		if t.IsSubSchemaTypeElem() || len(validators) > 0 {
			g.generateUnmarshaler(&decl, validators)
		}

	case codegen.MapType, *codegen.MapType:
		if t.IsSubSchemaTypeElem() {
			g.generateUnmarshaler(&decl, []validator{})
		}
	}

	return &codegen.NamedType{Decl: &decl}, nil
}

//nolint:gocyclo // todo: reduce cyclomatic complexity
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
		switch {
		case v.Type == schemas.TypeNameString:
			hasPattern := len(f.SchemaType.Pattern) != 0
			if f.SchemaType.MinLength != 0 || f.SchemaType.MaxLength != 0 || hasPattern || f.SchemaType.Const != nil {
				// Double escape the escape characters so we don't effectively parse the escapes within the value.
				escapedPattern := f.SchemaType.Pattern

				var constVal *string

				if f.SchemaType.Const != nil {
					if s, ok := f.SchemaType.Const.(string); ok {
						constVal = &s
					} else {
						g.warner(fmt.Sprintf("Ignoring non string const value: %v", f.SchemaType.Const))
					}
				}

				replaceJSONCharactersBy := []string{"\\b", "\\f", "\\n", "\\r", "\\t"}

				replaceJSONCharacters := []string{"\b", "\f", "\n", "\r", "\t"}
				for i, replace := range replaceJSONCharacters {
					with := replaceJSONCharactersBy[i]
					escapedPattern = strings.ReplaceAll(escapedPattern, replace, with)
				}

				validators = append(validators, &stringValidator{
					jsonName:   f.JSONName,
					fieldName:  f.Name,
					minLength:  f.SchemaType.MinLength,
					maxLength:  f.SchemaType.MaxLength,
					pattern:    escapedPattern,
					constVal:   constVal,
					isNillable: isNillable,
				})
			}

			if hasPattern {
				g.output.file.Package.AddImport("regexp", "")
			}

		case strings.Contains(v.Type, "int") || v.Type == float64Type:
			if f.SchemaType.MultipleOf != nil ||
				f.SchemaType.Maximum != nil ||
				f.SchemaType.ExclusiveMaximum != nil ||
				f.SchemaType.Minimum != nil ||
				f.SchemaType.ExclusiveMinimum != nil ||
				f.SchemaType.Const != nil {
				validators = append(validators, &numericValidator{
					jsonName:         f.JSONName,
					fieldName:        f.Name,
					isNillable:       isNillable,
					multipleOf:       f.SchemaType.MultipleOf,
					maximum:          f.SchemaType.Maximum,
					exclusiveMaximum: f.SchemaType.ExclusiveMaximum,
					minimum:          f.SchemaType.Minimum,
					exclusiveMinimum: f.SchemaType.ExclusiveMinimum,
					constVal:         f.SchemaType.Const,
					roundToInt:       strings.Contains(v.Type, "int"),
				})
			}

			if f.SchemaType.MultipleOf != nil && v.Type == float64Type {
				g.output.file.Package.AddImport("math", "")
			}

		case v.Type == "bool":
			if f.SchemaType.Const != nil {
				var constVal *bool

				if f.SchemaType.Const != nil {
					if b, ok := f.SchemaType.Const.(bool); ok {
						constVal = &b
					} else {
						g.warner(fmt.Sprintf("Ignoring non boolean const value: %v", f.SchemaType.Const))
					}
				}

				validators = append(validators, &booleanValidator{
					jsonName:   f.JSONName,
					fieldName:  f.Name,
					isNillable: isNillable,
					constVal:   constVal,
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

func (g *schemaGenerator) addValidatorImports(validators []validator) {
	for _, v := range validators {
		if _, ok := v.(*anyOfValidator); ok {
			g.output.file.Package.AddImport("errors", "")
		}

		if sv, ok := v.(*stringValidator); ok && (sv.minLength != 0 || sv.maxLength != 0) {
			g.output.file.Package.AddImport("unicode/utf8", "")
		}

		for _, pkg := range v.desc().imports {
			g.output.file.Package.AddImport(pkg.qualifiedName, "")
		}

		if v.desc().hasError {
			g.output.file.Package.AddImport("fmt", "")
		}
	}
}

func (g *schemaGenerator) generateUnmarshaler(decl *codegen.TypeDecl, validators []validator) {
	if g.config.OnlyModels {
		return
	}

	if hasUnmarshallers, ok := g.output.unmarshallersByTypeDecl[decl]; ok || hasUnmarshallers {
		return
	}

	defer func() {
		g.output.unmarshallersByTypeDecl[decl] = true
	}()

	g.addValidatorImports(validators)

	for _, formatter := range g.formatters {
		formatter.addImport(g.output.file, decl)

		g.output.file.Package.AddDecl(&codegen.Method{
			Impl: formatter.generate(g.output, decl, validators),
			Name: decl.GetName() + "_validator_" + formatter.getName(),
		})
	}
}

func (g *schemaGenerator) generateType(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if ext := t.GoJSONSchemaExtension; ext != nil {
		for _, pkg := range ext.Imports {
			g.output.file.Package.AddImport(pkg, "")
		}

		if ext.Type != nil {
			return &codegen.CustomNameType{Type: *ext.Type, Nillable: ext.Nillable}, nil
		}
	}

	if t.Enum != nil {
		return g.generateEnumType(t, scope)
	}

	if t.Ref != "" {
		return g.generateReferencedType(t)
	}

	typeName, typePtr := g.determineTypeName(t)

	switch typeName {
	case schemas.TypeNameArray:
		if t.Items == nil {
			return arrayTypeVal, nil
		}

		elemType, err := g.generateType(t.Items, g.singularScope(scope))
		if err != nil {
			return nil, err
		}

		return codegen.ArrayType{Type: elemType}, nil

	case schemas.TypeNameObject:
		return g.generateStructType(t, scope)

	case schemas.TypeNameNull:
		return emptyInterfaceTypeVal, nil

	default:
		cg, err := codegen.PrimitiveTypeFromJSONSchemaType(
			typeName,
			t.Format,
			typePtr,
			g.config.MinSizedInts,
			&t.Minimum,
			&t.Maximum,
			&t.ExclusiveMinimum,
			&t.ExclusiveMaximum,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid type %q: %w", typeName, err)
		}

		if ncg, ok := cg.(codegen.NamedType); ok {
			for _, imprt := range ncg.Package.Imports {
				g.output.file.Package.AddImport(imprt.QualifiedName, "")
			}

			return ncg, nil
		}

		if pcg, ok := cg.(*codegen.PointerType); ok {
			if ncg, ok := pcg.Type.(codegen.NamedType); ok {
				for _, imprt := range ncg.Package.Imports {
					g.output.file.Package.AddImport(imprt.QualifiedName, "")
				}

				return pcg, nil
			}
		}

		if dcg, ok := cg.(codegen.DurationType); ok {
			g.output.file.Package.AddImport("time", "")

			return dcg, nil
		}

		return cg, nil
	}
}

func (g *schemaGenerator) determineTypeName(t *schemas.Type) (string, bool) {
	if len(t.Type) == 0 {
		if len(t.AnyOf) == 0 && len(t.AllOf) == 0 {
			return schemas.TypeNameNull, false
		}

		if len(t.AnyOf) != 0 {
			refType := t.AnyOf[0]

			for k, v := range t.AnyOf {
				if k == 0 {
					continue
				}

				if !refType.Type.Equals(v.Type) {
					return schemas.TypeNameNull, false
				}
			}

			return g.determineTypeName(refType)
		}

		if len(t.AllOf) != 0 {
			refType := t.AllOf[0]

			for k, v := range t.AllOf {
				if k == 0 {
					continue
				}

				if !refType.Type.Equals(v.Type) {
					return schemas.TypeNameNull, false
				}
			}

			return g.determineTypeName(refType)
		}

		return schemas.TypeNameNull, false
	}

	if len(t.Type) == 1 {
		return t.Type[0], false
	}

	if len(t.Type) == 2 {
		tidx := 0
		isPtr := false

		for k, v := range t.Type {
			if v == "null" {
				isPtr = true

				continue
			}

			tidx = k
		}

		return t.Type[tidx], isPtr
	}

	g.warner("Property has multiple types; will be represented as interface{} with no validation")

	return schemas.TypeNameNull, false
}

func (g *schemaGenerator) generateStructType(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if len(t.Properties) == 0 && len(t.AllOf) == 0 && len(t.AnyOf) == 0 {
		if len(t.Required) > 0 {
			g.warner("Object type with no properties has required fields; " +
				"skipping validation code for them since we don't know their types")
		}

		valueType := codegen.Type(emptyInterfaceTypeVal)

		var err error

		if t.AdditionalProperties != nil {
			if valueType, err = g.generateType(t.AdditionalProperties, scope.add("Value")); err != nil {
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

	for _, name := range sortedKeys(t.Properties) {
		if err := g.addStructField(&structType, t, scope, name, uniqueNames, requiredNames); err != nil {
			return nil, err
		}
	}

	if len(t.AnyOf) > 0 {
		return g.generateAnyOfType(t, scope)
	}

	if len(t.AllOf) > 0 {
		return g.generateAllOfType(t, scope)
	}

	// Checking .Not here because `false` is unmarshalled to .Not = Type{}.
	if t.AdditionalProperties != nil && t.AdditionalProperties.Not == nil {
		var (
			defaultValue any          = nil
			fieldType    codegen.Type = emptyInterfaceTypeVal
		)

		if len(t.AdditionalProperties.Type) == 1 {
			switch t.AdditionalProperties.Type[0] {
			case schemas.TypeNameString:
				defaultValue = map[string]string{}
				fieldType = codegen.MapType{
					KeyType:   stringTypeVal,
					ValueType: stringTypeVal,
				}

			case schemas.TypeNameArray:
				defaultValue = map[string][]any{}
				fieldType = codegen.MapType{
					KeyType:   stringTypeVal,
					ValueType: arrayTypeVal,
				}

			case schemas.TypeNameNumber:
				defaultValue = map[string]float64{}
				fieldType = codegen.MapType{
					KeyType:   stringTypeVal,
					ValueType: codegen.PrimitiveType{Type: float64Type},
				}

			case schemas.TypeNameInteger:
				defaultValue = map[string]int{}
				fieldType = codegen.MapType{
					KeyType:   stringTypeVal,
					ValueType: intTypeVal,
				}

			case schemas.TypeNameBoolean:
				defaultValue = map[string]bool{}
				fieldType = codegen.MapType{
					KeyType:   stringTypeVal,
					ValueType: boolTypeVal,
				}

			default:
				defaultValue = map[string]any{}
				fieldType = codegen.MapType{
					KeyType:   stringTypeVal,
					ValueType: emptyInterfaceTypeVal,
				}
			}
		}

		structType.AddField(
			codegen.StructField{
				Name:         additionalProperties,
				DefaultValue: defaultValue,
				SchemaType:   &schemas.Type{},
				Type:         fieldType,
				Tags:         "mapstructure:\",remain\"",
			},
		)
	}

	if t.Default != nil {
		structType.DefaultValue = g.defaultPropertyValue(t)
	}

	return &structType, nil
}

func (g *schemaGenerator) addStructField(
	structType *codegen.StructType,
	t *schemas.Type,
	scope nameScope,
	name string,
	uniqueNames map[string]int,
	requiredNames map[string]bool,
) error {
	prop := t.Properties[name]
	isRequired := requiredNames[name]

	fieldName := g.caser.Identifierize(name)

	var extraTags []string

	if ext := prop.GoJSONSchemaExtension; ext != nil {
		for _, pkg := range ext.Imports {
			g.output.file.Package.AddImport(pkg, "")
		}

		if ext.Identifier != nil {
			fieldName = *ext.Identifier
		}

		for tagKey, tagVal := range ext.ExtraTags {
			extraTags = append(extraTags, fmt.Sprintf(`%s:"%s"`, tagKey, tagVal))
		}
	}

	slices.Sort(extraTags)

	if count, ok := uniqueNames[fieldName]; ok {
		uniqueNames[fieldName] = count + 1
		fieldName = fmt.Sprintf("%s_%d", fieldName, count+1)
		g.warner(fmt.Sprintf("Field %q maps to a field by the same name declared "+
			"in the same struct; it will be declared as %s", name, fieldName))
	} else {
		uniqueNames[fieldName] = 1
	}

	comment := prop.Description
	if comment == "" {
		comment = fmt.Sprintf("%s corresponds to the JSON schema field %q.", fieldName, name)
	}

	schemaTypeForField := prop
	if resolvedRefSchema, shouldSemanticInlineRef := g.resolveStructFieldSchemaType(prop); shouldSemanticInlineRef {
		schemaTypeForField = resolvedRefSchema
	}

	structFieldType, err := g.generateStructFieldType(schemaTypeForField, scope.add(fieldName), isRequired)
	if err != nil {
		return fmt.Errorf("cannot add struct field: %w", err)
	}

	structField := codegen.StructField{
		Name:         fieldName,
		Comment:      comment,
		JSONName:     name,
		SchemaType:   schemaTypeForField,
		Tags:         g.generateStructFieldTags(name, extraTags, isRequired),
		DefaultValue: g.generateStructFieldDefaultValue(prop),
		Type:         structFieldType,
	}

	g.appendStructRequiredJSONFields(&structField, structType, isRequired)

	structType.AddField(structField)

	return nil
}

func (g *schemaGenerator) resolveStructFieldSchemaType(prop *schemas.Type) (*schemas.Type, bool) {
	if prop.Ref == "" {
		return prop, false
	}

	defName, fileName, err := g.extractRefNames(prop)
	if err != nil {
		return prop, false
	}

	// For internal refs (#/$defs/...), resolve directly from schema definitions
	// and apply the same transparency checks as for external refs.  This makes
	// extracting a primitive schema into $defs behave equivalently to leaving it
	// inline, which is the intended $ref semantic-transparency policy.
	//
	// Phase 2: the wouldProducePlainGoType guard is extended to also allow
	// inlining when the backend type is a named/value Go type (e.g. time.Time
	// from a date-time format, netip.Addr from ipv4/ipv6).  Specialised integer
	// types from MinSizedInts are still excluded because the shared schema object
	// is mutated during standalone declaration generation (PrimitiveTypeFromJSONSchemaType
	// sets Minimum/Maximum to nil for elided bounds), which would cause the inline
	// path to see corrupted bounds and produce an inconsistent field type.
	if fileName == "" {
		if defName == "" {
			return prop, false
		}

		def, ok := g.schema.Definitions[defName]
		if !ok {
			return prop, false
		}

		if g.shouldKeepReferencedSchemaAsNamedType(def) {
			return prop, false
		}

		shouldSemanticInline := g.isSemanticInlinePrimitiveSchema(def)
		if !shouldSemanticInline && !g.hasLocalRefValidationOverride(prop, def) {
			return prop, false
		}

		// Inline only when the resulting Go type is either a plain primitive (int,
		// float64, string, bool) — already handled by the original check — or a
		// named/value type produced by format-mapping (e.g. time.Time).  Specialised
		// integer types (int8, uint16, …) are excluded to avoid the mutation issue
		// described above.
		if !g.wouldProducePlainGoType(def) && !g.wouldProduceNamedGoType(def) {
			return prop, false
		}

		return g.applyLocalRefValidationOverride(def, prop), true
	}

	// For external refs, load the schema directly rather than calling resolveRef.
	// resolveRef drives AddFile which materialises a full declaration for the
	// referenced schema; that declaration is unused when the field is semantically
	// inlined, and the call also has the side-effect of mutating the schema's
	// numeric bounds (via PrimitiveTypeFromJSONSchemaType), which can suppress
	// validator generation for the inlined field.  By loading the schema directly
	// we avoid both problems while still performing all the same transparency
	// checks.
	//
	// Phase 2: this replaces the previous resolveRef-based external-ref path.

	// Schemas with an x-go-ref mapping are handled by the named-type path.
	// If the referenced schema has an invalid x-go-ref mapping, do not suppress
	// that error by continuing down the semantic-inline path; instead, fall back
	// to the normal $ref handling path, which will surface the validation error.
	if _, _, _, hasRefMapping, mappingErr := g.resolveReferencedXGoRefMappingForRef(prop); mappingErr != nil || hasRefMapping {
		return prop, false
	}

	schema, serr := g.loader.Load(fileName, g.schemaFileName)
	if serr != nil {
		g.warner(fmt.Sprintf("Could not resolve ref %q for field validation/type semantics: %v", prop.Ref, serr))

		return prop, false
	}

	var resolvedRefSchema *schemas.Type
	if defName != "" {
		def, ok := schema.Definitions[defName]
		if !ok {
			return prop, false
		}

		resolvedRefSchema = def
	} else {
		if schema.ObjectAsType == nil {
			return prop, false
		}

		resolvedRefSchema = (*schemas.Type)(schema.ObjectAsType)
	}

	if g.shouldKeepReferencedSchemaAsNamedType(resolvedRefSchema) {
		return prop, false
	}

	shouldSemanticInline := g.isSemanticInlinePrimitiveSchema(resolvedRefSchema)
	if !shouldSemanticInline && !g.hasLocalRefValidationOverride(prop, resolvedRefSchema) {
		return prop, false
	}

	return g.applyLocalRefValidationOverride(resolvedRefSchema, prop), true
}

func (g *schemaGenerator) hasLocalRefValidationOverride(prop, resolvedRefSchema *schemas.Type) bool {
	if prop == nil || resolvedRefSchema == nil {
		return false
	}

	typeIndex, _ := g.isTypeNullable(resolvedRefSchema)
	if typeIndex == -1 {
		return false
	}

	switch resolvedRefSchema.Type[typeIndex] {
	case schemas.TypeNameString:
		return prop.MinLength != 0 ||
			prop.MaxLength != 0 ||
			prop.Pattern != "" ||
			prop.Const != nil
	case schemas.TypeNameInteger, schemas.TypeNameNumber:
		return prop.Minimum != nil ||
			prop.Maximum != nil ||
			prop.ExclusiveMinimum != nil ||
			prop.ExclusiveMaximum != nil ||
			prop.MultipleOf != nil ||
			prop.Const != nil
	case schemas.TypeNameBoolean:
		return prop.Const != nil
	default:
		return false
	}
}

func (g *schemaGenerator) applyLocalRefValidationOverride(resolvedRefSchema, prop *schemas.Type) *schemas.Type {
	if prop == nil || resolvedRefSchema == nil {
		return resolvedRefSchema
	}

	effectiveSchema := *resolvedRefSchema

	typeIndex, _ := g.isTypeNullable(resolvedRefSchema)
	if typeIndex == -1 {
		return &effectiveSchema
	}

	switch resolvedRefSchema.Type[typeIndex] {
	case schemas.TypeNameString:
		if prop.MinLength != 0 {
			effectiveSchema.MinLength = prop.MinLength
		}
		if prop.MaxLength != 0 {
			effectiveSchema.MaxLength = prop.MaxLength
		}
		if prop.Pattern != "" {
			effectiveSchema.Pattern = prop.Pattern
		}
		if prop.Const != nil {
			effectiveSchema.Const = prop.Const
		}
	case schemas.TypeNameInteger, schemas.TypeNameNumber:
		if prop.Minimum != nil {
			effectiveSchema.Minimum = cloneFloat64Ptr(prop.Minimum)
		}
		if prop.Maximum != nil {
			effectiveSchema.Maximum = cloneFloat64Ptr(prop.Maximum)
		}
		if prop.ExclusiveMinimum != nil {
			effectiveSchema.ExclusiveMinimum = cloneAnyPtr(prop.ExclusiveMinimum)
		}
		if prop.ExclusiveMaximum != nil {
			effectiveSchema.ExclusiveMaximum = cloneAnyPtr(prop.ExclusiveMaximum)
		}
		if prop.MultipleOf != nil {
			effectiveSchema.MultipleOf = cloneFloat64Ptr(prop.MultipleOf)
		}
		if prop.Const != nil {
			effectiveSchema.Const = prop.Const
		}
	case schemas.TypeNameBoolean:
		if prop.Const != nil {
			effectiveSchema.Const = prop.Const
		}
	}

	return &effectiveSchema
}

func (g *schemaGenerator) resolveReferencedXGoRefMappingForRef(
	refType *schemas.Type,
) (string, string, string, bool, error) {
	if refType.Ref == "" {
		return "", "", "", false, nil
	}

	defName, fileName, err := g.extractRefNames(refType)
	if err != nil || fileName == "" {
		return "", "", "", false, nil
	}

	schema, err := g.loader.Load(fileName, g.schemaFileName)
	if err != nil {
		return "", "", "", false, nil
	}

	if defName == "" {
		if schema.ObjectAsType == nil {
			return "", "", "", false, nil
		}

		rootTypeName := g.getRootTypeName(schema, fileName)
		rootType := (*schemas.Type)(schema.ObjectAsType)
		rootTypeCopy := *rootType
		rootTypeCopy.Title = rootTypeName

		return g.resolveReferencedXGoRefMapping(
			&rootTypeCopy,
			rootTypeName,
			refType.Ref,
		)
	}

	definition, ok := schema.Definitions[defName]
	if !ok {
		return "", "", "", false, nil
	}

	return g.resolveReferencedXGoRefMapping(definition, g.caser.Identifierize(defName), refType.Ref)
}

func (g *schemaGenerator) resolveReferencedDefinitionTypeName(
	definition *schemas.Type,
	fallback string,
) (string, error) {
	if definition == nil {
		return fallback, nil
	}

	if g.config.StructNameFromTitle && definition.Title != "" {
		return g.caser.Identifierize(definition.Title), nil
	}

	return fallback, nil
}

func (g *schemaGenerator) resolveReferencedXGoRefMapping(
	definition *schemas.Type,
	fallbackName string,
	ref string,
) (string, string, string, bool, error) {
	if definition == nil || definition.XGoRef == nil {
		return "", "", "", false, nil
	}

	goType, err := g.resolveReferencedDefinitionTypeName(definition, fallbackName)
	if err != nil {
		return "", "", "", false, err
	}

	importPath := strings.TrimSpace(definition.XGoRef.Path)
	if importPath == "" {
		return "", "", "", false, fmt.Errorf(
			"x-go-ref.path is required when x-go-ref is present for ref %q",
			ref,
		)
	}

	importAlias := strings.TrimSpace(definition.XGoRef.Alias)
	if importAlias == "" {
		return "", "", "", false, fmt.Errorf(
			"x-go-ref.alias is required when x-go-ref is present for ref %q",
			ref,
		)
	}

	if err := validateGoIdentifier(importAlias, "x-go-ref.alias", ref); err != nil {
		return "", "", "", false, err
	}

	return importAlias + "." + goType, importPath, importAlias, true, nil
}

func validateGoIdentifier(value, partName, ref string) error {
	if goIdentifierRe.MatchString(value) {
		return nil
	}

	return fmt.Errorf(
		"invalid %s %q for ref %q: must be a valid Go identifier",
		partName,
		value,
		ref,
	)
}

func (g *schemaGenerator) shouldKeepReferencedSchemaAsNamedType(schemaType *schemas.Type) bool {
	if schemaType.Title != "" {
		return true
	}

	if ext := schemaType.GoJSONSchemaExtension; ext != nil && ext.Type != nil {
		return true
	}

	return schemaType.XGoType != nil && strings.TrimSpace(*schemaType.XGoType) != ""
}

func (g *schemaGenerator) isSemanticInlinePrimitiveSchema(schemaType *schemas.Type) bool {
	typeIndex, _ := g.isTypeNullable(schemaType)
	if typeIndex == -1 {
		return false
	}

	primitiveType := schemaType.Type[typeIndex]
	if !g.hasSemanticInlinePrimitiveConstraints(schemaType, primitiveType) {
		return false
	}

	return schemas.IsPrimitiveType(primitiveType) && primitiveType != schemas.TypeNameNull
}

func (g *schemaGenerator) hasSemanticInlinePrimitiveConstraints(schemaType *schemas.Type, primitiveType string) bool {
	switch primitiveType {
	case schemas.TypeNameString:
		// A non-empty format (e.g. "date-time" → time.Time, "ipv4" → netip.Addr)
		// signals that the backend Go representation is a named/value type rather
		// than plain string.  Treat any format as a semantic-transparency trigger
		// so that the schema remains inline regardless of whether additional
		// constraints (pattern, minLength, …) are present.
		if schemaType.Format != "" {
			return true
		}

		return schemaType.Pattern != "" ||
			schemaType.MinLength != 0 ||
			schemaType.MaxLength != 0 ||
			schemaType.Const != nil
	case schemas.TypeNameInteger, schemas.TypeNameNumber:
		return schemaType.Minimum != nil ||
			schemaType.Maximum != nil ||
			schemaType.ExclusiveMinimum != nil ||
			schemaType.ExclusiveMaximum != nil ||
			schemaType.MultipleOf != nil ||
			schemaType.Const != nil
	case schemas.TypeNameBoolean:
		return schemaType.Const != nil
	default:
		return false
	}
}

// wouldProducePlainGoType reports whether def would produce a plain Go primitive
// type (int, float64, string, bool) rather than a specialized or named type (e.g.
// int8 from MinSizedInts, or time.Time from format "date-time").
//
// This is computed directly from the schema without relying on declsBySchema, so
// it is order-independent: it gives the correct answer even before the definition
// has been materialised by generateRootType.
//
// Deep copies of the numeric bounds are used so that the computation does not have
// the side-effect of setting schema fields to nil (which PrimitiveTypeFromJSONSchemaType
// does when it determines that bounds can be elided from the generated validator).
func (g *schemaGenerator) wouldProducePlainGoType(def *schemas.Type) bool {
	typeIndex, _ := g.isTypeNullable(def)
	if typeIndex == -1 {
		return false
	}

	primitiveType := def.Type[typeIndex]

	// Use deep copies of the bounds so we don't mutate def.Minimum / def.Maximum.
	minCopy := cloneFloat64Ptr(def.Minimum)
	maxCopy := cloneFloat64Ptr(def.Maximum)
	excMinCopy := cloneAnyPtr(def.ExclusiveMinimum)
	excMaxCopy := cloneAnyPtr(def.ExclusiveMaximum)

	goType, err := codegen.PrimitiveTypeFromJSONSchemaType(
		primitiveType,
		def.Format,
		false, // pointer=false; we only care about the base type
		g.config.MinSizedInts,
		&minCopy,
		&maxCopy,
		&excMinCopy,
		&excMaxCopy,
	)
	if err != nil {
		return false
	}

	pt, ok := goType.(codegen.PrimitiveType)
	if !ok {
		// NamedType (e.g. time.Time for date-time, netip.Addr for ipv4) or other
		// non-plain type: preserve the named declaration.
		return false
	}

	return pt.Type == "int" || pt.Type == "float64" || pt.Type == "string" || pt.Type == "bool"
}

// wouldProduceNamedGoType reports whether def would produce a named (non-primitive)
// Go type such as time.Time (for "date-time" format) or netip.Addr (for "ipv4"/"ipv6").
// Unlike wouldProducePlainGoType, this does not depend on MinSizedInts or numeric
// bounds — it only applies to string schemas whose format maps to a named Go type.
//
// This is used to extend the semantic-inline transparency policy to format-backed
// string schemas (phase 2): a $ref to {type:string,format:date-time} should be
// inlined as time.Time regardless of whether the schema was extracted into $defs or
// an external file, just as a plain-string $ref is inlined as string.
func (g *schemaGenerator) wouldProduceNamedGoType(def *schemas.Type) bool {
	typeIndex, _ := g.isTypeNullable(def)
	if typeIndex == -1 {
		return false
	}

	if def.Type[typeIndex] != schemas.TypeNameString {
		return false
	}

	// Use deep copies of bounds (unused for string, but required by the signature).
	minCopy := cloneFloat64Ptr(def.Minimum)
	maxCopy := cloneFloat64Ptr(def.Maximum)
	excMinCopy := cloneAnyPtr(def.ExclusiveMinimum)
	excMaxCopy := cloneAnyPtr(def.ExclusiveMaximum)

	goType, err := codegen.PrimitiveTypeFromJSONSchemaType(
		schemas.TypeNameString,
		def.Format,
		false,
		g.config.MinSizedInts,
		&minCopy,
		&maxCopy,
		&excMinCopy,
		&excMaxCopy,
	)
	if err != nil {
		return false
	}

	_, ok := goType.(codegen.NamedType)

	return ok
}

// cloneFloat64Ptr returns a pointer to a copy of the float64 value that p points
// to, or nil if p is nil.  Used to prevent PrimitiveTypeFromJSONSchemaType from
// mutating schema fields through the double-pointer it receives.
func cloneFloat64Ptr(p *float64) *float64 {
	if p == nil {
		return nil
	}

	v := *p

	return &v
}

// cloneAnyPtr returns a pointer to a copy of the any value that p points to, or
// nil if p is nil.
func cloneAnyPtr(p *any) *any {
	if p == nil {
		return nil
	}

	v := *p

	return &v
}

func (g *schemaGenerator) generateStructFieldTags(name string, extraTags []string, isRequired bool) string {
	var (
		tagsBuilder strings.Builder
		omitJson    string
		omitRest    string
	)

	if !isRequired {
		if !g.config.DisableOmitEmpty {
			omitJson += ",omitempty"
			omitRest += ",omitempty"
		}

		if !g.config.DisableOmitZero {
			omitJson += ",omitzero"
			// omitRest is not set as omitzero is supported only by the json package
		}
	}

	for _, tag := range g.config.Tags {
		switch tag {
		case "json":
			fmt.Fprintf(&tagsBuilder, `%s:"%s%s" `, tag, name, omitJson)
		default:
			fmt.Fprintf(&tagsBuilder, `%s:"%s%s" `, tag, name, omitRest)
		}
	}

	for _, tag := range extraTags {
		fmt.Fprintf(&tagsBuilder, `%s `, tag)
	}

	return strings.TrimSpace(tagsBuilder.String())
}

func (g *schemaGenerator) generateStructFieldDefaultValue(schemaType *schemas.Type) any {
	if schemaType.Default != nil {
		return g.defaultPropertyValue(schemaType)
	}

	return nil
}

func (g *schemaGenerator) generateStructFieldType(
	schemaType *schemas.Type,
	scope nameScope,
	isRequired bool,
) (codegen.Type, error) {
	fieldType, err := g.generateTypeInline(schemaType, scope)
	if err != nil {
		return nil, fmt.Errorf("could not generate type with scope '%s': %w", scope.string(), err)
	}

	var (
		shouldForcePtrToTrue  bool
		shouldForcePtrToFalse bool
	)

	if ext := schemaType.GoJSONSchemaExtension; ext != nil && ext.Pointer != nil {
		shouldForcePtrToTrue = *ext.Pointer
		shouldForcePtrToFalse = !*ext.Pointer
	}

	if fieldType.IsNillable() {
		return fieldType, nil
	}

	if shouldForcePtrToTrue {
		return codegen.WrapTypeInPointer(fieldType), nil
	}

	if shouldForcePtrToFalse {
		return fieldType, nil
	}

	if !isRequired && schemaType.Default == nil {
		return codegen.WrapTypeInPointer(fieldType), nil
	}

	return fieldType, nil
}

func (g *schemaGenerator) appendStructRequiredJSONFields(
	structField *codegen.StructField,
	structType *codegen.StructType,
	isRequired bool,
) {
	if isRequired {
		structType.RequiredJSONFields = append(structType.RequiredJSONFields, structField.JSONName)
	}
}

func (g *schemaGenerator) generateAnyOfType(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if len(t.AnyOf) == 0 {
		return nil, errEmptyInAnyOf
	}

	if g.config.AliasSingleAllOfAnyOfRefs && len(t.AnyOf) == 1 && t.IsEmptyObject() {
		childType := t.AnyOf[0]
		if childType.Ref != "" {
			resolvedType, err := g.resolveRef(childType)
			if err == nil {
				return g.generateTypeInline(resolvedType, scope)
			} else {
				g.warner(fmt.Sprintf("Could not resolve ref %q: %v", childType.Ref, err))
			}
		}
	}

	isCycle := false
	rAnyOf, hasNull := g.resolveRefs(t.AnyOf, false)

	for i, typ := range rAnyOf {
		// infer type from base if not set
		if len(typ.Type) == 0 {
			typ.Type = append(schemas.TypeList{}, t.Type...)
		}

		typ.SetSubSchemaTypeElem()

		ic, cleanupCycle, cycleErr := g.detectCycle(typ)
		if cycleErr != nil {
			return nil, cycleErr
		}

		defer cleanupCycle()

		if ic {
			isCycle = true

			continue
		}

		if hasNull {
			typ.Type.Add(schemas.TypeNameNull)
		}

		if _, err := g.generateTypeInline(typ, scope.add(fmt.Sprintf("_%d", i))); err != nil {
			return nil, err
		}
	}

	if isCycle {
		return emptyInterfaceTypeVal, nil
	}

	anyOfType, err := schemas.AnyOf(rAnyOf, t)
	if err != nil {
		return nil, fmt.Errorf("could not merge anyOf types: %w", err)
	}

	anyOfType.AnyOf = nil

	return g.generateTypeInline(anyOfType, scope)
}

func (g *schemaGenerator) generateAllOfType(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	if g.config.AliasSingleAllOfAnyOfRefs && len(t.AllOf) == 1 && t.IsEmptyObject() {
		subType := t.AllOf[0]
		if subType.Ref != "" {
			resolvedType, err := g.resolveRef(subType)
			if err == nil {
				return g.generateTypeInline(resolvedType, scope)
			} else {
				g.warner(fmt.Sprintf("Could not resolve subtype ref %q: %v", subType.Ref, err))
			}
		}
	}

	rAllOf, _ := g.resolveRefs(t.AllOf, true)

	allOfType, err := schemas.AllOf(rAllOf, t)
	if err != nil {
		return nil, fmt.Errorf("could not merge allOf types: %w", err)
	}

	allOfType.AllOf = nil

	return g.generateTypeInline(allOfType, scope)
}

func (g *schemaGenerator) defaultPropertyValue(prop *schemas.Type) any {
	if prop.AdditionalProperties != nil {
		if len(prop.AdditionalProperties.Type) == 0 {
			return prop.Default
		}

		if len(prop.AdditionalProperties.Type) != 1 {
			g.warner("Additional property has multiple types; will be represented as an empty interface with no validation")

			return prop.Default
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
			return prop.Default
		}
	}

	return prop.Default
}

//nolint:gocyclo // todo: reduce cyclomatic complexity
func (g *schemaGenerator) generateTypeInline(t *schemas.Type, scope nameScope) (codegen.Type, error) {
	typeIndex, typeIsNullable := g.isTypeNullable(t)

	// Handle x-go-oneof-envelope: generate a union container struct.
	if t.GoOneOfEnvelope != nil && len(t.OneOf) > 0 {
		return g.generateOneOfEnvelopeValueType(t, scope)
	}

	if t.Enum == nil && t.Ref == "" {
		if ext := t.GoJSONSchemaExtension; ext != nil {
			for _, pkg := range ext.Imports {
				g.output.file.Package.AddImport(pkg, "")
			}

			if ext.Type != nil {
				return &codegen.CustomNameType{Type: *ext.Type, Nillable: ext.Nillable}, nil
			}
		}

		if len(t.AnyOf) > 0 {
			return g.generateAnyOfType(t, scope)
		}

		if len(t.AllOf) > 0 {
			return g.generateAllOfType(t, scope)
		}

		if len(t.Type) == 2 && typeIsNullable {
			dt, err := g.generateDeclaredType(t, scope)
			if err != nil {
				return nil, err
			}

			return codegen.WrapTypeInPointer(dt), nil
		}

		if len(t.Type) > 1 && !typeIsNullable {
			g.warner(fmt.Sprintf("Property %v has multiple types; will be represented as interface{} with no validation", scope))

			return emptyInterfaceTypeVal, nil
		}

		if len(t.Type) == 0 {
			return emptyInterfaceTypeVal, nil
		}

		if typeIndex != -1 && schemas.IsPrimitiveType(t.Type[typeIndex]) {
			if t.IsSubSchemaTypeElem() {
				return nil, nil //nolint: nilnil // TODO: this should be fixed, but it requires a rework.
			}

			cg, err := codegen.PrimitiveTypeFromJSONSchemaType(
				t.Type[typeIndex],
				t.Format,
				typeIsNullable,
				g.config.MinSizedInts,
				&t.Minimum,
				&t.Maximum,
				&t.ExclusiveMinimum,
				&t.ExclusiveMaximum,
			)
			if err != nil {
				return nil, fmt.Errorf("invalid type %q: %w", t.Type[typeIndex], err)
			}

			if ncg, ok := cg.(codegen.NamedType); ok {
				for _, imprt := range ncg.Package.Imports {
					g.output.file.Package.AddImport(imprt.QualifiedName, "")
				}

				return ncg, nil
			}

			if pcg, ok := cg.(*codegen.PointerType); ok {
				if ncg, ok := pcg.Type.(codegen.NamedType); ok {
					for _, imprt := range ncg.Package.Imports {
						g.output.file.Package.AddImport(imprt.QualifiedName, "")
					}

					return pcg, nil
				}
			}

			if dcg, ok := cg.(codegen.DurationType); ok {
				g.output.file.Package.AddImport("time", "")

				return dcg, nil
			}

			return cg, nil
		}

		if typeIndex != -1 && t.Type[typeIndex] == schemas.TypeNameArray {
			var theType codegen.Type = emptyInterfaceTypeVal

			if t.Items != nil {
				var err error

				theType, err = g.generateTypeInline(t.Items, g.singularScope(scope))
				if err != nil {
					return nil, err
				}
			}

			return &codegen.ArrayType{Type: theType}, nil
		}

		if typeIndex == -1 {
			return emptyInterfaceTypeVal, nil
		}
	}

	if t.Ref != "" {
		mappedType, importPath, importAlias, hasXGoTypeMapping, err := g.resolveReferencedXGoRefMappingForRef(t)
		if err != nil {
			return nil, err
		}

		if hasXGoTypeMapping {
			if importPath != "" {
				g.output.file.Package.AddImport(importPath, importAlias)
			}

			return &codegen.CustomNameType{Type: mappedType}, nil
		}
	}

	dt, err := g.generateDeclaredType(t, scope)
	if err != nil {
		return nil, err
	}

	if typeIsNullable {
		return codegen.WrapTypeInPointer(dt), nil
	}

	return dt, nil
}

// singularScope attempts to create a name scope for an element of a collection. If the parent collection
// has a plural name like "Actions", then the singular name for the element will be "Action". If the collection
// is not plural, like "WhateverElse", then the element's name will be "WhateverElseElem".
func (g *Generator) singularScope(scope nameScope) nameScope {
	last, ok := scope.last()
	if g.minimalNames && ok && strings.HasSuffix(last, "s") {
		return scope.add(strings.TrimSuffix(last, "s"))
	}

	return scope.add("Elem")
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
		if enumType, err = codegen.PrimitiveTypeFromJSONSchemaType(
			t.Type[0],
			t.Format,
			false,
			g.config.MinSizedInts,
			&t.Minimum,
			&t.Maximum,
			&t.ExclusiveMinimum,
			&t.ExclusiveMaximum,
		); err != nil {
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
					valueType = float64Type

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
		Name:       g.output.uniqueTypeName(scope),
		Type:       enumType,
		SchemaType: t,
	}
	g.output.file.Package.AddDecl(&enumDecl)

	g.output.declsByName[enumDecl.Name] = &enumDecl
	g.output.declsBySchema[t] = &enumDecl

	if !g.config.OnlyModels {
		valueConstant := &codegen.Var{
			Name:  schemas.PrefixEnumValue + enumDecl.Name,
			Value: t.Enum,
		}
		g.output.file.Package.AddDecl(valueConstant)

		g.output.file.Package.AddImport("fmt", "")
		g.output.file.Package.AddImport("reflect", "")

		for _, formatter := range g.formatters {
			formatter.addImport(g.output.file, &enumDecl)

			if wrapInStruct {
				g.output.file.Package.AddDecl(&codegen.Method{
					Impl: formatter.enumMarshal(&enumDecl),
					Name: enumDecl.GetName() + "_enum_" + formatter.getName(),
				})
			}

			g.output.file.Package.AddDecl(&codegen.Method{
				Impl: formatter.enumUnmarshal(enumDecl, enumType, valueConstant, wrapInStruct),
				Name: enumDecl.GetName() + "_enum_unmarshal_" + formatter.getName(),
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

func (g *schemaGenerator) resolveRefs(types []*schemas.Type, withNulls bool) ([]*schemas.Type, bool) {
	resolvedTypes := make([]*schemas.Type, 0, len(types))
	hasNulls := false

	for _, typ := range types {
		resolvedType, err := g.resolveRef(typ)
		if err != nil {
			g.warner(fmt.Sprintf("Could not resolve ref %q: %v", typ.Ref, err))

			continue
		}

		if !withNulls && len(resolvedType.Type) == 1 && resolvedType.Type[0] == schemas.TypeNameNull {
			hasNulls = true

			continue
		}

		resolvedTypes = append(resolvedTypes, resolvedType)
	}

	return resolvedTypes, hasNulls
}

func (g *schemaGenerator) resolveRef(t *schemas.Type) (*schemas.Type, error) {
	if t.Ref == "" {
		return t, nil
	}

	if _, ok := g.schemaTypesByRef[t.Ref]; ok {
		return g.schemaTypesByRef[t.Ref], nil
	}

	typ, err := g.generateReferencedType(t)
	if err != nil {
		return nil, err
	}

	ntyp, err := g.extractPointedType(typ)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCannotResolveRef, err)
	}

	// After resolving the ref type we lose info about the original schema
	// so rewrite all nested refs to include the original schema id
	_, fileName, err := g.extractRefNames(t)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCannotResolveRef, err)
	}

	if fileName != "" {
		err = ntyp.Decl.SchemaType.ConvertAllRefs(fileName)
		if err != nil {
			return nil, fmt.Errorf("convert refs: %w", err)
		}
	}

	ntyp.Decl.SchemaType.Dereferenced = true

	g.schemaTypesByRef[t.Ref] = ntyp.Decl.SchemaType

	return ntyp.Decl.SchemaType, nil
}

func (g *schemaGenerator) extractPointedType(typ codegen.Type) (*codegen.NamedType, error) {
	if rtyp, ok := typ.(*codegen.PointerType); ok {
		if ntyp, ok := rtyp.Type.(*codegen.NamedType); ok {
			return ntyp, nil
		}
	}

	if ntyp, ok := typ.(*codegen.NamedType); ok {
		return ntyp, nil
	}

	return nil, fmt.Errorf("%w: got %T", errExpectedNamedType, typ)
}

func (g *schemaGenerator) detectCycle(t *schemas.Type) (bool, func(), error) {
	defName, filename, err := g.extractRefNames(t)
	if err != nil {
		return false, func() {}, err
	}

	if defName == "" && filename == "" && !t.Dereferenced {
		return false, func() {}, nil
	}

	qual := qualifiedDefinition{
		schema:     g.schema,
		schemaType: t,
		filename:   filename,
		name:       defName,
	}

	_, isCycle := g.inScope[qual]
	if !isCycle {
		g.inScope[qual] = struct{}{}
	}

	return isCycle, func() {
		delete(g.inScope, qual)
	}, nil
}

// isTypeNullable checks if a type is nullable and returns the index of the type array where to find the actual type,
// or -1 if none was found.
func (g *schemaGenerator) isTypeNullable(t *schemas.Type) (int, bool) {
	if len(t.Type) == 1 && (t.Type[0] == schemas.TypeNameArray || t.Type[0] == schemas.TypeNameNull) {
		return 0, true
	}

	if len(t.Type) == 1 && t.Type[0] != schemas.TypeNameNull {
		return 0, false
	}

	if len(t.Type) == 2 && t.Type[0] == schemas.TypeNameNull {
		return 1, true
	}

	if len(t.Type) == 2 && t.Type[1] == schemas.TypeNameNull {
		return 0, true
	}

	if slices.Contains(t.Type, schemas.TypeNameNull) {
		return -1, true
	}

	return -1, false
}
