package codegen

type Rule interface {
	GenerateValidation(out *Emitter, varName, errorContext string)
}

type ArrayNotEmpty struct{}

func (ArrayNotEmpty) GenerateValidation(out *Emitter, varName, errorContext string) {
	out.Println(`if len(%s) == 0 {`, varName)
	out.Indent(1)
	out.Println(`return fmt.Errorf("%s: array cannot be empty")`, errorContext)
	out.Indent(-1)
	out.Println("}")
}

type NilStructFieldRequired struct{}

func (NilStructFieldRequired) GenerateValidation(out *Emitter, varName, errorContext string) {
	out.Println(`if %s == nil {`, varName)
	out.Indent(1)
	out.Println(`return fmt.Errorf("%s: must be set")`, errorContext)
	out.Indent(-1)
	out.Println("}")
}
