package generator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sanity-io/litter"
	"github.com/sosodev/duration"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/mathutils"
)

const typeInt = "int"

var (
	ErrDurationIsEmpty                = errors.New("duration default value must not be an empty string")
	ErrCannotConvertISO8601ToGoFormat = errors.New("could not convert duration from ISO8601 to Go format")
	ErrInvalidDefaultValue            = errors.New("invalid default value")
	ErrCannotFindSlideToDump          = errors.New("didn't find a slice to dump")
	ErrDefaultDurationIsNotAString    = errors.New("duration default value must be a string")
)

type validator interface {
	generate(out *codegen.Emitter, format string) error
	desc() *validatorDesc
}

type packageImport struct {
	qualifiedName string
}

type validatorDesc struct {
	hasError            bool
	beforeJSONUnmarshal bool
	requiresRawAfter    bool
	imports             []packageImport
	// decls lists package-level declarations the validator needs (e.g.
	// precompiled regex vars). The orchestrator adds them via
	// Package.AddDecl, which dedupes by name — so multiple validators sharing
	// the same regex emit the var only once per output file.
	decls []codegen.Decl
}

var (
	_ validator = new(requiredValidator)
	_ validator = new(readOnlyValidator)
	_ validator = new(nullTypeValidator)
	_ validator = new(defaultValidator)
	_ validator = new(arrayValidator)
	_ validator = new(stringValidator)
	_ validator = new(numericValidator)
	_ validator = new(anyOfValidator)
	_ validator = new(formatValidator)

	ErrCannotDumpDefaultSlice = errors.New("cannot dump default slice")
)

type requiredValidator struct {
	jsonName string
	declName string
}

func (v *requiredValidator) generate(out *codegen.Emitter, format string) error {
	// The container itself may be null (if the type is ["null", "object"]), in which case
	// the map will be nil and none of the properties are present. This shouldn't fail
	// the validation, though, as that's allowed as long as the container is allowed to be null.
	out.Printlnf(`if _, ok := %s["%s"]; %s != nil && !ok {`, varNameRawMap, v.jsonName, varNameRawMap)
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("field %s in %s: required")`, v.jsonName, v.declName)
	out.Indent(-1)
	out.Printlnf("}")

	return nil
}

func (v *requiredValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: true,
	}
}

type readOnlyValidator struct {
	jsonName string
	declName string
}

func (v *readOnlyValidator) generate(out *codegen.Emitter, format string) error {
	// The container itself may be null (if the type is ["null", "object"]), in which case
	// the map will be nil and none of the properties are present. This shouldn't fail
	// the validation, though, as that's allowed as long as the container is allowed to be null.
	out.Printlnf(`if _, ok := %s["%s"]; %s != nil && ok {`, varNameRawMap, v.jsonName, varNameRawMap)
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("field %s in %s: read only")`, v.jsonName, v.declName)
	out.Indent(-1)
	out.Printlnf("}")

	return nil
}

func (v *readOnlyValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: true,
	}
}

type nullTypeValidator struct {
	jsonName   string
	fieldName  string
	arrayDepth int
}

