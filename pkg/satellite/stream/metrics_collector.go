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
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
)

const InstantMinSamples = 5
const InstantSampleSeconds = 5

type telemetryWithTimestamp struct {
	ReceivedTime         time.Time
	DataBytes            int
	TimeLastByteReceived *timestamp.Timestamp
}

type MetricsCollector struct {
	planId                string
	timerStart            time.Time
	elapsed               float64
	totalBytesReceived    int64
	totalMessagesReceived int64
	azimuth               float64
	elevation             float64
	frequency             float64
	delayNanos            int64
	messageBuffer         []telemetryWithTimestamp
	logger                func(format string, v ...interface{})
}

// run with "go test -v" in this folder to see output
func NewMetricsCollector(logger func(format string, v ...interface{})) *MetricsCollector {
	logger("[STATS] using local time to calculate telemetry delay\n")
	return &MetricsCollector{
		logger: logger,
	}
}

// planId is used to identify when to reset the statistics; upon reset, stats is printed to output
func (metrics *MetricsCollector) setPlanId(planId string) {
	if metrics.planId != planId {
		if metrics.totalMessagesReceived > 0 {
			metrics.logStats()
			metrics.logger("\n", metrics.planId)
		}
		metrics.planId = planId
		metrics.reset()
	}
}

func (metrics *MetricsCollector) reset() {
	metrics.timerStart = time.Now()
	metrics.totalBytesReceived = 0
	metrics.totalMessagesReceived = 0
	metrics.azimuth = 0
	metrics.elevation = 0
	metrics.frequency = 0
	metrics.delayNanos = 0
	metrics.messageBuffer = make([]telemetryWithTimestamp, 0)
}

// collects metrics for telemetry data message
func (metrics *MetricsCollector) collectTelemetry(telemetry *stellarstation.Telemetry) {
	metrics.delayNanos += time.Now().UTC().UnixNano() - ((telemetry.TimeLastByteReceived.Seconds * 1e9) + int64(telemetry.TimeLastByteReceived.Nanos))
	metrics.collectMessage(len(telemetry.Data))

	// save details for instantaneous rates
	msg := telemetryWithTimestamp{
		ReceivedTime:         time.Now(),
		DataBytes:            len(telemetry.Data),
		TimeLastByteReceived: telemetry.TimeLastByteReceived,
	}
	metrics.messageBuffer = append(metrics.messageBuffer, msg)
	// always keep 5 seconds worth of dta, but no less than 5 samples
	for len(metrics.messageBuffer) > InstantMinSamples && metrics.messageBuffer[0].ReceivedTime.UnixNano() < time.Now().UnixNano()-(InstantSampleSeconds*1e9) {
		metrics.messageBuffer = metrics.messageBuffer[1:]
	}
}

// record telemetry data message received with size=messageSizeBytes
// deprecated, kept for unit-tests
func (metrics *MetricsCollector) collectMessage(messageSizeBytes int) {
	if metrics.totalBytesReceived == 0 {
		metrics.timerStart = time.Now()
	}
	metrics.totalMessagesReceived++
	metrics.elapsed = time.Since(metrics.timerStart).Seconds()
	metrics.totalBytesReceived += int64(messageSizeBytes)
}

func (metrics *MetricsCollector) collectAntenna(azimuth, elevation float64) {
	metrics.azimuth = azimuth
	metrics.elevation = elevation
}

func (metrics *MetricsCollector) collectReceiver(frequency float64) {
	metrics.frequency = frequency
}

// report collected statistics
func (metrics *MetricsCollector) logStats() {
	avgDelayNanos := humanReadableNanoSeconds(metrics.avgDelay())
	iDelayNanos := humanReadableNanoSeconds(metrics.instantDelay())
	rateStr := humanReadableCountSI(metrics.avgRate())
	iRateStr := humanReadableCountSI(metrics.instantRate())
	size := humanReadableBytes(metrics.totalBytesReceived)
	metrics.logger("[STATS] %s, plan_id: %s, azm: %5.2f, ele: %5.1f, freq: %5.1f MHz [DATA] %5d msgs, bytes: %s, rate: %sbps (avg %sbps), delay: %s (avg %s)",
		time.Now().Format("20060102 15:04:05"), metrics.planId, metrics.azimuth, metrics.elevation, metrics.frequency, metrics.totalMessagesReceived, size, iRateStr, rateStr, iDelayNanos, avgDelayNanos)
}

