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

package stream

import (
	"fmt"
	"testing"
	"time"
)

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("'%v' != '%v'", a, b)
	}
	t.Fatal(message)
}

func TestMetricLogging(t *testing.T) {
	metrics := *NewMetricsCollector(t.Logf)
	metrics.setPlanId("plan1")
	for i := 1; i <= 10e17; i *= 4 {
		metrics.collectMessage(i)
		metrics.logStats()
		time.Sleep(time.Millisecond * time.Duration(50))
	}
}

func TestReset(t *testing.T) {
	metrics := *NewMetricsCollector(t.Logf)
	metrics.setPlanId("plan1")
	metrics.collectMessage(100)
	metrics.logStats()
	metrics.setPlanId("plan2")
	metrics.collectMessage(50)
	metrics.logStats()
	assertEqual(t, metrics.totalMessagesReceived, int64(1), "")
	assertEqual(t, metrics.totalBytesReceived, int64(50), "")
}

func TestDelay(t *testing.T) {
	for i := int64(6); i < 1e15; i *= 10 {
		t.Log(humanReadableNanoSeconds(i))
	}
	assertEqual(t, humanReadableNanoSeconds(1), "    1 ns", "")
	assertEqual(t, humanReadableNanoSeconds(100), "  100 ns", "")
	assertEqual(t, humanReadableNanoSeconds(1e3), "  1.0 µs", "")
	assertEqual(t, humanReadableNanoSeconds(1e6), "  1.0 ms", "")
	assertEqual(t, humanReadableNanoSeconds(6e9), "  6.0 s ", "")
	assertEqual(t, humanReadableNanoSeconds(6e11), " 10.0 m ", "")
	assertEqual(t, humanReadableNanoSeconds(6e13), " 16.7 h ", "")
	assertEqual(t, humanReadableNanoSeconds(6e15), "1666.7 h ", "")
}
