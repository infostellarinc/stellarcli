package interactive

import (
	"hash"
	"strconv"
	"sync"

	stellarstation "github.com/infostellarinc/go-stellarstation/api/v1"
)

type TransmitterState struct {
	sweepOn       bool
	modulationOn  bool
	carrierOn     bool
	idlePatternOn bool
}

func (cs TransmitterState) toStringColumns() []string {
	return []string{
		greenRedBoolText(cs.sweepOn, false, strconv.FormatBool(cs.sweepOn)),
		greenRedBoolText(cs.modulationOn, false, strconv.FormatBool(cs.modulationOn)),
		greenRedBoolText(cs.carrierOn, false, strconv.FormatBool(cs.carrierOn)),
		greenRedBoolText(cs.idlePatternOn, false, strconv.FormatBool(cs.idlePatternOn)),
	}
}

type ReceiverState struct {
	phaseBitLock  bool
	frameLock     bool
	normalizedSnr float64
}

func (cs ReceiverState) toStringColumns() []string {
	return []string{
		greenRedBoolText(cs.phaseBitLock, false, strconv.FormatBool(cs.phaseBitLock)),
		greenRedBoolText(cs.frameLock, false, strconv.FormatBool(cs.frameLock)),
		strconv.FormatFloat(cs.normalizedSnr, 'f', 4, 64),
	}
}

type AntennaState struct {
	estMaxElevation float64
	elevation       float64
	azimuth         float64
}

func (as AntennaState) toStringColumns() []string {
	return []string{
		strconv.FormatFloat(as.estMaxElevation, 'f', 2, 64),
		strconv.FormatFloat(as.elevation, 'f', 2, 64),
		strconv.FormatFloat(as.azimuth, 'f', 2, 64),
	}
}

type DataState struct {
	totalPayloadBytes           uint64
	receivedEndTelemetryMessage bool
	inboundCrc32cStr            string

	configurationChangeSentCount uint64
	commandSentCount             uint64
}

func (ds DataState) toStringColumns() []string {
	return []string{
		strconv.FormatUint(ds.totalPayloadBytes, 10),
	}
}

// StreamState will be written to directly from the streamer in the background
// this should be a separate state from the view
type StreamState struct {
	currentTxState      TransmitterState
	currentRxState      ReceiverState
	currentAntennaState AntennaState
	streamId            string
	lastAck             string
	closed              bool
	err                 error

	totalPayloadBytes           uint64
	receivedEndTelemetryMessage bool

	inboundCrc32c hash.Hash32
}

// consider this group within the stateMux lock
// --- stateMux group start
var stateMux sync.Mutex
var streamState StreamState
var STREAM_CLIENT stellarstation.StellarStationService_OpenSatelliteStreamClient

// --- stateMux group end
