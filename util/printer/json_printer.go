package printer

import (
	"bufio"
	"bytes"
	"encoding/json"
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

// Write a header
func (p *JSONPrinter) WriteHeader(t []TemplateItem) {
	// Do nothing
}

func (p *JSONPrinter) Write(r []interface{}) {
	log.Fatal("JSON printer has not been implemented yet.")
}

// Write fields with the template.
// Prefer use of json encode then indent over marshalIndent to prevent HTML escaping,
// which causes invalid URLs to be printed.
func (p *JSONPrinter) WriteWithTemplate(r []map[string]interface{}, t []TemplateItem) {
	encBuffer := &bytes.Buffer{}
	encoder := json.NewEncoder(encBuffer)
	encoder.SetEscapeHTML(false)
	encErr := encoder.Encode(r)
	if encErr != nil {
		log.Fatal(encErr)
	}

	indBuffer := &bytes.Buffer{}
	indErr := json.Indent(indBuffer, encBuffer.Bytes(), "", p.Options.Indent)
	if indErr != nil {
		log.Fatal(indErr)
	}

	p.writer.Write(indBuffer.Bytes())
	p.writer.WriteString("\n")
}

func NewJSONPrinterOptions(output io.Writer) JSONPrinterOptions {
	return JSONPrinterOptions{
		Out: output,

		Indent: indent,
	}
}