func (v *nullTypeValidator) generate(out *codegen.Emitter, format string) error {
	value := getPlainName(v.fieldName)
	fieldName := v.jsonName

	indexes := make([]string, v.arrayDepth)

	for i := range v.arrayDepth {
		index := fmt.Sprintf("i%d", i)
		indexes[i] = index
		out.Printlnf(`for %s := range %s {`, index, value)
		value += fmt.Sprintf("[%s]", index)
		fieldName += "[%d]"

		out.Indent(1)
	}

	fieldName = fmt.Sprintf(`"%s"`, fieldName)
	if len(indexes) > 0 {
		fieldName = fmt.Sprintf(`fmt.Sprintf(%s, %s)`, fieldName, strings.Join(indexes, ", "))
	}

	out.Printlnf(`if %s != nil {`, value)
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("field %%s: must be null", %s)`, fieldName)
	out.Indent(-1)
	out.Printlnf("}")

	for range v.arrayDepth {
		out.Indent(-1)
		out.Printlnf("}")
	}

	return nil
}

func (v *nullTypeValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
		requiresRawAfter:    true,
	}
}

type defaultValidator struct {
	jsonName         string
	fieldName        string
	defaultValueType codegen.Type
	defaultValue     any
	isPointer        bool
}

func (v *defaultValidator) generate(out *codegen.Emitter, format string) error {
	defaultValue, err := v.dumpDefaultValueAssignment(out)
	if err != nil {
		return fmt.Errorf("cannot generate default validator: %w", err)
	}

	out.Printlnf(`if v, ok := %s["%s"]; !ok || v == nil {`, varNameRawMap, v.jsonName)
	out.Indent(1)
	out.Printlnf("%s", defaultValue)
	out.Indent(-1)
	out.Printlnf("}")

	return nil
}

func (v *defaultValidator) dumpDefaultValueAssignment(out *codegen.Emitter) (any, error) {
	if v.defaultValueType != nil {
		if nt, ok := v.defaultValueType.(*codegen.NamedType); ok {
			dvm, ok := v.defaultValue.(map[string]any)
			if ok {
				var b strings.Builder

				for _, k := range sortedKeys(dvm) {
					fmt.Fprintf(&b, "\n%s: %s,", upperFirst(k), litter.Sdump(dvm[k]))
				}

				b.WriteString("\n")

				defaultValue := fmt.Sprintf(`%s{%s}`, nt.Decl.GetName(), b.String())

				return v.assignDefault(out, defaultValue), nil
			}
		}

		if _, ok := v.defaultValueType.(codegen.DurationType); ok {
			defaultDurationISO8601, ok := v.defaultValue.(string)

			if !ok {
				return nil, fmt.Errorf("%w: %T given", ErrDefaultDurationIsNotAString, v.defaultValue)
			}

			if defaultDurationISO8601 == "" {
				return nil, ErrDurationIsEmpty
			}

			duration, err := duration.Parse(defaultDurationISO8601)
			if err != nil {
				return nil, ErrCannotConvertISO8601ToGoFormat
			}

			tmpEmitter := codegen.NewEmitter(out.MaxLineLength())

			defaultValue := "defaultDuration"
			goDurationStr := duration.ToTimeDuration().String()

			tmpEmitter.Printlnf("%s, err := time.ParseDuration(\"%s\")", defaultValue, goDurationStr)
			tmpEmitter.Printlnf("if err != nil {")
			tmpEmitter.Indent(1)
			tmpEmitter.Printlnf(
				"return fmt.Errorf(\"failed to parse the \\\"%s\\\" default value for field %s: %%w\", err)",
				goDurationStr,
				v.jsonName,
			)
			tmpEmitter.Indent(-1)
			tmpEmitter.Printlnf("}")

			if v.isPointer {
				tmpEmitter.Printlnf(`%s.%s = &%s`, varNamePlainStruct, v.fieldName, defaultValue)
			} else {
				tmpEmitter.Printlnf(`%s.%s = %s`, varNamePlainStruct, v.fieldName, defaultValue)
			}

			return tmpEmitter.String(), nil
		}
	}

	if defaultValue, err := v.tryDumpDefaultSlice(out.MaxLineLength()); err == nil {
		return v.assignDefault(out, defaultValue), nil
	}

	// Special handling for pointer-to-integer types (e.g., *int or NamedType wrapping *int).
	// We need to create a temp variable and take its address.
	if v.isPointerToInteger() {
		if f, ok := v.defaultValue.(float64); ok {
			intVal := int(f)
			tmpEmitter := codegen.NewEmitter(out.MaxLineLength())
			tmpEmitter.Printlnf("defaultInt := %d", intVal)
			tmpEmitter.Printlnf(`%s = &defaultInt`, getPlainName(v.fieldName))

			return tmpEmitter.String(), nil
		}
	}

	// Fallback to sdump in case we couldn't dump it properly.
	// Special handling for integer types: JSON numbers are float64, but we need int literals.
	defaultValue := v.defaultValue
	if v.isIntegerType() {
		if f, ok := v.defaultValue.(float64); ok {
			defaultValue = int(f)
		}
	}

	return v.assignDefault(out, litter.Sdump(defaultValue)), nil
}

// assignDefault generates the assignment of a default value to the field.
// When the field is a pointer type, it creates a typed temporary variable
// and assigns its address to the field.
func (v *defaultValidator) assignDefault(out *codegen.Emitter, valueExpr string) string {
	if !v.isPointer {
		return fmt.Sprintf(`%s = %s`, getPlainName(v.fieldName), valueExpr)
	}

	tmpVarName := "default" + v.fieldName
	tmpEmitter := codegen.NewEmitter(out.MaxLineLength())

	// Use a var declaration with explicit type so that untyped constants
	// (e.g. 42.0 from JSON) are correctly converted to the target type.
	typeEmitter := codegen.NewEmitter(out.MaxLineLength())

	if err := v.defaultValueType.Generate(typeEmitter); err == nil {
		typeName := strings.TrimSpace(typeEmitter.String())
		tmpEmitter.Printlnf("var %s %s = %s", tmpVarName, typeName, valueExpr)
	} else {
		tmpEmitter.Printlnf("%s := %s", tmpVarName, valueExpr)
	}

	tmpEmitter.Printlnf("%s = &%s", getPlainName(v.fieldName), tmpVarName)

	return strings.TrimRight(tmpEmitter.String(), "\n")
}

func (v *defaultValidator) isIntegerType() bool {
	return isIntegerType(v.defaultValueType)
}

func isIntegerType(t codegen.Type) bool {
	switch tt := t.(type) {
	case codegen.PointerType:
		return isIntegerType(tt.Type)
	case *codegen.PointerType:
		return isIntegerType(tt.Type)
	case codegen.NamedType:
		return isIntegerType(tt.Decl.Type)
	case *codegen.NamedType:
		return isIntegerType(tt.Decl.Type)
	case codegen.PrimitiveType:
		return tt.Type == typeInt
	}

	return false
}

func (v *defaultValidator) isPointerToInteger() bool {
	return isPointerToInteger(v.defaultValueType)
}

func isPointerToInteger(t codegen.Type) bool {
	switch tt := t.(type) {
	case codegen.NamedType:
		return isPointerToInteger(tt.Decl.Type)
	case *codegen.NamedType:
		return isPointerToInteger(tt.Decl.Type)
	case codegen.PointerType:
		if pt, ok := tt.Type.(codegen.PrimitiveType); ok {
			return pt.Type == typeInt
		}
	case *codegen.PointerType:
		if pt, ok := tt.Type.(codegen.PrimitiveType); ok {
			return pt.Type == typeInt
		}
	}

	return false
}

func (v *defaultValidator) tryDumpDefaultSlice(maxLineLen int32) (string, error) {
	tmpEmitter := codegen.NewEmitter(maxLineLen)

	if err := v.defaultValueType.Generate(tmpEmitter); err != nil {
		return "", fmt.Errorf("%w: %w", ErrCannotDumpDefaultSlice, err)
	}

	tmpEmitter.Printlnf("{")

	kind := reflect.ValueOf(v.defaultValue).Kind()

	if kind == reflect.Slice {
		df, ok := v.defaultValue.([]any)
		if !ok {
			return "", ErrInvalidDefaultValue
		}

		for _, value := range df {
			tmpEmitter.Printlnf("%s,", litter.Sdump(value))
		}
	} else {
		return "", ErrCannotFindSlideToDump
	}

	tmpEmitter.Printf("}")

	return tmpEmitter.String(), nil
}

func (v *defaultValidator) desc() *validatorDesc {
	var packages []packageImport

	_, ok := v.defaultValueType.(codegen.DurationType)
	if v.defaultValueType != nil && ok {
		defaultDurationISO8601, ok := v.defaultValue.(string)
		if ok && defaultDurationISO8601 != "" {
			packages = []packageImport{
				{qualifiedName: "fmt"},
				{qualifiedName: "time"},
			}
		}
	}

	return &validatorDesc{
		hasError:            false,
		beforeJSONUnmarshal: false,
		requiresRawAfter:    true,
		imports:             packages,
	}
}

type arrayValidator struct {
	jsonName   string
	fieldName  string
	arrayDepth int
	minItems   int
	maxItems   int
}

func (v *arrayValidator) generate(out *codegen.Emitter, format string) error {
	if v.minItems == 0 && v.maxItems == 0 {
		return nil
	}

	value := getPlainName(v.fieldName)
	fieldName := v.jsonName

	var indexes []string

	for i := 1; i < v.arrayDepth; i++ {
		index := fmt.Sprintf("i%d", i)
		indexes = append(indexes, index)
		out.Printlnf(`for %s := range %s {`, index, value)
		value += fmt.Sprintf("[%s]", index)
		fieldName += "[%d]"

		out.Indent(1)
	}

	fieldName = fmt.Sprintf(`"%s"`, fieldName)
	if len(indexes) > 0 {
		fieldName = fmt.Sprintf(`fmt.Sprintf(%s, %s)`, fieldName, strings.Join(indexes, ", "))
	}

	if v.minItems != 0 {
		out.Printlnf(`if %s != nil && len(%s) < %d {`, value, value, v.minItems)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s length: must be >= %%d", %s, %d)`, fieldName, v.minItems)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.maxItems != 0 {
		out.Printlnf(`if len(%s) > %d {`, value, v.maxItems)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s length: must be <= %%d", %s, %d)`, fieldName, v.maxItems)
		out.Indent(-1)
		out.Printlnf("}")
	}

	for i := 1; i < v.arrayDepth; i++ {
		out.Indent(-1)
		out.Printlnf("}")
	}

	return nil
}

func (v *arrayValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}

type stringValidator struct {
	jsonName   string
	fieldName  string
	minLength  int
	maxLength  int
	isNillable bool
	pattern    string
	constVal   *string
}

func (v *stringValidator) generate(out *codegen.Emitter, format string) error {
	value := getPlainName(v.fieldName)
	fieldName := v.jsonName
	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
	}

	if len(v.pattern) != 0 {
		if v.isNillable {
			out.Printlnf("if %s != nil {", value)
			out.Indent(1)
		}

		out.Printlnf(
			`if matched, _ := regexp.MatchString(`+"`%s`"+`, string(%s%s)); !matched {`,
			v.pattern, pointerPrefix, value,
		)
		out.Indent(1)
		out.Printlnf(
			`return fmt.Errorf("field %%s pattern match: must match %%s", "%s", `+"`%s`"+`)`,
			v.fieldName, v.pattern,
		)
		out.Indent(-1)
		out.Printlnf("}")

		if v.isNillable {
			out.Indent(-1)
			out.Printlnf("}")
		}
	}

	if v.constVal != nil {
		out.Printlnf(`if %s%s%s != "%s" {`, checkPointer, pointerPrefix, value, *v.constVal)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be equal to %%s", "%s", "%s")`, fieldName, *v.constVal)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.minLength == 0 && v.maxLength == 0 {
		return nil
	}

	if v.minLength != 0 {
		out.Printlnf(`if %sutf8.RuneCountInString(string(%s%s)) < %d {`, checkPointer, pointerPrefix, value, v.minLength)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s length: must be >= %%d", "%s", %d)`, fieldName, v.minLength)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.maxLength != 0 {
		out.Printlnf(`if %sutf8.RuneCountInString(string(%s%s)) > %d {`, checkPointer, pointerPrefix, value, v.maxLength)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s length: must be <= %%d", "%s", %d)`, fieldName, v.maxLength)
		out.Indent(-1)
		out.Printlnf("}")
	}

	return nil
}

func (v *stringValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}

type numericValidator struct {
	jsonName         string
	fieldName        string
	isNillable       bool
	multipleOf       *float64
	maximum          *float64
	exclusiveMaximum *any
	minimum          *float64
	exclusiveMinimum *any
	constVal         any
	roundToInt       bool
}

func (v *numericValidator) generate(out *codegen.Emitter, format string) error {
	value := getPlainName(v.fieldName)
	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
	}

	if v.constVal != nil {
		out.Printlnf(`if %s%s%s != %v {`, checkPointer, pointerPrefix, value, v.constVal)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be equal to %%v", "%s", %v)`, v.jsonName, v.constVal)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.multipleOf != nil {
		if v.roundToInt {
			out.Printlnf(`if %s %s%s %% %v != 0 {`, checkPointer, pointerPrefix, value, v.valueOf(*v.multipleOf))
			out.Indent(1)
			out.Printlnf(`return fmt.Errorf("field %%s: must be a multiple of %%v", "%s", %f)`, v.jsonName, *v.multipleOf)
			out.Indent(-1)
			out.Printlnf("}")
		} else {
			if v.isNillable {
				out.Printlnf(`if %s != nil {`, value)
			} else {
				out.Printlnf("{")
			}

			out.Indent(1)
			out.Printlnf("remainder := math.Mod(%s%s, %v)", pointerPrefix, value, v.valueOf(*v.multipleOf))
			out.Printlnf(
				`if !(math.Abs(remainder) < 1e-10 || math.Abs(remainder - %v) < 1e-10) {`, v.valueOf(*v.multipleOf))
			out.Indent(1)
			out.Printlnf(`return fmt.Errorf("field %%s: must be a multiple of %%v", "%s", %f)`, v.jsonName, *v.multipleOf)
			out.Indent(-1)
			out.Printlnf("}")

			out.Indent(-1)
			out.Printlnf("}")
		}
	}

	nMin, nMax, nMinExclusive, nMaxExclusive := mathutils.NormalizeBounds(
		v.minimum, v.maximum, v.exclusiveMinimum, v.exclusiveMaximum,
	)

	v.genBoundary(out, checkPointer, pointerPrefix, value, nMax, nMaxExclusive, "<")
	v.genBoundary(out, checkPointer, pointerPrefix, value, nMin, nMinExclusive, ">")

	return nil
}

