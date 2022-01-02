package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

const ACK_TIMEOUT uint = 1000 // millis to wait for an ack meg
const MSG_DELAY uint = 100    // millis to wait between Rx and Tx, to give the other side time to switch from Tx to Rx
const LISTEN_DELAY uint = 50  // millis to wait between checks of the receive buffer when receiving
const USE_ENCRYPTION bool = true

// message IDs for serialization
const MESSAGE_TELEMETRY uint8 = 0
const MESSAGE_TELEMETRY_ACK uint8 = 1
const MESSAGE_COMMAND_READY uint8 = 2
const MESSAGE_COMMAND uint8 = 3
const MESSAGE_COMMAND_ACK uint8 = 4

// serialization / deserialization code on this end currently assumes that we will have
// five extra header bytes on the head of the payload that we need to (for the moment)
// ignore. RadioHead invisibly deals with these on the rover end but the rfm69 library
// we use on this end does not take them back off. the first byte is the total payload
// length (including the five header bytes) and the next four are TO, FROM, ID, FLAGS
// currently hardcoded to vec![0xff, 0xff, 0x00, 0x00]

// recognizing that I could declare serialize() / deserialize() as an interface here,
// but I'm currently not seeing a reason to.

var serbuf4 []byte = make([]byte, 4) // temp buffers for serializing, rather than make()ing a new one each time
var serbuf2 []byte = make([]byte, 2)

func serializeInt16(i int16) []byte {
	u := uint16(i)
	binary.LittleEndian.PutUint16(serbuf2, u)
	return serbuf2
}

func deserializeInt16(buf []byte) int16 {
	return int16(binary.LittleEndian.Uint16(buf[0:2]))
}

func serializeUint16(i uint16) []byte {
	binary.LittleEndian.PutUint16(serbuf2, i)
	return serbuf2
}

func deserializeUint16(buf []byte) uint16 {
	return binary.LittleEndian.Uint16(buf[0:2])
}

func serializeFloat32(f float32) []byte {
	i := math.Float32bits(f)
	binary.LittleEndian.PutUint32(serbuf4, i)
	return serbuf4
}

func deserializeFloat32(buf []byte) float32 {
	u := binary.LittleEndian.Uint32(buf[0:4])
	return math.Float32frombits(u)
}

// string serialization and deserialization functions assume 1-byte code points
// serialize will encode higher code points as '#'
func serializeString(buf *[]byte, s string) {
	for _, r := range s {
		if r > 255 {
			*buf = append(*buf, '#')
		} else {
			*buf = append(*buf, byte(r))
		}
	}
	*buf = append(*buf, 0)
}

func deserializeString(buf []byte) string {
	var b strings.Builder
	for _, r := range buf {
		fmt.Fprint(&b, r)
	}
	return b.String()
}

type RoverTimestamp struct {
	year   byte
	month  byte
	day    byte
	hour   byte
	minute byte
	second byte
}

const ROVER_TIMESTAMP_SIZE int = 6

func (rt RoverTimestamp) serialize(buf *[]byte) {
	*buf = append(*buf, rt.year)
	*buf = append(*buf, rt.month)
	*buf = append(*buf, rt.day)
	*buf = append(*buf, rt.hour)
	*buf = append(*buf, rt.minute)
	*buf = append(*buf, rt.second)
}

func (rt *RoverTimestamp) deserialize(buf []byte) {
	rt.year = buf[0]
	rt.month = buf[0]
	rt.day = buf[0]
	rt.hour = buf[0]
	rt.minute = buf[0]
	rt.second = buf[0]
}

type RoverLocData struct {
	gpsLat   float32
	gpsLong  float32
	gpsAlt   float32
	gpsSpeed float32
	gpsSats  byte
	gpsHdg   uint16
}

const ROVER_LOC_DATA_SIZE int = 19

func (rld RoverLocData) serialize(buf *[]byte) {
	*buf = append(*buf, serializeFloat32(rld.gpsLat)...)
	*buf = append(*buf, serializeFloat32(rld.gpsLong)...)
	*buf = append(*buf, serializeFloat32(rld.gpsAlt)...)
	*buf = append(*buf, serializeFloat32(rld.gpsSpeed)...)
	*buf = append(*buf, rld.gpsSats)
	*buf = append(*buf, serializeUint16(rld.gpsHdg)...)
}

func (rld *RoverLocData) deserialize(buf []byte) {
	rld.gpsLat = deserializeFloat32(buf[0:3])
	rld.gpsLong = deserializeFloat32(buf[4:7])
	rld.gpsAlt = deserializeFloat32(buf[8:11])
	rld.gpsSpeed = deserializeFloat32(buf[12:15])
	rld.gpsSats = buf[16]
	rld.gpsHdg = deserializeUint16(buf[17:18])
}

type TelemetryMessage struct {
	timestamp      RoverTimestamp
	location       RoverLocData
	signalStrength int16
	freeMemory     int16
	status         string
}

func (tm TelemetryMessage) serialize(buf *[]byte) {
	tm.timestamp.serialize(buf)
	tm.location.serialize(buf)
	*buf = append(*buf, serializeInt16(tm.signalStrength)...)
	*buf = append(*buf, serializeInt16(tm.freeMemory)...)
	serializeString(buf, tm.status)
}

func (tm *TelemetryMessage) deserialize(buf []byte) {
	start := 0                  // using these to index the beginning and end of the next subslice to be deserialized
	end := ROVER_TIMESTAMP_SIZE // to hopefully make this clearer
	tm.timestamp.deserialize(buf[start : end-1])
	start = end
	end = start + ROVER_LOC_DATA_SIZE
	tm.location.deserialize(buf[start : end-1])
	start = end
	end = start + 2
	tm.signalStrength = deserializeInt16(buf[start : end-1])
	start = end
	end = start + 2
	tm.freeMemory = deserializeInt16(buf[start : end-1])
	start = end
	tm.status = deserializeString(buf[start:])
}

type TelemetryAck struct {
	timestamp      RoverTimestamp
	ack            bool
	commandWaiting bool
}

type CommandReady struct {
	timestamp RoverTimestamp
	ready     bool
}

type CommandMessage struct {
	timestamp        RoverTimestamp
	sequenceComplete bool
	command          string
}

type CommandAck struct {
	timestamp RoverTimestamp
	ack       bool
}
