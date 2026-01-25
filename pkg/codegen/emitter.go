package codegen

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"
)

type Emitter struct {
	sb            strings.Builder
	maxLineLength int32
	start         bool
	indent        int32
}

func NewEmitter(maxLineLength int32) *Emitter {
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

func (e *Emitter) Indent(n int32) {
	if e.indent+n < 0 {
		panic("unexpected unbalanced indentation")
	}

	e.indent += n
}

func (e *Emitter) Comment(s string) {
	if s != "" {
		limit := max(e.maxLineLength-e.indent, 0)

		//nolint:gosec // limit is guarded against negative values
		lines := strings.SplitSeq(wordwrap.WrapString(s, uint(limit)), "\n")

		for line := range lines {
			e.Printlnf("// %s", line)
		}
	}
}

func (e *Emitter) Commentf(s string, args ...any) {
	s = fmt.Sprintf(s, args...)
	if s != "" {
		limit := max(e.maxLineLength-e.indent, 0)

		//nolint:gosec // limit is guarded against negative values
		lines := strings.SplitSeq(wordwrap.WrapString(s, uint(limit)), "\n")

		for line := range lines {
			e.Printlnf("// %s", line)
		}
	}
}

func (e *Emitter) Printf(format string, args ...any) {
	e.checkIndent()
	fmt.Fprintf(&e.sb, format, args...)
	e.start = false
}

func (e *Emitter) Printlnf(format string, args ...any) {
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

func (e *Emitter) MaxLineLength() int32 {
	return e.maxLineLength
}
