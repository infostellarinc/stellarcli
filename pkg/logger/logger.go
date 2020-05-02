// Copyright © 2020 Infostellar, Inc.
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

package logger

import (
	"fmt"
	"log"
	"time"
)

var lastLoggedTimestamp time.Time
var emitRateMillis = 500
var isVerbose = false
var isDebug = false
var isNewLine = true

// SetVerbose enables or disables verbose logs
func SetVerbose(v bool) {
	isVerbose = v
}

// SetDebug enables or disables debugging logs
func SetDebug(d bool) {
	isDebug = d
}

// writes to log un-throttled
func info(format string, v ...interface{}) {
	lineCheck()
	log.Printf(format, v...)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	lineCheck()
	log.Println(v...)
}

// Printf calls Output to print to the standard
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	lineCheck()
	log.Printf(format, v...)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	lineCheck()
	log.Fatalf(format, v...)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	lineCheck()
	log.Fatal(v...)
}

// writes to log but is throttled by emitRateMillis; throttled messages are dropped
func InfoThrottled(format string, v ...interface{}) {
	if throttleCheck() {
		log.Printf(format, v...)
	}
}

func Verbose(format string, v ...interface{}) {
	if isVerbose {
		lineCheck()
		log.Printf(format, v...)
	}
}

func VerboseThrottled(format string, v ...interface{}) {
	if isVerbose && throttleCheck() {
		log.Printf(format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if isDebug {
		lineCheck()
		log.Printf(format, v...)
	}
}

func DebugThrottled(format string, v ...interface{}) {
	if isDebug && throttleCheck() {
		log.Printf(format, v...)
	}
}

func throttleCheck() bool {
	if lastLoggedTimestamp.Add(time.Millisecond * time.Duration(emitRateMillis)).Before(time.Now()) {
		lastLoggedTimestamp = time.Now()
		return true
	}
	return false
}

// helper function for LastLine(); this function will track if cursor is on the "LastLine" and print out a newline to
// prevent overwriting text written by LastLine()
func lineCheck() {
	if !isNewLine {
		fmt.Println()
		isNewLine = true
	}
}

// LastLine overwrites last line of stdout without creating a new line
func LastLine(format string, v ...interface{}) {
	fmt.Printf("\r"+format+" ", v...)
	isNewLine = false
}

// LastLine overwrites last line of stdout without creating a new line, throttled
func LastLineThrottled(format string, v ...interface{}) {
	if throttleCheck() {
		LastLine(format, v...)
		isNewLine = false
	}
}