func (v *numericValidator) genBoundary(
	out *codegen.Emitter,
	checkPointer,
	pointerPrefix,
	value string,
	boundary *float64,
	exclusive bool,
	sign string,
) {
	if boundary == nil {
		return
	}

	// Technically, this should be based on schema version, but that information is lost.
	comp := sign
	if exclusive {
		// We're putting the other number first, so we need the = if it's exclusive.
		comp += "="
	} else {
		sign += "="
	}

	out.Printlnf(`if %s%v %s%s %s {`, checkPointer, v.valueOf(*boundary), comp, pointerPrefix, value)
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("field %%s: must be %s %%v", "%s", %v)`, sign, v.jsonName, v.valueOf(*boundary))
	out.Indent(-1)
	out.Printlnf("}")
}

func (v *numericValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}

func (v *numericValidator) valueOf(val float64) any {
	if v.roundToInt {
		return int64(val)
	}

	return val
}

type booleanValidator struct {
	jsonName   string
	fieldName  string
	isNillable bool
	constVal   *bool
}

func (v *booleanValidator) generate(out *codegen.Emitter, unmarshalTemplate string) error {
	value := getPlainName(v.fieldName)
	fieldName := v.jsonName
	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
	}

	if v.constVal != nil {
		out.Printlnf(`if %s%s%s != %t {`, checkPointer, pointerPrefix, value, *v.constVal)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be equal to %%t", "%s", %t)`, fieldName, *v.constVal)
		out.Indent(-1)
		out.Printlnf("}")
	}

	return nil
}

