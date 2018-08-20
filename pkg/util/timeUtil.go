package util

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

// Default layout of string representation.
const defautlLayout = time.RFC3339

// ToTime returns a time.Time from a *timestamp.Timestamp.
func ToTime(t *timestamp.Timestamp) time.Time {
	return time.Unix(t.Seconds, int64(t.Nanos)).UTC()
}

// FromTime returns a new TimeConverter from time.Time.
func ToTimestamp(t time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{Seconds: t.Unix(), Nanos: 0}
}
