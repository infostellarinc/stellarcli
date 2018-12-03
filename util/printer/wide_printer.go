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
	"bytes"
	"io"
	"log"
	"text/tabwriter"
)

type WidePrinter struct {
	Options WidePrinterOptions

	buf        *bytes.Buffer
	csvPrinter *CSVPrinter
	tabWriter  *tabwriter.Writer
}

type WidePrinterOptions struct {
	Out io.Writer

	CRLF       string
	DateFormat string
	Flags      uint
	MinWidth   int
	TabWidth   int
	Padding    int
	PadChar    byte
}

// Parameters used in tabwriter.
// Details are in Go lang official doc.
// https://golang.org/pkg/text/tabwriter/
const (
	// Formatting control.
	// See possible values in https://golang.org/pkg/text/tabwriter/
	flags = 0
	// Minimal cell width including any padding.
	minWidth = 4
	// Padding added to a cell before computing its width
	padding = 2
	// ASCII char used for padding.
	padChar = ' '
	//Width of tab characters (equivalent number of spaces)
	tabWidth = 4

	// Tab char used in the input.
	tab = "\t"
)

// Create a new WidePrinter.
func NewWidePrinter(o WidePrinterOptions) *WidePrinter {
	writer := tabwriter.NewWriter(o.Out, o.MinWidth, o.TabWidth, o.Padding, o.PadChar, o.Flags)

	// CSVPrinter is used internally to generated CSV formatted text with tabs as separator.
	// It uses ByteBuffer as the output destination.
	byteBuf := &bytes.Buffer{}
	csvOptions := NewCSVPrinterOptions(byteBuf)
	csvOptions.DateFormat = o.DateFormat
	csvOptions.Separator = tab
	csvOptions.QuoteString = false

	csvPrinter := NewCSVPrinter(csvOptions)

	printer := &WidePrinter{
		csvPrinter: csvPrinter,
		tabWriter:  writer,
		buf:        byteBuf,
	}
	return printer
}

// Flush data in the buffer.
// This function must be called to output all data.
func (p *WidePrinter) Flush() {
	err := p.tabWriter.Flush()
	if err != nil {
		log.Fatal(err)
	}
}

// Format and write fields represented as an array.
func (p *WidePrinter) Write(r []interface{}) {
	// Convert given values to tab separated string.
	p.csvPrinter.Write(r)
	p.csvPrinter.Flush()

	_, err := p.tabWriter.Write(p.buf.Bytes())
	p.buf.Reset()
	if err != nil {
		log.Fatal(err)
	}
}

// Create a WidePrinterOptions with default values set.
func NewWidePrinterOptions(output io.Writer) WidePrinterOptions {
	return WidePrinterOptions{
		Out: output,

		CRLF:       crlf,
		DateFormat: dateFormat,
		Flags:      flags,
		MinWidth:   minWidth,
		TabWidth:   tabWidth,
		Padding:    padding,
		PadChar:    padChar,
	}
}
