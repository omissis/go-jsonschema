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
	multipleOf   float64
	min          float64
	exclusiveMin float64
	max          float64
	exclusiveMax float64
}

// todo fix combinations of them
func (v *numericValidator) generate(out *codegen.Emitter) {
	if v.multipleOf != 0 {
		// wtf printing "plain.MyMultipleOf10%10.000000" NO SPACES
		out.Println(`if %s.%s %% %f != 0 {`, varNamePlainStruct, v.fieldName, v.multipleOf)
		out.Indent(1)
		out.Println(`return fmt.Errorf("field %s: must be multiple of %f")`, v.jsonName, v.multipleOf)
		out.Indent(-1)
		out.Println("}")
		return
	}

	var operand, constraint string
	var reference float64

	if v.max != 0 {
		operand, constraint, reference = ">", "smaller or equal", v.max
	}
	if v.exclusiveMax != 0 {
		operand, constraint, reference = ">=", "smaller", v.exclusiveMax
	}
	if v.min != 0 {
		operand, constraint, reference = "<", "bigger or equal", v.min
	}
	if v.exclusiveMin != 0 {
		operand, constraint, reference = "<=", "bigger", v.exclusiveMin
	}

	out.Println(`if %s.%s %s %f {`, varNamePlainStruct, v.fieldName, operand, reference)
	out.Indent(1)
	out.Println(`return fmt.Errorf("field %s: must be %s than %f")`, v.jsonName, constraint, reference)
	out.Indent(-1)
	out.Println("}")
}

func (v *numericValidator) desc() *validatorDesc {
	return &validatorDesc{
		hasError:            true,
		beforeJSONUnmarshal: false,
	}
}
