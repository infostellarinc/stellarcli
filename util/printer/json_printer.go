package printer

import (
	"bufio"
	"io"
	"log"
)

type JSONPrinter struct {
	// Options of JSON format.
	Options JSONPrinterOptions
	// A writer to buffer values in a line.
	writer *bufio.Writer
}

type JSONPrinterOptions struct {
	Out io.Writer

	Indent string
}

const (
	// Default indentation of the output json file.
	indent = "  "
)

func NewJSONPrinter(o JSONPrinterOptions) *JSONPrinter {
	writer := bufio.NewWriter(o.Out)

	printer := &JSONPrinter{Options: o, writer: writer}
	return printer
}

func (p *JSONPrinter) Flush() {
	err := p.writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

func (p *JSONPrinter) Write(r []interface{}) {
	log.Fatal("JSON printer has not been implemented yet.")
}

func NewJSONPrinterOptions(output io.Writer) JSONPrinterOptions {
	return JSONPrinterOptions{
		Out: output,

		Indent: indent,
	}
}
