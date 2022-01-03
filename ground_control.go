package main

import (
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const telemetryTimeout time.Duration = 10000000 // 10 seconds

func main() {
	r := initRadio()
	// main loop - receive telemetry packets
	for {
		msgType, buf, rssi, error := receiveMessage(*r, telemetryTimeout)
		if error != nil {
			log.Fatal(r.Error())
		}
		switch msgType {
		case messageTypeTelemetry:
			log.Print("Received TELEMETRY message:")
			log.Printf("RSSI: %d", rssi)
			tm, error := receiveTelemetryMessage(buf)
			spew.Dump(tm)
			if error != nil {
				log.Fatal(r.Error())
			}
			// send the ack
			ta := telemetryAck{
				timestamp:      currentTimestamp(),
				ack:            true,
				commandWaiting: false,
			}
			error = sendMessage(*r, messageTypeTelemetryAck, ta)
			if error != nil {
				log.Fatal(r.Error())
			}
		default:
			log.Printf("Unexpected message type: %s", getMessageType(msgType))
		}
	}
}

func currentTimestamp() roverTimestamp {
	now := time.Now().Local()
	t := roverTimestamp{
		year:   byte(now.Year() - 2000),
		month:  byte(now.Month()),
		day:    byte(now.Day()),
		hour:   byte(now.Hour()),
		minute: byte(now.Minute()),
		second: byte(now.Second()),
	}
	return t
}

// TODO
// func setup_display() {

// }
