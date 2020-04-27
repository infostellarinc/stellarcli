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
	timerStart            time.Time
	elapsed               float64
	totalBytesReceived    int64
	totalMessagesReceived int64
	logger                func(format string, v ...interface{})
}

func NewMetricsCollector(logger func(format string, v ...interface{})) *MetricsCollector {
	return &MetricsCollector{
		timerStart:            time.Now(),
		totalBytesReceived:    0,
		totalMessagesReceived: 0,
		logger:                logger,
	}
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

// report collected statistics
func (metrics *MetricsCollector) logStats() {
	if metrics.totalMessagesReceived > 0 {
		rate := humanReadableCountSI(int64(float64(metrics.totalBytesReceived) / metrics.elapsed * 8.00))
		size := humanReadableBytes(metrics.totalBytesReceived)
		metrics.logger("[STATS] total: %6d msgs, bytes: %s, rate: %sbps", metrics.totalMessagesReceived, size, rate)
	}
}

// converts number (typically bytes or bits) to human readable SI string
// eg: bytes=1200 -> 1.2k
// 	   bytes=1200000 0> 1.2M
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableCountSI(bytes int64) string {
	if -1000 < bytes && bytes < 1000 {
		return fmt.Sprintf("%6d ", bytes)
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
		return fmt.Sprintf("%6d ", bytes)
	}
	ci := "KMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1024
		idx++
	}
	return fmt.Sprintf("%6.1f %ciB", float64(bytes)/1000.0, ci[idx])
}
