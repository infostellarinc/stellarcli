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

package flag

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infostellarinc/stellarcli/cmd/util"
	"github.com/infostellarinc/stellarcli/util/printer"
)

var (
	// Supported output formats.
	availableFormats = []string{"csv", "wide", "json"}
	// Default output format.
	defaultOutputFormat = "wide"
	// Default output.
	defaultOutput = os.Stdout
)

type OutputFormatFlags struct {
	Format string
}

// Add flags to the command.
func (f *OutputFormatFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Format, "output", "o", defaultOutputFormat,
		"Output format. One of: "+strings.Join(availableFormats, "|"))
}

// Validate flag values.
func (f *OutputFormatFlags) Validate() error {
	if !util.Contains(availableFormats, f.Format) {
		return fmt.Errorf("invalid output format: %v. Expected one of : %v", f.Format,
			strings.Join(availableFormats, "|"))
	}

	return nil
}

// Return a Printer corresponding to the output format.
func (f *OutputFormatFlags) ToPrinter(FlagsOn ...bool) printer.Printer {
	format := util.ToLower(f.Format)

	switch format {
	case "wide":
		isVerbose := len(FlagsOn) > 0 && FlagsOn[0]
		if isVerbose {
			log.Println("wide format not supported when verbose flag is on")
			o := printer.NewJSONPrinterOptions(defaultOutput)
			return printer.NewJSONPrinter(o)
		}

		o := printer.NewWidePrinterOptions(defaultOutput)
		return printer.NewWidePrinter(o)
	case "csv":
		o := printer.NewCSVPrinterOptions(defaultOutput)
		return printer.NewCSVPrinter(o)
	case "json":
		o := printer.NewJSONPrinterOptions(defaultOutput)
		return printer.NewJSONPrinter(o)
	}

	log.Fatalf("unsupported output format: %v", format)
	return nil
}

// Create a new OutputFormatFlags with default values set.
func NewOutputFormatFlags() *OutputFormatFlags {
	return &OutputFormatFlags{
		Format: defaultOutputFormat,
	}
}
