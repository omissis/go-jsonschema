package generator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/sanity-io/litter"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type validator interface {
	generate(out *codegen.Emitter)
	desc() *validatorDesc
}

type validatorDesc struct {
	hasError            bool
	beforeJSONUnmarshal bool
}

var (
	_ validator = new(requiredValidator)
	_ validator = new(nullTypeValidator)
	_ validator = new(defaultValidator)
	_ validator = new(arrayValidator)
	_ validator = new(stringValidator)
	_ validator = new(numberValidator)
	_ validator = new(integerValidator)
)

type requiredValidator struct {
	jsonName string
	declName string
}

func (v *requiredValidator) generate(out *codegen.Emitter) {
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

func (v *nullTypeValidator) generate(out *codegen.Emitter) {
	value := fmt.Sprintf("%s.%s", varNamePlainStruct, v.fieldName)
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
	}
}

type defaultValidator struct {
	jsonName         string
	fieldName        string
	defaultValueType codegen.Type
	defaultValue     interface{}
}

func (v *defaultValidator) generate(out *codegen.Emitter) {
	defaultValue, err := v.tryDumpDefaultSlice(out.MaxLineLength())
	if err != nil {
		// Fallback to sdump in case we couldn't dump it properly.
		defaultValue = litter.Sdump(v.defaultValue)
	}

	out.Printlnf(`if v, ok := %s["%s"]; !ok || v == nil {`, varNameRawMap, v.jsonName)
	out.Indent(1)
	out.Printlnf(`%s.%s = %s`, varNamePlainStruct, v.fieldName, defaultValue)
	out.Indent(-1)
	out.Printlnf("}")
}

func (v *defaultValidator) tryDumpDefaultSlice(maxLineLen uint) (string, error) {
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
	}
}

type arrayValidator struct {
	jsonName   string
	fieldName  string
	arrayDepth int
	minItems   int
	maxItems   int
}

func (v *arrayValidator) generate(out *codegen.Emitter) {
	if v.minItems == 0 && v.maxItems == 0 {
		return
	}

	value := fmt.Sprintf("%s.%s", varNamePlainStruct, v.fieldName)
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
}

func (v *stringValidator) generate(out *codegen.Emitter) {
	if v.minLength == 0 && v.maxLength == 0 {
		return
	}

	value := fmt.Sprintf("%s.%s", varNamePlainStruct, v.fieldName)
	fieldName := v.jsonName

	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
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

type numberValidator struct {
	jsonName         string
	fieldName        string
	minimum          *float64
	maximum          *float64
	exclusiveMinimum *float64
	exclusiveMaximum *float64
	isNillable       bool
}

func (v *numberValidator) generate(out *codegen.Emitter) {
	if v.minimum == nil && v.maximum == nil && v.exclusiveMinimum == nil && v.exclusiveMaximum == nil {
		return
	}

	value := fmt.Sprintf("%s.%s", varNamePlainStruct, v.fieldName)
	fieldName := v.jsonName

	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
	}

	if v.minimum != nil {
		out.Printlnf(`if %s%s%s >= %f {`, checkPointer, pointerPrefix, value, *v.minimum)
		out.Indent(1)
		//TODO: How do we make the %d as precise as the Go float64?
		// Currently it doesn't print all numbers after the decimal.
		// Does the Json schema spec specify the decimal precision?
		out.Printlnf(`return fmt.Errorf("field %%s must be >= %%f", "%s", %f)`, fieldName, *v.minimum)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.maximum != nil {
		out.Printlnf(`if %s%s%s <= %f {`, checkPointer, pointerPrefix, value, *v.maximum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be <= %%f", "%s", %f)`, fieldName, *v.maximum)
		out.Indent(-1)
		out.Printlnf("}")
	}

	// The spec allows having both minimum and exclusiveMinimum.
	if v.exclusiveMinimum != nil {
		out.Printlnf(`if %s%s%s > %f {`, checkPointer, pointerPrefix, value, *v.exclusiveMinimum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be > %%f", "%s", %f)`, fieldName, *v.exclusiveMinimum)
		out.Indent(-1)
		out.Printlnf("}")
	}

	// The spec allows having both maximum and exclusiveMaximum.
	if v.exclusiveMaximum != nil {
		out.Printlnf(`if %s%s%s < %f {`, checkPointer, pointerPrefix, value, *v.exclusiveMaximum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be < %%f", "%s", %f)`, fieldName, *v.exclusiveMaximum)
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (v *numberValidator) desc() *validatorDesc {
	//TODO: What should desc() return? I'm not sure about the below implementation.
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}

type integerValidator struct {
	jsonName         string
	fieldName        string
	minimum          *int
	maximum          *int
	exclusiveMinimum *int
	exclusiveMaximum *int
	isNillable       bool
}

func (v *integerValidator) generate(out *codegen.Emitter) {
	if v.minimum == nil && v.maximum == nil && v.exclusiveMinimum == nil && v.exclusiveMaximum == nil {
		return
	}

	value := fmt.Sprintf("%s.%s", varNamePlainStruct, v.fieldName)
	fieldName := v.jsonName

	checkPointer := ""
	pointerPrefix := ""

	if v.isNillable {
		checkPointer = fmt.Sprintf("%s != nil && ", value)
		pointerPrefix = "*"
	}

	if v.minimum != nil {
		out.Printlnf(`if %s%s%s >= %d {`, checkPointer, pointerPrefix, value, *v.minimum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be >= %%d", "%s", %d)`, fieldName, *v.minimum)
		out.Indent(-1)
		out.Printlnf("}")
	}

	if v.maximum != nil {
		out.Printlnf(`if %s%s%s <= %d {`, checkPointer, pointerPrefix, value, *v.maximum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be <= %%d", "%s", %d)`, fieldName, *v.maximum)
		out.Indent(-1)
		out.Printlnf("}")
	}

	// The spec allows having both minimum and exclusiveMinimum.
	if v.exclusiveMinimum != nil {
		out.Printlnf(`if %s%s%s > %d {`, checkPointer, pointerPrefix, value, *v.exclusiveMinimum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be > %%d", "%s", %d)`, fieldName, *v.exclusiveMinimum)
		out.Indent(-1)
		out.Printlnf("}")
	}

	// The spec allows having both maximum and exclusiveMaximum.
	if v.exclusiveMaximum != nil {
		out.Printlnf(`if %s%s%s < %d {`, checkPointer, pointerPrefix, value, *v.exclusiveMaximum)
		out.Indent(1)
		out.Printlnf(`return fmt.Errorf("field %%s must be < %%d", "%s", %d)`, fieldName, *v.exclusiveMaximum)
		out.Indent(-1)
		out.Printlnf("}")
	}
}

func (v *integerValidator) desc() *validatorDesc {
	//TODO: What should desc() return? I'm not sure about the below implementation.
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}
