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

package stream

import (
	"fmt"
	"log"
	"time"
)

// logger optios to enable rate-limited logging
// use infoThrottled(...) for throttled logging, or simply info(...) for un-throttled logging
// emitRateMillis = throttle rate in milliseconds
// throttled logs are dropped
type ThrottledLogger struct {
	lastLoggedTimestamp time.Time
	emitRateMillis      int
	isVerbose           bool
	isDebug             bool
}

func NewThrottledLogger(emitRateMillis int, isVerbose, isDebug bool) *ThrottledLogger {
	return &ThrottledLogger{
		lastLoggedTimestamp: time.Now(),
		emitRateMillis:      emitRateMillis,
		isVerbose:           isVerbose,
		isDebug:             isDebug,
	}
}

// writes to log un-throttled
func (logger *ThrottledLogger) info(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// writes to log but is throttled by emitRateMillis; throttled messages are dropped
func (logger *ThrottledLogger) infoThrottled(format string, v ...interface{}) {
	if throttleCheck() {
		log.Printf(format, v...)
	}
}

func (logger *ThrottledLogger) verbose(format string, v ...interface{}) {
	if logger.isVerbose {
		log.Printf(format, v...)
	}
}

func (logger *ThrottledLogger) verboseThrottled(format string, v ...interface{}) {
	if logger.isVerbose && throttleCheck() {
		log.Printf(format, v...)
	}
}

func (logger *ThrottledLogger) debug(format string, v ...interface{}) {
	if logger.isDebug {
		log.Printf(format, v...)
	}
}

func (logger *ThrottledLogger) debugThrottled(format string, v ...interface{}) {
	if logger.isDebug && throttleCheck() {
		log.Printf(format, v...)
	}
}

func throttleCheck() bool {
	if logger.lastLoggedTimestamp.Add(time.Millisecond * time.Duration(logger.emitRateMillis)).Before(time.Now()) {
		logger.lastLoggedTimestamp = time.Now()
		return true
	}
	return false
}

// overwrites last line of stdout without creating a new line
func (logger *ThrottledLogger) lastLine(format string, v ...interface{}) {
	fmt.Printf("\r"+format+" ", v...)
}

// overwrites last line of stdout without creating a new line, throttled
func (logger *ThrottledLogger) lastLineThrottled(format string, v ...interface{}) {
	if throttleCheck() {
		logger.lastLine(format, v...)
	}
}
