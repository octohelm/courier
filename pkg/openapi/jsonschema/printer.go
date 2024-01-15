package jsonschema

import (
	"fmt"
	"io"
)

type Printer interface {
	io.Writer

	PrintFrom(pt PrinterTo)
	Print(values ...any)
	Printf(fmt string, values ...any)

	Indent() func()
	Return()
}

type PrinterTo interface {
	PrintTo(w io.Writer)
}

func Print(w io.Writer, fn func(p Printer)) {
	p, ok := w.(Printer)
	if !ok {
		p = newPrinter(w)
	}
	fn(p)
}

func newPrinter(w io.Writer) Printer {
	return &printer{
		Writer: w,
	}
}

type printer struct {
	indent  int
	newline bool
	io.Writer
}

func (p *printer) printIndent() {
	if p.newline {
		for i := 0; i < p.indent; i++ {
			_, _ = fmt.Fprintf(p, "  ")
		}
		p.newline = false
	}
}

func (p *printer) PrintFrom(pt PrinterTo) {
	pt.PrintTo(p)
}

func (p *printer) Print(values ...any) {
	p.printIndent()

	_, _ = fmt.Fprint(p.Writer, values...)
}

func (p *printer) Printf(format string, values ...any) {
	p.printIndent()

	_, _ = fmt.Fprintf(p.Writer, format, values...)
}

func (p *printer) Return() {
	_, _ = fmt.Fprintf(p, "\n")
	p.newline = true
}

func (p *printer) Indent() func() {
	p.indent++
	return func() {
		p.indent--
	}
}