func (v *booleanValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}

func getPlainName(fieldName string) string {
	if fieldName == "" {
		return varNamePlainStruct
	}

	return fmt.Sprintf("%s.%s", varNamePlainStruct, fieldName)
}

type anyOfValidator struct {
	fieldName string
	elemCount int
}

func (v *anyOfValidator) generate(out *codegen.Emitter, format string) error {
	for i := range v.elemCount {
		out.Printlnf(`var %s_%d %s_%d`, lowerFirst(v.fieldName), i, upperFirst(v.fieldName), i)
	}

	out.Printlnf(`var errs []error`)

	for i := range v.elemCount {
		out.Printlnf(
			`if err := %s_%d.Unmarshal%s(value); err != nil {`,
			lowerFirst(v.fieldName),
			i,
			strings.ToUpper(format),
		)
		out.Indent(1)
		out.Printlnf(`errs = append(errs, err)`)
		out.Indent(-1)
		out.Printlnf(`}`)
	}

	out.Printlnf("if len(errs) == %d {", v.elemCount)
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("all validators failed: %%s", errors.Join(errs...))`)
	out.Indent(-1)
	out.Printlnf("}")

	return nil
}

func (v *anyOfValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: true,
	}
}

func lowerFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func upperFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// JSON Schema format keywords supported by the built-in formatValidator.
const (
	formatKeywordUUID         = "uuid"
	formatKeywordEmail        = "email"
	formatKeywordURI          = "uri"
	formatKeywordURIReference = "uri-reference"
	formatKeywordHostname     = "hostname"
	formatKeywordRegex        = "regex"
)

// Embedded validation patterns expressed in RE2 (Go's regexp syntax).
const (
	formatRegexUUID     = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	formatRegexHostname = `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	// formatRegexURIRef enforces the RFC 3986 character set for URI references:
	// unreserved + reserved characters and properly formed percent-encoded
	// triplets. Used in addition to net/url.Parse, which accepts almost any
	// string and so cannot reject e.g. whitespace or malformed pct-encoding.
	formatRegexURIRef = `^([A-Za-z0-9\-._~:/?#\[\]@!$&'()*+,;=]|%[0-9A-Fa-f]{2})*$`
)

