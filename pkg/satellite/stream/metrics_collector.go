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
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
)

// InstantMinSamples - Minimum number of samples to calculate instantaneous stats with (rate & delay)
const InstantMinSamples = 5

// InstantMaxSamples - Maximum number of samples to calculate instantaneous stats with (rate & delay)
const InstantMaxSamples = 100

// InstantSampleSeconds - Duration of data samples to calculate instantaneous stats with
const InstantSampleSeconds = 5

type telemetryWithTimestamp struct {
	ReceivedTime         time.Time
	DataBytes            int
	TimeLastByteReceived *timestamp.Timestamp
}

// MetricsCollector holds metrics used to display pass report and instantaneous stats
type MetricsCollector struct {
	planId                        string
	streamId                      string
	timerStart                    time.Time
	elapsed                       float64
	totalBytesReceived            int64
	totalMessagesReceived         int64
	azimuth                       float64
	elevation                     float64
	frequency                     float64
	delayNanos                    int64
	throttleCheckSchedulerRunning bool

	messageBuffer                 []telemetryWithTimestamp
	starpassTimeFirstByteReceived *timestamp.Timestamp
	starpassTimeLastByteReceived  *timestamp.Timestamp
	localTimeFirstByteReceived    *timestamp.Timestamp
	localTimeLastByteReceived     *timestamp.Timestamp
	logger                        func(format string, v ...interface{})
}

// NewMetricsCollector creates a stats collector
func NewMetricsCollector(logger func(format string, v ...interface{})) *MetricsCollector {
	logger("[STATS] using local time to calculate telemetry delay")
	return &MetricsCollector{
		logger:                        logger,
		throttleCheckSchedulerRunning: false,
	}
}

// planId is used to identify when to reset the statistics; upon reset, stats is printed to output
func (metrics *MetricsCollector) setPlanId(planId string) {
	if metrics.planId != planId {
		metrics.logReport()
		metrics.planId = planId
		metrics.reset()
	}
}

func (metrics *MetricsCollector) setStreamId(streamId string) {
	metrics.streamId = streamId
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
	metrics.starpassTimeFirstByteReceived = nil
	metrics.starpassTimeLastByteReceived = nil
	metrics.localTimeFirstByteReceived = nil
	metrics.localTimeLastByteReceived = nil
}

