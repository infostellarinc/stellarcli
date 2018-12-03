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

package util

import (
	"fmt"
	"strings"
	"time"
)

var (
	// Acceptable time format used to parse datetime.
	acceptableTimeFormats = []string{
		"20060102",
		"20060102150405",
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006/01/02",
		"2006/01/02 15:04:05",
		"2006/01/02T15:04:05",
	}
)

// Parse the given time and return time.Time.
func ParseDateTime(dateStr string) (*time.Time, error) {
	for _, f := range acceptableTimeFormats {
		parsedTime, err := time.Parse(f, dateStr)
		if err == nil {
			return &parsedTime, nil
		}
	}
	return &time.Time{},
		fmt.Errorf("failed to parse the date %v. Date format must to be one of: \n%s", dateStr,
			strings.Join(acceptableTimeFormats, ", "))
}
