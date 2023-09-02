package generator

import (
	"strings"

	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type formatter interface {
	generate(out *codegen.Emitter)
}

const (
	formatJSON = "json"
	formatYAML = "yaml"
)

type jsonFormatter struct {
	declaredType string
	validators   []validator
}

func (jf *jsonFormatter) generate(out *codegen.Emitter) {
	out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatJSON), formatJSON)
	out.Printlnf("func (j *%s) Unmarshal%s(b []byte) error {", jf.declaredType, strings.ToUpper(formatJSON))
	out.Indent(1)
	out.Printlnf("var %s map[string]interface{}", varNameRawMap)
	out.Printlnf("if err := %s.Unmarshal(b, &%s); err != nil { return err }",
		formatJSON, varNameRawMap)

	for _, v := range jf.validators {
		if v.desc().beforeJSONUnmarshal {
			v.generate(out)
		}
	}

	out.Printlnf("type Plain %s", jf.declaredType)
	out.Printlnf("var %s Plain", varNamePlainStruct)
	out.Printlnf("if err := %s.Unmarshal(b, &%s); err != nil { return err }",
		formatJSON, varNamePlainStruct)

	for _, v := range jf.validators {
		if !v.desc().beforeJSONUnmarshal {
			v.generate(out)
		}
	}

	out.Printlnf("*j = %s(%s)", jf.declaredType, varNamePlainStruct)
	out.Printlnf("return nil")
	out.Indent(-1)
	out.Printlnf("}")
}

type yamlFormatter struct {
	declaredType string
	validators   []validator
}

func (yf *yamlFormatter) generate(out *codegen.Emitter) {
	out.Commentf("Unmarshal%s implements %s.Unmarshaler.", strings.ToUpper(formatYAML), formatYAML)
	out.Printlnf("func (j *%s) Unmarshal%s(unmarshal func(interface{}) error) error {", yf.declaredType,
		strings.ToUpper(formatYAML))
	out.Indent(1)
	out.Printlnf("var %s map[string]interface{}", varNameRawMap)
	out.Printlnf("if err := unmarshal(&%s); err != nil { return err }", varNameRawMap)

	for _, v := range yf.validators {
		if v.desc().beforeJSONUnmarshal {
			v.generate(out)
		}
	}

	out.Printlnf("type Plain %s", yf.declaredType)
	out.Printlnf("var %s Plain", varNamePlainStruct)
	out.Printlnf("if err := unmarshal(&%s); err != nil { return err }", varNamePlainStruct)

	for _, v := range yf.validators {
		if !v.desc().beforeJSONUnmarshal {
			v.generate(out)
		}
	}

	out.Printlnf("*j = %s(%s)", yf.declaredType, varNamePlainStruct)
	out.Printlnf("return nil")
	out.Indent(-1)
	out.Printlnf("}")
}