// collects metrics for telemetry data message
func (metrics *MetricsCollector) collectTelemetry(telemetry *stellarstation.Telemetry) {
	metrics.delayNanos += time.Now().UTC().UnixNano() - ((telemetry.TimeLastByteReceived.Seconds * 1e9) + int64(telemetry.TimeLastByteReceived.Nanos))
	metrics.collectMessage(len(telemetry.Data))

	// update first and last byte timestamp for the pass
	if metrics.starpassTimeFirstByteReceived == nil {
		metrics.starpassTimeFirstByteReceived = telemetry.TimeFirstByteReceived
		metrics.localTimeFirstByteReceived = timestampNow()
	}
	metrics.starpassTimeLastByteReceived = telemetry.TimeLastByteReceived
	metrics.localTimeLastByteReceived = timestampNow()

	if len(metrics.messageBuffer) == 0 || metrics.messageBuffer[len(metrics.messageBuffer)-1].ReceivedTime.UnixNano() < time.Now().UnixNano()-(1e6) {
		// if no samples, or most recent sample arrive later than 1 milliseconds of last sample, save details for instantaneous rates
		msg := telemetryWithTimestamp{
			ReceivedTime:         time.Now(),
			DataBytes:            len(telemetry.Data),
			TimeLastByteReceived: telemetry.TimeLastByteReceived,
		}
		metrics.messageBuffer = append(metrics.messageBuffer, msg)
	} else {
		// merge sample with newest sample if current timestamp is within 1 milliseconds of most recent sample
		// this is done to improve stats reporting performance, but we discard instantaneous info about TimeLastByteReceived
		metrics.messageBuffer[len(metrics.messageBuffer)-1].DataBytes += len(telemetry.Data)
	}

	// Keep 5 seconds worth of samples, but no less than InstantMinSamples samples, and no more than InstantMaxSamples; remove oldest sample if:
	// 1. list larger than InstantMinSamples && oldest sample is older than "now - InstantSampleSeconds"
	// 2. list larger than InstantMaxSamples
	for (len(metrics.messageBuffer) > InstantMinSamples && metrics.messageBuffer[0].ReceivedTime.UnixNano() < time.Now().UnixNano()-(InstantSampleSeconds*1e9)) ||
		len(metrics.messageBuffer) > InstantMaxSamples {
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

func toTime(ts *timestamp.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := time.Unix(ts.Seconds, int64(ts.Nanos))
	return &t
}

func timestampNow() *timestamp.Timestamp {
	now := time.Now().UTC()
	return &timestamp.Timestamp{
		Seconds: now.Unix(),
		Nanos:   int32(now.Nanosecond()),
	}
}

func formatTimeLocal(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.Format(time.RFC3339)
}

func formatTimestampLocal(ts *timestamp.Timestamp) string {
	if ts == nil {
		return ""
	}
	return toTime(ts).Format(time.RFC3339)
}

func formatTimestampUTC(ts *timestamp.Timestamp) string {
	if ts == nil {
		return ""
	}
	return toTime(ts).UTC().Format(time.RFC3339)
}

func duration(start, end *timestamp.Timestamp) string {
	if start == nil || end == nil {
		return ""
	}
	return fmt.Sprintf("%s", toTime(end).Sub(*toTime(start)))
}

func (metrics *MetricsCollector) logReport() {
	if metrics.totalMessagesReceived > 0 {
		// Dont use metrics.logger because it might be in overwrite mode
		logger := fmt.Printf
		logger("\n\n")
		logger("[STATS] %s, Pass summary:\n", time.Now().Format("20060102 15:04:05"))
		logger("\n")
		logger("  Plan ID   : %s\n", metrics.planId)
		logger("  Stream ID : %s\n", metrics.streamId)
		logger("\n")
		logger("  Datatake (Starpass timestamp)\n")
		logger("  First byte received   : %s (UTC %s)\n", formatTimestampLocal(metrics.starpassTimeFirstByteReceived), formatTimestampUTC(metrics.starpassTimeFirstByteReceived))
		logger("  Last  byte received   : %s (UTC %s)\n", formatTimestampLocal(metrics.starpassTimeLastByteReceived), formatTimestampUTC(metrics.starpassTimeLastByteReceived))
		logger("  Duration              : %s\n", duration(metrics.starpassTimeFirstByteReceived, metrics.starpassTimeLastByteReceived))
		logger("\n")
		logger("  CLI data receive (local timestamp)\n")
		logger("  First chunk received  : %s (%s after datatake first byte)\n", formatTimestampLocal(metrics.localTimeFirstByteReceived), duration(metrics.starpassTimeFirstByteReceived, metrics.localTimeFirstByteReceived))
		logger("  Last  chunk received  : %s (%s after datatake last byte)\n", formatTimestampLocal(metrics.localTimeLastByteReceived), duration(metrics.starpassTimeLastByteReceived, metrics.localTimeLastByteReceived))
		logger("  Total bytes received  : %d (%s)\n", metrics.totalBytesReceived, humanReadableBytes(metrics.totalBytesReceived))
		logger("  Total chunks          : %d\n", metrics.totalMessagesReceived)
		logger("  Average rate (bits/s) : %sbps\n", humanReadableCountSI(metrics.avgRate()))
		logger("  Average delay         : %s\n", humanReadableNanoSeconds(metrics.avgDelay()))
		logger("\n\n")
	}
}

// report instantaneous statistics
func (metrics *MetricsCollector) logStats() {
	iDelayNanos := humanReadableNanoSeconds(metrics.instantDelay())
	iRateStr := humanReadableCountSI(metrics.instantRate())
	size := humanReadableBytes(metrics.totalBytesReceived)
	metrics.logger("[STATS] %s, plan_id: %s, azm: %5.2f, ele: %5.1f, freq: %5.1f MHz [DATA] %3d msgs, bytes: %9v, rate: %9vbps, delay: %9v",
		time.Now().Format("20060102 15:04:05"), metrics.planId, metrics.azimuth, metrics.elevation, metrics.frequency, metrics.totalMessagesReceived, size, iRateStr, iDelayNanos)
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
	duration := float64(metrics.messageBuffer[len(metrics.messageBuffer)-1].ReceivedTime.UnixNano()-metrics.messageBuffer[0].ReceivedTime.UnixNano()) / float64(1e9)
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
		return fmt.Sprintf("%d ", bytes)
	}
	ci := "kMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1000
		idx++
	}
	return fmt.Sprintf("%.1f %c", float64(bytes)/1000.0, ci[idx])
}

// converts typically bytes to human readable string (KiB, MiB, etc)
// eg: bytes=1024 -> 1 KiB
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableBytes(bytes int64) string {
	if -1024 < bytes && bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	ci := "KMGTPE"
	idx := 0
	for bytes <= -999_950 || bytes >= 999_950 {
		bytes /= 1024
		idx++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/1000.0, ci[idx])
}

// converts nanoseconds to human readable string (ns, µs, ms, s, m, h)
func humanReadableNanoSeconds(delay int64) string {
	var nanos = float32(delay)
	if -1000 < nanos && nanos < 1000 {
		return fmt.Sprintf("%.0f ns", nanos)
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
	return fmt.Sprintf("%.1f %s", nanos, ci[idx])
}

// StartStatsEmitScheduler this should be ran in separate thred
func (metrics *MetricsCollector) startStatsEmitSchedulerWorker(emitRateMillis int) {
	metrics.throttleCheckSchedulerRunning = true
	uptimeTicker := time.NewTicker(time.Duration(emitRateMillis) * time.Millisecond)
	for {
		<-uptimeTicker.C
		if metrics.throttleCheckSchedulerRunning {
			// check for expired samples
			for len(metrics.messageBuffer) > 0 && metrics.messageBuffer[0].ReceivedTime.UnixNano() < time.Now().UnixNano()-(InstantSampleSeconds*1e9) {
				metrics.messageBuffer = metrics.messageBuffer[1:]
			}
			metrics.logStats()
		} else {
			// stop scheduler
			return
		}
	}
}

// StartStatsEmitScheduler start process to emit stats at defined interval
func (metrics *MetricsCollector) StartStatsEmitScheduler(emitRateMillis int) {
	go metrics.startStatsEmitSchedulerWorker(emitRateMillis)
}

// StopStatsEmitScheduler stop the emittting stats process
func (metrics *MetricsCollector) StopStatsEmitScheduler(emitRateMillis int) {
	metrics.throttleCheckSchedulerRunning = false
}