// Names of the precompiled regex vars emitted at package scope when a
// generated file uses a regex-backed format validator. Reused across all
// validation call sites so we pay regexp.MustCompile once per pattern per
// generated package, not per call.
const (
	regexpVarNameUUID     = "regexpFormatUUID"
	regexpVarNameHostname = "regexpFormatHostname"
	regexpVarNameURIRef   = "regexpFormatURIRef"
)

// formatRegexpDecl returns the precompiled-regex package-level decl, if any,
// that the named format keyword needs at runtime. Formats with no regex
// dependency (e.g. "email", "regex") return nil.
func formatRegexpDecl(format string) *codegen.RegexpVar {
	switch format {
	case formatKeywordUUID:
		return &codegen.RegexpVar{Name: regexpVarNameUUID, Pattern: formatRegexUUID}
	case formatKeywordHostname:
		return &codegen.RegexpVar{Name: regexpVarNameHostname, Pattern: formatRegexHostname}
	case formatKeywordURI, formatKeywordURIReference:
		return &codegen.RegexpVar{Name: regexpVarNameURIRef, Pattern: formatRegexURIRef}
	}

	return nil
}

// isKnownFormatKeyword reports whether the named format has a built-in
// runtime validator implemented by formatValidator.
func isKnownFormatKeyword(format string) bool {
	switch format {
	case formatKeywordUUID,
		formatKeywordEmail,
		formatKeywordURI,
		formatKeywordURIReference,
		formatKeywordHostname,
		formatKeywordRegex:
		return true
	}

	return false
}