// return avg rate for entire plan
func (metrics *MetricsCollector) avgDelay() int64 {
	if metrics.totalMessagesReceived > 0 {
		return metrics.delayNanos / metrics.totalMessagesReceived
	}
	return 0
}

// returns the instantaneous data delay
func (metrics *MetricsCollector) instantDelay() int64 {
	if len(metrics.messageBuffer) < 2 {
		return 0
	}
	delayNanos := int64(0)
	for _, msg := range metrics.messageBuffer {
		delayNanos += msg.ReceivedTime.UTC().UnixNano() - ((msg.TimeLastByteReceived.Seconds * 1e9) + int64(msg.TimeLastByteReceived.Nanos))
	}
	return delayNanos / int64(len(metrics.messageBuffer))
}

// return avg rate for entire plan
func (metrics *MetricsCollector) avgRate() int64 {
	if metrics.totalMessagesReceived > 0 {
		return int64(float64(metrics.totalBytesReceived) / metrics.elapsed * 8.00)
	}
	return 0
}

// returns the instantaneous data rate
func (metrics *MetricsCollector) instantRate() int64 {
	if len(metrics.messageBuffer) < 3 {
		return 0
	}
	bytes := int64(0)
	for i, msg := range metrics.messageBuffer {
		if i > 0 {
			// we discard the first message size, but use its ReceivedTime as the "start time" for rate calculations
			bytes += int64(msg.DataBytes)
		}
	}
	duration := (metrics.messageBuffer[len(metrics.messageBuffer)-1].ReceivedTime.UnixNano() - metrics.messageBuffer[0].ReceivedTime.UnixNano()) / 1e9
	if duration == 0 {
		return 0
	}
	return int64(float64(bytes) / float64(duration) * 8.00)
}

// converts number (typically bytes or bits) to human readable SI string
// eg: bytes=1200 -> 1.2k
// 	   bytes=1200000 0> 1.2M
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableCountSI(bytes int64) string {
	if -1000 < bytes && bytes < 1000 {
		return fmt.Sprintf(" %5d ", bytes)
	}
	ci := "kMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1000
		idx++
	}
	return fmt.Sprintf("%5.1f %c", float64(bytes)/1000.0, ci[idx])
}

// converts typically bytes to human readable string (KiB, MiB, etc)
// eg: bytes=1024 -> 1 KiB
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableBytes(bytes int64) string {
	if -1024 < bytes && bytes < 1024 {
		return fmt.Sprintf("  %5d B", bytes)
	}
	ci := "KMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1024
		idx++
	}
	return fmt.Sprintf("%5.1f %ciB", float64(bytes)/1000.0, ci[idx])
}

// converts nanoseconds to human readable string (ns, µs, ms, s, m, h)
func humanReadableNanoSeconds(delay int64) string {
	var nanos = float32(delay)
	if -1000 < nanos && nanos < 1000 {
		return fmt.Sprintf(" %4.0f ns", nanos)
	}
	ci := []string{"µs", "ms", "s ", "m ", "h "}
	idx := 0
	nanos /= 1000 // µs
	if nanos <= -1000 || nanos >= 1000 {
		nanos /= 1000 // ms
		idx++
		if nanos <= -1000 || nanos >= 1000 {
			nanos /= 1000 // s
			idx++
			if nanos <= -60 || nanos >= 60 {
				nanos /= 60 // m
				idx++
				if nanos <= -60 || nanos >= 60 {
					nanos /= 60 // h
					idx++
				}
			}
		}
	}
	return fmt.Sprintf("%5.1f %s", nanos, ci[idx])
}
