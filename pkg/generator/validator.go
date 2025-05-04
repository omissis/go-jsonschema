package generator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/sanity-io/litter"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/mathutils"
)

type validator interface {
	generate(out *codegen.Emitter, format string)
	desc() *validatorDesc
}

type validatorDesc struct {
	hasError            bool
	beforeJSONUnmarshal bool
	requiresRawAfter    bool
}

var (
	_ validator = new(requiredValidator)
	_ validator = new(nullTypeValidator)
	_ validator = new(defaultValidator)
	_ validator = new(arrayValidator)
	_ validator = new(stringValidator)
	_ validator = new(numericValidator)
	_ validator = new(anyOfValidator)
)

type requiredValidator struct {
	jsonName string
	declName string
}

func (v *requiredValidator) generate(out *codegen.Emitter, format string) {
	// The container itself may be null (if the type is ["null", "object"]), in which case
	// the map will be nil and none of the properties are present. This shouldn't fail
	// the validation, though, as that's allowed as long as the container is allowed to be null.
	out.Printlnf(`if _, ok := %s["%s"]; %s != nil && !ok {`, varNameRawMap, v.jsonName, varNameRawMap)
	out.Indent(1)
	out.Printlnf(`return fmt.Errorf("field %s in %s: required")`, v.jsonName, v.declName)
	out.Indent(-1)
	out.Printlnf("}")
}

func (v *requiredValidator) desc() *validatorDesc {
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

func (v *nullTypeValidator) generate(out *codegen.Emitter, format string) {
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
	defaultValue     interface{}
}

func (v *defaultValidator) generate(out *codegen.Emitter, format string) {
	defaultValue := v.dumpDefaultValue(out)

	out.Printlnf(`if v, ok := %s["%s"]; !ok || v == nil {`, varNameRawMap, v.jsonName)
	out.Indent(1)
	out.Printlnf(`%s = %s`, getPlainName(v.fieldName), defaultValue)
	out.Indent(-1)
	out.Printlnf("}")
}

func (v *defaultValidator) dumpDefaultValue(out *codegen.Emitter) any {
	nt, ok := v.defaultValueType.(*codegen.NamedType)
	if v.defaultValueType != nil && ok {
		dvm, ok := v.defaultValue.(map[string]any)
		if ok {
			namedFields := ""
			for _, k := range sortedKeys(dvm) {
				namedFields += fmt.Sprintf("\n%s: %s,", upperFirst(k), litter.Sdump(dvm[k]))
			}

			namedFields += "\n"

			return fmt.Sprintf(`%s{%s}`, nt.Decl.GetName(), namedFields)
		}
	}

	if defaultValue, err := v.tryDumpDefaultSlice(out.MaxLineLength()); err == nil {
		return defaultValue
	}

	// Fallback to sdump in case we couldn't dump it properly.
	return litter.Sdump(v.defaultValue)
}

func (v *defaultValidator) tryDumpDefaultSlice(maxLineLen int32) (string, error) {
	tmpEmitter := codegen.NewEmitter(maxLineLen)
	v.defaultValueType.Generate(tmpEmitter)
	tmpEmitter.Printlnf("{")

	kind := reflect.ValueOf(v.defaultValue).Kind()
	switch kind {
	case reflect.Slice:
		df, ok := v.defaultValue.([]interface{})
		if !ok {
			return "", errors.New("invalid default value")
		}

		for _, value := range df {
			tmpEmitter.Printlnf("%s,", litter.Sdump(value))
		}

	default:
		return "", errors.New("didn't find a slice to dump")
	}

	tmpEmitter.Printf("}")

	return tmpEmitter.String(), nil
}

func (v *defaultValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            false,
		beforeJSONUnmarshal: false,
		requiresRawAfter:    true,
	}
}

type arrayValidator struct {
	jsonName   string
	fieldName  string
	arrayDepth int
	minItems   int
	maxItems   int
}

func (v *arrayValidator) generate(out *codegen.Emitter, format string) {
	if v.minItems == 0 && v.maxItems == 0 {
		return
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
}

func (v *stringValidator) generate(out *codegen.Emitter, format string) {
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

	if v.minLength == 0 && v.maxLength == 0 {
		return
	}

	if v.minLength != 0 {
		out.Printlnf(`if %slen(%s%s) < %d {`, checkPointer, pointerPrefix, value, v.minLength)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s length: must be >= %%d", "%s", %d)`, fieldName, v.minLength)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.maxLength != 0 {
		out.Printlnf(`if %slen(%s%s) > %d {`, checkPointer, pointerPrefix, value, v.maxLength)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s length: must be <= %%d", "%s", %d)`, fieldName, v.maxLength)
		out.Indent(-1)
		out.Printlnf("}")
	}
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
	roundToInt       bool
}

func (v *numericValidator) generate(out *codegen.Emitter, format string) {
	value := getPlainName(v.fieldName)
	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
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

func (v *anyOfValidator) generate(out *codegen.Emitter, format string) {
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
