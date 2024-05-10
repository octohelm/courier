package jsonschema

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func PrintWithDoc() SchemaPrintOption {
	return func(o *printOption) {
		o.WithDoc = true
	}
}

type SchemaPrintOption func(o *printOption)

type printOption struct {
	WithDoc bool
}

func (o *printOption) Build(optionFns ...SchemaPrintOption) {
	for _, fn := range optionFns {
		fn(o)
	}
}

type Printer interface {
	io.Writer

	PrintFrom(pt PrinterTo, optionFns ...SchemaPrintOption)
	Print(values ...any)
	Printf(fmt string, values ...any)

	Indent() func()
	Return()
	PrintDoc(desc string)
}

type PrinterTo interface {
	PrintTo(w io.Writer, optionFns ...SchemaPrintOption)
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

func (p *printer) PrintDoc(desc string) {
	if len(desc) > 0 {
		scanner := bufio.NewScanner(bytes.NewBufferString(desc))
		for scanner.Scan() {
			p.Print("// ", scanner.Text())
			p.Return()
		}
	}

}

func (p *printer) printIndent() {
	if p.newline {
		for i := 0; i < p.indent; i++ {
			_, _ = fmt.Fprintf(p, "  ")
		}
		p.newline = false
	}
}

func (p *printer) PrintFrom(pt PrinterTo, optionFns ...SchemaPrintOption) {
	pt.PrintTo(p, optionFns...)
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
