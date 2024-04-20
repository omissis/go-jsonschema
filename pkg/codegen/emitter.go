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
		limit := e.maxLineLength - e.indent
		lines := strings.Split(wordwrap.WrapString(s, limit), "\n")

		for _, line := range lines {
			e.Printlnf("// %s", line)
		}
	}
}

func (e *Emitter) Commentf(s string, args ...interface{}) {
	s = fmt.Sprintf(s, args...)
	if s != "" {
		limit := e.maxLineLength - e.indent
		lines := strings.Split(wordwrap.WrapString(s, limit), "\n")

		for _, line := range lines {
			e.Printlnf("// %s", line)
		}
	}
}

func (e *Emitter) Printf(format string, args ...interface{}) {
	e.checkIndent()
	fmt.Fprintf(&e.sb, format, args...)
	e.start = false
}

func (e *Emitter) Printlnf(format string, args ...interface{}) {
	e.Printf(format, args...)
	e.Newline()
}

func (e *Emitter) Newline() {
	e.sb.WriteRune('\n')
	e.start = true
}

func (e *Emitter) checkIndent() {
	if e.start {
		for range e.indent {
			e.sb.WriteRune('\t')
		}

		e.start = false
	}
}

func (e *Emitter) MaxLineLength() uint {
	return e.maxLineLength
}
