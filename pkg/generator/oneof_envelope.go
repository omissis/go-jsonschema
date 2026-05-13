package generator

import (
	"fmt"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

// envelopeBranchInfo holds per-branch information for a oneOf envelope type.
type envelopeBranchInfo struct {
	// discVal is the discriminator value that selects this branch (e.g. "a").
	discVal string
	// goField is the Go struct field name in the union container (e.g. "A").
	goField string
	// typeName is the Go type name of the branch (e.g. "AValue").
	typeName string
}

// findOneOfEnvelopeField returns the JSON property name and schema of the first
// property on t that carries a GoOneOfEnvelope extension, if any.
func findOneOfEnvelopeField(t *schemas.Type) (jsonName string, prop *schemas.Type, found bool) {
	for name, p := range t.Properties {
		if p.GoOneOfEnvelope != nil {
			return name, p, true
		}
	}

	return "", nil, false
}

// generateOneOfEnvelopeValueType generates the union container type (e.g. "Payload")
// for a schema property that carries x-go-oneof-envelope.  It also attaches a
// MarshalJSON method to the container.
func (g *schemaGenerator) generateOneOfEnvelopeValueType(
	t *schemas.Type, // value-field schema: has OneOf + GoOneOfEnvelope
	scope nameScope,
) (codegen.Type, error) {
	// Short-circuit if we already generated this type.
	if existing, ok := g.output.declsBySchema[t]; ok {
		return &codegen.NamedType{Decl: existing}, nil
	}

	ext := t.GoOneOfEnvelope

	// Choose the union struct name: prefer title, fall back to scope.
	typeName := scope.string()
	if t.Title != "" {
		typeName = g.caser.Identifierize(t.Title)
	}

	typeName = g.output.uniqueTypeName(newNameScope(typeName))

	// Resolve each mapping entry to a branch type.
	discVals := sortedKeys(ext.Mapping)
	branches := make([]envelopeBranchInfo, 0, len(discVals))

	for _, discVal := range discVals {
		branchTitle := ext.Mapping[discVal]

		// Locate the oneOf branch whose (resolved) title matches branchTitle.
		var matchedBranch *schemas.Type // original branch (may carry a $ref)

		for _, branch := range t.OneOf {
			if branch.Ref != "" {
				resolved, err := g.resolveRef(branch)
				if err != nil {
					return nil, fmt.Errorf("cannot resolve oneOf $ref %q: %w", branch.Ref, err)
				}

				if resolved.Title == branchTitle {
					matchedBranch = branch

					break
				}
			} else if branch.Title == branchTitle {
				matchedBranch = branch

				break
			}
		}

		if matchedBranch == nil {
			return nil, fmt.Errorf(
				"x-go-oneof-envelope: no oneOf branch found with title %q (for discriminator value %q)",
				branchTitle, discVal,
			)
		}

		// Generate (or look up) the branch type.
		var branchType codegen.Type
		var err error

		branchScope := newNameScope(g.caser.Identifierize(branchTitle))

		if matchedBranch.Ref != "" {
			branchType, err = g.generateReferencedType(matchedBranch)
		} else {
			branchType, err = g.generateDeclaredType(matchedBranch, branchScope)
		}

		if err != nil {
			return nil, fmt.Errorf("cannot generate oneOf branch type %q: %w", branchTitle, err)
		}

		// Unwrap pointer if needed to get the NamedType.
		var branchDecl *codegen.TypeDecl

		switch v := branchType.(type) {
		case *codegen.NamedType:
			branchDecl = v.Decl
		case *codegen.PointerType:
			if nt, ok := v.Type.(*codegen.NamedType); ok {
				branchDecl = nt.Decl
			}
		}

		if branchDecl == nil {
			return nil, fmt.Errorf(
				"oneOf branch %q did not produce a named type (got %T)", branchTitle, branchType,
			)
		}

		branches = append(branches, envelopeBranchInfo{
			discVal:  discVal,
			goField:  g.caser.Identifierize(discVal),
			typeName: branchDecl.Name,
		})
	}

	// Build the union container struct.
	structType := &codegen.StructType{}

	for _, b := range branches {
		branchDecl := g.output.declsByName[b.typeName]

		structType.AddField(codegen.StructField{
			Name:       b.goField,
			Type:       &codegen.PointerType{Type: &codegen.NamedType{Decl: branchDecl}},
			SchemaType: &schemas.Type{},
		})
	}

	decl := &codegen.TypeDecl{
		Name:       typeName,
		Type:       structType,
		SchemaType: t,
	}

	g.output.declsBySchema[t] = decl
	g.output.declsByName[typeName] = decl
	g.output.file.Package.AddDecl(decl)

	// Prevent the standard unmarshaler from being generated for this type.
	g.output.unmarshallersByTypeDecl[decl] = true

	// Attach MarshalJSON.
	g.generateOneOfEnvelopeMarshal(decl, branches)

	return &codegen.NamedType{Decl: decl}, nil
}

// generateOneOfEnvelopeMarshal attaches a MarshalJSON to the union container type.
// It marshals whichever of the pointer fields is non-nil; if the count is not
// exactly one it returns an error.
//
// Note: UnmarshalJSON is intentionally NOT generated for the union container.
// Decoding depends on the discriminator field of the outer envelope struct, so
// the routing entry point is always the outer type's UnmarshalJSON. Generating
// a standalone UnmarshalJSON on the container would require guessing the branch
// without the discriminator context and could silently accept ambiguous payloads.
func (g *schemaGenerator) generateOneOfEnvelopeMarshal(
	decl *codegen.TypeDecl,
	branches []envelopeBranchInfo,
) {
	g.output.file.Package.AddImport("encoding/json", "")
	g.output.file.Package.AddImport("errors", "")

	capturedBranches := branches // capture for closure

	g.output.file.Package.AddDecl(&codegen.Method{
		Name: decl.Name + "_envelope_marshal_json",
		Impl: func(out *codegen.Emitter) error {
			out.Commentf("MarshalJSON implements json.Marshaler.")
			out.Printlnf("func (j %s) MarshalJSON() ([]byte, error) {", decl.Name)
			out.Indent(1)
			out.Printlnf("var count int")
			out.Printlnf("var value interface{}")

			for _, b := range capturedBranches {
				out.Printlnf("if j.%s != nil {", b.goField)
				out.Indent(1)
				out.Printlnf("count++")
				out.Printlnf("value = j.%s", b.goField)
				out.Indent(-1)
				out.Printlnf("}")
			}

			out.Printlnf("if count != 1 {")
			out.Indent(1)
			out.Printlnf(`return nil, errors.New("%s: exactly one branch must be non-nil")`, decl.Name)
			out.Indent(-1)
			out.Printlnf("}")
			out.Printlnf("return json.Marshal(value)")
			out.Indent(-1)
			out.Printlnf("}")

			return nil
		},
	})
}

// generateEnvelopeOuterUnmarshal attaches a custom UnmarshalJSON to the outer
// struct (e.g. OneOfEnvelope).  It:
//  1. Validates required fields.
//  2. Decodes the discriminator value.
//  3. Re-marshals the raw "value" payload for a second pass.
//  4. Uses a type-alias unmarshal for all regular fields.
//  5. Routes the value payload through the correct branch based on the discriminator.
func (g *schemaGenerator) generateEnvelopeOuterUnmarshal(
	decl *codegen.TypeDecl,
	schemaType *schemas.Type,
	envJSONName string, // e.g. "value"
	envGoName string, // e.g. "Value"
	envFieldSchema *schemas.Type,
) {
	ext := envFieldSchema.GoOneOfEnvelope

	// Look up the Payload type and its branch fields.
	payloadDecl := g.output.declsBySchema[envFieldSchema]

	var branches []envelopeBranchInfo

	discVals := sortedKeys(ext.Mapping)

	for _, discVal := range discVals {
		goField := g.caser.Identifierize(discVal)

		// Look up the concrete type name from the Payload struct fields.
		var typeName string

		if payloadDecl != nil {
			if ps, ok := payloadDecl.Type.(*codegen.StructType); ok {
				for _, pf := range ps.Fields {
					if pf.Name == goField {
						if pt, ok2 := pf.Type.(*codegen.PointerType); ok2 {
							if nt, ok3 := pt.Type.(*codegen.NamedType); ok3 {
								typeName = nt.Decl.Name
							}
						}

						break
					}
				}
			}
		}

		if typeName == "" {
			// Fallback: identifierize the branch title.
			typeName = g.caser.Identifierize(ext.Mapping[discVal])
		}

		branches = append(branches, envelopeBranchInfo{
			discVal:  discVal,
			goField:  goField,
			typeName: typeName,
		})
	}

	payloadTypeName := ""
	if payloadDecl != nil {
		payloadTypeName = payloadDecl.Name
	}

	// Required fields (for presence check).
	requiredFields := schemaType.Required
	discJSONName := ext.Discriminator
	discGoName := g.caser.Identifierize(discJSONName)
	discTypeName := ""
	useEnumRouting := false

	// Try to resolve the generated Go field/type for discriminator routing.
	if outerStruct, ok := decl.Type.(*codegen.StructType); ok {
		for _, sf := range outerStruct.Fields {
			if sf.JSONName != discJSONName {
				continue
			}

			discGoName = sf.Name

			switch ft := sf.Type.(type) {
			case *codegen.NamedType:
				discTypeName = ft.Decl.Name
			case *codegen.PointerType:
				if nt, ok := ft.Type.(*codegen.NamedType); ok {
					discTypeName = nt.Decl.Name
				}
			}

			break
		}
	}

	// Use enum constants for switch routing only when discriminator is an enum.
	if discProp, ok := schemaType.Properties[discJSONName]; ok && discTypeName != "" && len(discProp.Enum) > 0 {
		useEnumRouting = true
	}

	capturedBranches := branches
	capturedPayloadTypeName := payloadTypeName
	capturedRequired := requiredFields
	capturedDiscJSON := discJSONName
	capturedDiscGoName := discGoName
	capturedDiscTypeName := discTypeName
	capturedUseEnumRouting := useEnumRouting
	capturedEnvJSONName := envJSONName
	capturedEnvGoName := envGoName
	capturedDeclName := decl.Name

	g.output.file.Package.AddImport("encoding/json", "")
	g.output.file.Package.AddImport("fmt", "")

	g.output.unmarshallersByTypeDecl[decl] = true

	g.output.file.Package.AddDecl(&codegen.Method{
		Name: decl.Name + "_envelope_json",
		Impl: func(out *codegen.Emitter) error {
			out.Commentf("UnmarshalJSON implements json.Unmarshaler.")
			out.Printlnf("func (j *%s) UnmarshalJSON(b []byte) error {", capturedDeclName)
			out.Indent(1)

			// raw is always needed for discriminator and value extraction.
			out.Printlnf("var raw map[string]interface{}")
			out.Printlnf("if err := json.Unmarshal(b, &raw); err != nil { return err }")

			// --- required-field checks ---
			for _, req := range capturedRequired {
				out.Printlnf(
					`if _, ok := raw["%s"]; raw != nil && !ok {`, req,
				)
				out.Indent(1)
				out.Printlnf(
					`return fmt.Errorf("field %s in %s: required")`,
					req, capturedDeclName,
				)
				out.Indent(-1)
				out.Printlnf("}")
			}

			// --- grab raw value payload ---
			out.Printlnf("valueRaw, err := json.Marshal(raw[%q])", capturedEnvJSONName)
			out.Printlnf("if err != nil { return err }")

			// --- decode all regular fields via type-alias trick ---
			out.Printlnf("type Plain %s", capturedDeclName)
			out.Printlnf("var plain Plain")
			out.Printlnf("if err := json.Unmarshal(b, &plain); err != nil { return err }")
			out.Printlnf("*j = %s(plain)", capturedDeclName)
			out.Printlnf("if validator, ok := interface{}(j).(interface{ Validate() error }); ok {")
			out.Indent(1)
			out.Printlnf("if err := validator.Validate(); err != nil { return err }")
			out.Indent(-1)
			out.Printlnf("}")

			if capturedUseEnumRouting {
				out.Printlnf("discriminator := j.%s", capturedDiscGoName)
			} else {
				out.Printlnf("discriminator, _ := raw[%q].(string)", capturedDiscJSON)
			}

			// --- discriminator routing ---
			if capturedUseEnumRouting {
				out.Printlnf("switch j.%s {", capturedDiscGoName)
			} else {
				out.Printlnf("switch discriminator {")
			}

			for _, b := range capturedBranches {
				if capturedUseEnumRouting {
					out.Printlnf("case %s:", g.makeEnumConstantName(capturedDiscTypeName, b.discVal))
				} else {
					out.Printlnf("case %q:", b.discVal)
				}
				out.Indent(1)
				out.Printlnf("var v %s", b.typeName)
				out.Printlnf("if err := json.Unmarshal(valueRaw, &v); err != nil {")
				out.Indent(1)
				out.Printlnf(
					`return fmt.Errorf("%s: invalid value for type %%q: %%w", discriminator, err)`,
					capturedDeclName,
				)
				out.Indent(-1)
				out.Printlnf("}")
				out.Printlnf("j.%s = %s{%s: &v}", capturedEnvGoName, capturedPayloadTypeName, b.goField)
				out.Indent(-1)
			}

			out.Printlnf("default:")
			out.Indent(1)
			out.Printlnf(
				`return fmt.Errorf("%s: unknown discriminator value %%q", discriminator)`,
				capturedDeclName,
			)
			out.Indent(-1)
			out.Printlnf("}")

			out.Printlnf("return nil")
			out.Indent(-1)
			out.Printlnf("}")

			return nil
		},
	})
}

// generateEnvelopeOuterUnmarshalYAML generates a UnmarshalYAML that bridges
// YAML decoding through the custom UnmarshalJSON so that discriminator routing
// is preserved even when deserializing from YAML sources.
func (g *schemaGenerator) generateEnvelopeOuterUnmarshalYAML(decl *codegen.TypeDecl) {
	g.output.file.Package.AddImport("encoding/json", "")

	capturedDeclName := decl.Name

	g.output.file.Package.AddDecl(&codegen.Method{
		Name: decl.Name + "_envelope_yaml",
		Impl: func(out *codegen.Emitter) error {
			out.Commentf("UnmarshalYAML implements yaml.Unmarshaler.")
			out.Printlnf(
				"func (j *%s) UnmarshalYAML(value *yaml.Node) error {",
				capturedDeclName,
			)
			out.Indent(1)
			out.Printlnf("var raw interface{}")
			out.Printlnf("if err := value.Decode(&raw); err != nil { return err }")
			out.Printlnf("jsonData, err := json.Marshal(raw)")
			out.Printlnf("if err != nil { return err }")
			out.Printlnf("return json.Unmarshal(jsonData, j)")
			out.Indent(-1)
			out.Printlnf("}")

			return nil
		},
	})
}
