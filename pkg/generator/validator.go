package generator

import (
	"fmt"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type validator interface {
	generate(out *codegen.Emitter)
	desc() *validatorDesc
}

type validatorDesc struct {
	hasError            bool
	beforeJSONUnmarshal bool
	requiredImports     []string
}

var (
	_ validator = new(requiredValidator)
	_ validator = new(nullTypeValidator)
	_ validator = new(defaultValidator)
	_ validator = new(numericValidator)
)

type requiredValidator struct {
	jsonName string
}

func (v *requiredValidator) generate(out *codegen.Emitter) {
	out.Println(`if v, ok := %s["%s"]; !ok || v == nil {`, varNameRawMap, v.jsonName)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: required")`, v.jsonName)
	out.Indent(-1)
	out.Println("}")
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
	var indexes []string
	for i := 0; i < v.arrayDepth; i++ {
		index := fmt.Sprintf("i%d", i)
		indexes = append(indexes, index)
		out.Println(`for %s := range %s {`, index, value)
		value += fmt.Sprintf("[%s]", index)
		fieldName += "[%d]"
		out.Indent(1)
	}

	fieldName = fmt.Sprintf(`"%s"`, fieldName)
	if len(indexes) > 0 {
		fieldName = fmt.Sprintf(`fmt.Sprinf(%s, %s)`, fieldName, strings.Join(indexes, ", "))
	}

	out.Println(`if %s != nil {`, value)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %%s: must be null", %s)`, fieldName)
	out.Indent(-1)
	out.Println("}")

	for i := 0; i < v.arrayDepth; i++ {
		out.Indent(-1)
		out.Println("}")
	}
}

func (v *nullTypeValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}

type defaultValidator struct {
	jsonName     string
	fieldName    string
	defaultValue string
}

func (v *defaultValidator) generate(out *codegen.Emitter) {
	out.Println(`if v, ok := %s["%s"]; !ok || v == nil {`, varNameRawMap, v.jsonName)
	out.Indent(1)
	out.Println(`%s.%s = %s`, varNamePlainStruct, v.fieldName, v.defaultValue)
	out.Indent(-1)
	out.Println("}")
}

func (v *defaultValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            false,
		beforeJSONUnmarshal: false,
	}
}

type numericValidator struct {
	jsonName     string
	fieldName    string
	multipleOf   *float64
	min          *float64
	exclusiveMin *float64
	max          *float64
	exclusiveMax *float64
}

func (v *numericValidator) generate(out *codegen.Emitter) {
	if v.multipleOf != nil {
		v.generateMultipleOf(out)
	}
	if v.exclusiveMax != nil {
		v.generateExclusiveMax(out)
	}
	if v.max != nil {
		v.generateMax(out)
	}
	if v.exclusiveMin != nil {
		v.generateExclusiveMin(out)
	}
	if v.min != nil {
		v.generateMin(out)
	}
}

func (v *numericValidator) desc() *validatorDesc {
	requiredImports := make([]string, 0)
	if v.multipleOf != nil {
		requiredImports = append(requiredImports, "math")
	}
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
		requiredImports:     requiredImports,
	}
}

func (v *numericValidator) generateMultipleOf(out *codegen.Emitter) {
	out.Println(`if math.Mod(%s.%s, %f) != 0 {`, varNamePlainStruct, v.fieldName, *v.multipleOf)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: must be multiple of %f")`, v.jsonName, *v.multipleOf)
	out.Indent(-1)
	out.Println("}")
}

func (v *numericValidator) generateExclusiveMax(out *codegen.Emitter) {
	out.Println(`if %s.%s >= %f {`, varNamePlainStruct, v.fieldName, *v.exclusiveMax)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: must be smaller than %f")`, v.jsonName, *v.exclusiveMax)
	out.Indent(-1)
	out.Println("}")
}

func (v *numericValidator) generateMax(out *codegen.Emitter) {
	out.Println(`if %s.%s > %f {`, varNamePlainStruct, v.fieldName, *v.max)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: must be smaller or equal than %f")`, v.jsonName, *v.max)
	out.Indent(-1)
	out.Println("}")
}

func (v *numericValidator) generateExclusiveMin(out *codegen.Emitter) {
	out.Println(`if %s.%s <= %f {`, varNamePlainStruct, v.fieldName, *v.exclusiveMin)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: must be bigger than %f")`, v.jsonName, *v.exclusiveMin)
	out.Indent(-1)
	out.Println("}")
}

func (v *numericValidator) generateMin(out *codegen.Emitter) {
	out.Println(`if %s.%s < %f {`, varNamePlainStruct, v.fieldName, *v.min)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: must be bigger or equal than %f")`, v.jsonName, *v.min)
	out.Indent(-1)
	out.Println("}")
}
