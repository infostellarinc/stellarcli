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
)

type MetricsCollector struct {
	planId                string
	timerStart            time.Time
	elapsed               float64
	totalBytesReceived    int64
	totalMessagesReceived int64
	azimuth               float64
	elevation             float64
	frequency             float64
	logger                func(format string, v ...interface{})
}

// run with "go test -v" in this folder to see output
func NewMetricsCollector(logger func(format string, v ...interface{})) *MetricsCollector {
	return &MetricsCollector{
		logger: logger,
	}
}

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
}

// record 1 message received with size=messageSizeBytes
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
	rate := int64(0)
	if metrics.totalMessagesReceived > 1 {
		rate = int64(float64(metrics.totalBytesReceived) / metrics.elapsed * 8.00)
	}
	rateStr := humanReadableCountSI(rate)
	size := humanReadableBytes(metrics.totalBytesReceived)
	metrics.logger("[STATS] plan_id: %s, azm: %6.2f, ele: %6.2f, freq: %6.1f MHz [DATA] %5d msgs, bytes: %s, rate: %sbps",
		metrics.planId, metrics.azimuth, metrics.elevation, metrics.frequency, metrics.totalMessagesReceived, size, rateStr)
}

// converts number (typically bytes or bits) to human readable SI string
// eg: bytes=1200 -> 1.2k
// 	   bytes=1200000 0> 1.2M
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableCountSI(bytes int64) string {
	if -1000 < bytes && bytes < 1000 {
		return fmt.Sprintf(" %6d ", bytes)
	}
	ci := "kMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1000
		idx++
	}
	return fmt.Sprintf("%6.1f %c", float64(bytes)/1000.0, ci[idx])
}

// converts typically bytes to human readable string (KiB, MiB, etc)
// eg: bytes=1024 -> 1 KiB
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableBytes(bytes int64) string {
	if -1024 < bytes && bytes < 1024 {
		return fmt.Sprintf("  %6d B", bytes)
	}
	ci := "KMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1024
		idx++
	}
	return fmt.Sprintf("%6.1f %ciB", float64(bytes)/1000.0, ci[idx])
}