// formatValidatorImports returns the package imports the generated validator
// for the named format requires.
// Stdlib import paths used by the generated format validators. Centralised so
// the per-format dispatch below names them consistently.
const (
	importPathRegexp  = "regexp"
	importPathStrings = "strings"
	importPathNetMail = "net/mail"
	importPathNetURL  = "net/url"
)

func formatValidatorImports(format string) []packageImport {
	switch format {
	case formatKeywordUUID, formatKeywordRegex:
		return []packageImport{{qualifiedName: importPathRegexp}}
	case formatKeywordHostname:
		// strings.TrimSuffix lets the length cap exclude the optional trailing
		// root dot per RFC 1034 §3.1, matching the doc comment on the emit.
		return []packageImport{
			{qualifiedName: importPathRegexp},
			{qualifiedName: importPathStrings},
		}
	case formatKeywordEmail:
		return []packageImport{{qualifiedName: importPathNetMail}}
	case formatKeywordURI, formatKeywordURIReference:
		return []packageImport{
			{qualifiedName: importPathNetURL},
			{qualifiedName: importPathRegexp},
		}
	}

	return nil
}

// formatValidator emits runtime validation for JSON Schema `format` keywords
// on string-typed fields. It mirrors stringValidator's nillable-pointer
// handling and runs after the typed struct has been decoded.
type formatValidator struct {
	jsonName   string
	fieldName  string
	format     string
	isNillable bool
}

