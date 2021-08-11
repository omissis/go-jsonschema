package codegen

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
)

type Emitter struct {
	sb            strings.Builder
	maxLineLength uint
	start         bool
	indent        uint
}

func NewEmitter(maxLineLength uint) *Emitter {
	return &Emitter{
		maxLineLength: maxLineLength,
		start:         true,
	}
}

func (e *Emitter) String() string {
	return e.sb.String()
}

func (e *Emitter) Bytes() []byte {
	return []byte(e.sb.String())
}

func (e *Emitter) Indent(n int) {
	if int(e.indent)+n < 0 {
		panic("unexpected unbalanced indentation")
	}
	e.indent += uint(n)
}

func (e *Emitter) Comment(s string) {
	if s != "" {
		limit := e.maxLineLength - uint(e.indent)
		lines := strings.Split(wordwrap.WrapString(s, limit), "\n")
		for _, line := range lines {
			e.Println("// %s", line)
		}
	}
}

func (e *Emitter) Print(format string, args ...interface{}) {
	e.checkIndent()
	fmt.Fprintf(&e.sb, format, args...)
	e.start = false
}

func (e *Emitter) Println(format string, args ...interface{}) {
	e.Print(format, args...)
	e.Newline()
}

func (e *Emitter) Newline() {
	e.sb.WriteRune('\n')
	e.start = true
}

func (e *Emitter) checkIndent() {
	if e.start {
		for i := uint(0); i < e.indent; i++ {
			e.sb.WriteRune('\t')
		}
		e.start = false
	}
}

func (e *Emitter) MaxLineLength() uint {
	return e.maxLineLength
}
