package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ecc1/rfm69"
)

const useEncryption bool = true

// message types for serialization
const messageTypeTelemetry byte = 0
const messageTypeTelemetryAck byte = 1
const messageTypeCommandReady byte = 2
const messageTypeCommand byte = 3
const messageTypeCommandAck byte = 4
const messageTypeNone byte = 255

func getMessageType(messageType byte) string {
	switch messageType {
	case messageTypeTelemetry:
		return "TELEMETRY"
	case messageTypeTelemetryAck:
		return "TELEMETRY_ACK"
	case messageTypeCommandReady:
		return "COMMAND_READY"
	case messageTypeCommand:
		return "COMMAND"
	case messageTypeCommandAck:
		return "COMMAND_ACK"
	default:
		return "UNKNOWN"
	}
}

// serialization / deserialization code on this end currently assumes that we will have
// five extra header bytes on the head of the payload that we need to (for the moment)
// ignore. RadioHead invisibly deals with these on the rover end but the rfm69 library
// we use on this end does not. the first byte is the total payload length (including
// the five header bytes) and the next four are TO, FROM, ID, FLAGS currently hardcoded
// to [0xff, 0xff, 0x00, 0x00]

type sendableMessage interface {
	serialize(buf *[]byte)
	length() int
}

// there's no analagous ReceivableMessage interface because Go's interface semantics
// don't allow for methods with pointer receivers.

func sendMessage(r rfm69.Radio, messageType byte, message sendableMessage) error {
	maxMessageLength := 255
	if useEncryption {
		maxMessageLength = 64
	}
	messageLength := message.length() + 6
	if messageLength > maxMessageLength {
		return fmt.Errorf("Could not send %s message, too long (%d bytes)", getMessageType(messageType), messageLength)
	}
	var buf []byte
	// payload length
	buf = append(buf, byte(messageLength))
	// static (for now) header bytes
	buf = append(buf, 0xff, 0xff, 0x00, 0x00)
	// message type
	buf = append(buf, messageType)
	// now the message
	message.serialize(&buf)
	// send it
	r.Send(buf)
	return nil
}

func receiveMessage(r rfm69.Radio, timeout time.Duration) (byte, []byte, int, error) { // messageType, messageBuf, rssi, error
	// receive a packet
	buf, rssi := r.Receive(timeout)
	if buf == nil {
		return messageTypeNone, nil, 0, nil
	}
	// throw away the 5 header bytes
	// get the message type
	messageType := buf[5]
	// the rest is the message
	messageBuf := buf[6:]
	return messageType, messageBuf, rssi, nil
}

// because there's no ReceivableMessage interface (see above), the caller needs to switch
// on messageType and call the corresponding receiveFooMessage() function:

func receiveTelemetryMessage(messageBuf []byte) (telemetryMessage, error) {
	var msg telemetryMessage
	msg.deserialize(messageBuf)
	return msg, nil
}

func receiveCommandReady(messageBuf []byte) (commandReady, error) {
	var msg commandReady
	msg.deserialize(messageBuf)
	return msg, nil
}

func receiveCommandAck(messageBuf []byte) (commandAck, error) {
	var msg commandAck
	msg.deserialize(messageBuf)
	return msg, nil
}

var serbuf4 []byte = make([]byte, 4) // temp buffers for serializing, rather than make()ing a new one each time
var serbuf2 []byte = make([]byte, 2)

// not allowing a cast from a bool to a byte is apparently the hill Go will die on
func serializeBool(b bool) byte {
	if b {
		return 1
	} else {
		return 0
	}
}

func deserializeBool(b byte) bool {
	if b > 0 {
		return true
	} else {
		return false
	}
}

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

type roverTimestamp struct {
	year   byte
	month  byte
	day    byte
	hour   byte
	minute byte
	second byte
}

const roverTimestampSize int = 6

func (rt roverTimestamp) serialize(buf *[]byte) {
	*buf = append(*buf, rt.year)
	*buf = append(*buf, rt.month)
	*buf = append(*buf, rt.day)
	*buf = append(*buf, rt.hour)
	*buf = append(*buf, rt.minute)
	*buf = append(*buf, rt.second)
}

func (rt *roverTimestamp) deserialize(buf []byte) {
	rt.year = buf[0]
	rt.month = buf[0]
	rt.day = buf[0]
	rt.hour = buf[0]
	rt.minute = buf[0]
	rt.second = buf[0]
}

type roverLocData struct {
	gpsLat   float32
	gpsLong  float32
	gpsAlt   float32
	gpsSpeed float32
	gpsSats  byte
	gpsHdg   uint16
}

const roverLocDataSize int = 19

func (rld *roverLocData) deserialize(buf []byte) {
	rld.gpsLat = deserializeFloat32(buf[0:3])
	rld.gpsLong = deserializeFloat32(buf[4:7])
	rld.gpsAlt = deserializeFloat32(buf[8:11])
	rld.gpsSpeed = deserializeFloat32(buf[12:15])
	rld.gpsSats = buf[16]
	rld.gpsHdg = deserializeUint16(buf[17:18])
}

type telemetryMessage struct { // ReceivableMessage
	timestamp      roverTimestamp
	location       roverLocData
	signalStrength int16
	freeMemory     int16
	status         string
}

func (tm *telemetryMessage) deserialize(buf []byte) {
	start := 0                // using these to index the beginning and end of the next subslice to be deserialized
	end := roverTimestampSize // to hopefully make this clearer
	tm.timestamp.deserialize(buf[start : end-1])
	start = end
	end = start + roverLocDataSize
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

type telemetryAck struct { // SendableMessage
	timestamp      roverTimestamp
	ack            bool
	commandWaiting bool
}

func (ta telemetryAck) length() int {
	return roverTimestampSize + roverLocDataSize + 1 + 1
}

func (ta telemetryAck) serialize(buf *[]byte) {
	ta.timestamp.serialize(buf)
	*buf = append(*buf, serializeBool(ta.ack))
	*buf = append(*buf, serializeBool(ta.commandWaiting))
}

type commandReady struct { // ReceivableMessage
	timestamp roverTimestamp
	ready     bool
}

func (cr *commandReady) deserialize(buf []byte) {
	start := 0                // using these to index the beginning and end of the next subslice to be deserialized
	end := roverTimestampSize // to hopefully make this clearer
	cr.timestamp.deserialize(buf[start : end-1])
	start = end
	cr.ready = deserializeBool(buf[start])
}

type commandMessage struct { // SendableMessage
	timestamp        roverTimestamp
	sequenceComplete bool
	command          string
}

func (cm commandMessage) length() int {
	return roverTimestampSize + 1 + len(cm.command) + 1
}

func (cm commandMessage) serialize(buf *[]byte) {
	cm.timestamp.serialize(buf)
	*buf = append(*buf, serializeBool(cm.sequenceComplete))
	serializeString(buf, cm.command)
}

type commandAck struct { // ReceivableMessage
	timestamp roverTimestamp
	ack       bool
}

func (ca *commandAck) deserialize(buf []byte) {
	start := 0                // using these to index the beginning and end of the next subslice to be deserialized
	end := roverTimestampSize // to hopefully make this clearer
	ca.timestamp.deserialize(buf[start : end-1])
	start = end
	ca.ack = deserializeBool(buf[start])
}