func (v *formatValidator) generate(out *codegen.Emitter, _ string) error {
	value := getPlainName(v.fieldName)

	pointerPrefix := ""
	if v.isNillable {
		pointerPrefix = "*"
	}

	target := pointerPrefix + value

	if v.isNillable {
		out.Printlnf("if %s != nil {", value)
		out.Indent(1)
	}

	switch v.format {
	case formatKeywordUUID:
		out.Printlnf("if !%s.MatchString(string(%s)) {", regexpVarNameUUID, target)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid uuid", "%s")`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")

	case formatKeywordHostname:
		// RFC 1123: each label is enforced by the regex (1 + {0,61} + 1 chars);
		// the overall hostname length cap (253 octets, exclusive of any trailing
		// root dot per RFC 1034 §3.1) is checked separately because regex
		// backtracking on the per-label structure would not bound the total
		// length on its own. TrimSuffix excludes the optional trailing dot
		// from the cap so a 253-char hostname plus root dot is still accepted.
		out.Printlnf(
			`if !%s.MatchString(string(%s)) || len(strings.TrimSuffix(string(%s), ".")) > 253 {`,
			regexpVarNameHostname, target, target,
		)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid hostname", "%s")`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")

	case formatKeywordEmail:
		// RFC 5321 addr-spec only: reject display-name forms ("Alice <a@b>")
		// and require the parsed address to round-trip exactly. net/mail
		// accepts the broader RFC 5322 syntax by default.
		out.Printlnf(
			`if addr, err := mail.ParseAddress(string(%s));`+
				` err != nil || addr.Name != "" || addr.Address != string(%s) {`,
			target, target,
		)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid email (RFC 5321 addr-spec)", "%s")`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")

	case formatKeywordURI:
		out.Printlnf("if u, err := url.Parse(string(%s)); err != nil || !u.IsAbs() {", target)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid absolute uri", "%s")`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if !%s.MatchString(string(%s)) {", regexpVarNameURIRef, target)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid absolute uri", "%s")`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")

	case formatKeywordURIReference:
		out.Printlnf("if _, err := url.Parse(string(%s)); err != nil {", target)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid uri reference: %%w", "%s", err)`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")
		out.Printlnf("if !%s.MatchString(string(%s)) {", regexpVarNameURIRef, target)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid uri reference", "%s")`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")

	case formatKeywordRegex:
		// JSON Schema specifies the `regex` format as ECMA-262 (JavaScript)
		// regular expressions, but Go's stdlib only supports RE2 syntax.
		// Patterns valid in ECMA-262 but not RE2 (backreferences, lookaround)
		// will be rejected here. The error message reflects what we actually
		// validate against.
		out.Printlnf("if _, err := regexp.Compile(string(%s)); err != nil {", target)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s: must be a valid RE2 regular expression: %%w", "%s", err)`, v.jsonName)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.isNillable {
		out.Indent(-1)
		out.Printlnf("}")
	}

	return nil
}

func (v *formatValidator) desc() *validatorDesc {
	d := &validatorDesc{
		hasError: true,
		imports:  formatValidatorImports(v.format),
	}

	if regexpDecl := formatRegexpDecl(v.format); regexpDecl != nil {
		d.decls = append(d.decls, regexpDecl)
	}

	return d
}
