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
	"sync"
	"time"
)

var lastLoggedTimestamp time.Time
var emitRateMillis = 2000
var isVerbose = false
var isDebug = false
var isNewLine = true
var lastThrottledLine *string
var throttleCheckSchedulerRunning = false
var throttleSchedulerLock sync.Mutex

// SetEmitRateMillis sets the throttle rate for throttle log methods
func SetEmitRateMillis(e int) {
	emitRateMillis = e
}

// SetVerbose enables or disables verbose logs
func SetVerbose(v bool) {
	isVerbose = v
}

// SetDebug enables or disables debugging logs
func SetDebug(d bool) {
	isDebug = d
}

// Info writes to log un-throttled
func Info(format string, v ...interface{}) {
	lineCheck()
	log.Printf(format, v...)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	lineCheck()
	log.Println(v...)
}

// PrintfRawLn calls Output to print to the standard stdout (without logger).
// Arguments are handled in the manner of fmt.Printf.
func PrintfRawLn(format string, v ...interface{}) {
	lineCheck()
	fmt.Printf(format+"\n", v...)
}

// Printf calls Output to print to the standard
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	lineCheck()
	log.Printf(format, v...)
}

// PrintlnThrottled writes to stdout (via fmt) but is throttled by emitRateMillis; throttled messages are dropped
func PrintlnThrottled(format string, v ...interface{}) {
	if throttleCheck() {
		lineCheck()
		fmt.Printf(format+"\n", v...)
	} else {
		deferPrint(format+"\n", v...)
	}
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

// Verbose writes to log iff verbose is set
func Verbose(format string, v ...interface{}) {
	if isVerbose {
		lineCheck()
		fmt.Printf("%s ", time.Now().Format("2006/01/02 15:04:05"))
		fmt.Printf(format, v...)
	}
}

// Debug writes to log iff debug is set
func Debug(format string, v ...interface{}) {
	if isDebug {
		lineCheck()
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

func deferPrint(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	lastThrottledLine = &s

	throttleSchedulerLock.Lock()
	if !throttleCheckSchedulerRunning {
		throttleCheckSchedulerRunning = true
		go scheduleDeferLogCheck()
	}
	throttleSchedulerLock.Unlock()
}

func scheduleDeferLogCheck() {
	uptimeTicker := time.NewTicker(time.Duration(emitRateMillis) / 5 * time.Millisecond)
	for {
		<-uptimeTicker.C
		if lastThrottledLine != nil && throttleCheck() {
			lineCheck()
			fmt.Printf("%s", *lastThrottledLine)
			lastThrottledLine = nil
			throttleSchedulerLock.Lock()
			throttleCheckSchedulerRunning = false
			throttleSchedulerLock.Unlock()
			return
		}
	}
}

// LastLineThrottled overwrites last line of stdout without creating a new line, throttled
func LastLineThrottled(format string, v ...interface{}) {
	if throttleCheck() {
		LastLine(format, v...)
		isNewLine = false
	} else {
		deferPrint("\r"+format+" ", v...)
	}
}
