// Copyright © 2018 Infostellar, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package printer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"
)

// A Printer which output values as a CSV format.
type CSVPrinter struct {
	// Options of CSV format.
	Options CSVPrinterOptions
	// A writer to buffer values in a line.
	writer *bufio.Writer
}

// A set of options used for CSV format.
type CSVPrinterOptions struct {
	Out io.Writer

	// A string that represents end of line.
	CRLF string
	// Date format used to convert time.Time to string.
	DateFormat string
	// CSVPrinter quote the string value if QuoteString is true.
	QuoteString bool
	// Separator between columns.
	Separator string
}

// Default option values
const (
	// Default line end character
	crlf = "\n"

	// Default dateFormat of time.Timestamp when converting it to a textual representation
	dateFormat = time.RFC3339

	// Default setting of string quotation
	quoteString = true

	// Default separator between columns
	separator = ","
)

// Create a new CSVPrinter
func NewCSVPrinter(o CSVPrinterOptions) *CSVPrinter {
	writer := bufio.NewWriter(o.Out)

	printer := &CSVPrinter{Options: o, writer: writer}
	return printer
}

// Flush data in the buffer.
// This function must be called to output all data.
func (p *CSVPrinter) Flush() {
	err := p.writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

// Format and write fields represented as an array.
func (p *CSVPrinter) Write(r []interface{}) {
	var err error

	for i, v := range r {
		if i > 0 {
			_, err = p.writer.WriteString(p.Options.Separator)
			if err != nil {
				log.Fatal(err)
			}
		}

		switch val := v.(type) {
		case string:
			if p.Options.QuoteString {
				_, err = fmt.Fprintf(p.writer, "%q", val)
			} else {
				_, err = fmt.Fprintf(p.writer, "%v", val)
			}
		case time.Time:
			_, err = p.writer.WriteString(val.Format(p.Options.DateFormat))
		default:
			_, err = fmt.Fprintf(p.writer, "%v", val)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = p.writer.WriteString(p.Options.CRLF)

	if err != nil {
		log.Fatal(err)
	}
}

// Write a header.
func (p *CSVPrinter) WriteHeader(t []TemplateItem) {
	p.Write(GetLabels(t))
}

// Write fields with the template.
func (p *CSVPrinter) WriteWithTemplate(r []map[string]interface{}, t []TemplateItem) {
	for _, obj := range r {
		p.Write(Flatten(obj, t))
	}
}

// Create a CSVPrinterOptions with default values set.
func NewCSVPrinterOptions(output io.Writer) CSVPrinterOptions {
	return CSVPrinterOptions{
		Out:         output,
		CRLF:        crlf,
		DateFormat:  dateFormat,
		QuoteString: quoteString,
		Separator:   separator,
	}
}
