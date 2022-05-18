package main

import (
	"image"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"
)

const telemetryTimeout time.Duration = 10 * time.Second  // time to wait for a telemetry message from the rover
const ackTimeout time.Duration = 1000 * time.Millisecond // time to wait for an ack message from the rover
const msgDelay time.Duration = 100 * time.Millisecond    // time to wait between Rx and Tx, to give the other side time to switch from Tx to Rx

func initHost() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
}

func testDisplay() {
	b, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	opts := ssd1306.DefaultOpts
	opts.Sequential = true

	dev, err := ssd1306.NewI2C(b, &opts)
	if err != nil {
		log.Fatalf("failed to initialize ssd1306: %v", err)
	}

	// Draw on it.
	img := image1bit.NewVerticalLSB(dev.Bounds())
	f := basicfont.Face7x13
	drawer := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: f,
		Dot:  fixed.P(0, img.Bounds().Dy()-1-f.Descent),
	}
	drawer.DrawString("GROUND STATION READY")
	if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}
}

func main() {
	initHost()
	testDisplay()
	r := initRadioEcc1()
	//defer spiPort.Close()
	// main loop - receive telemetry packets
	messageCount := 0
	groundRssi := 0.0
	roverRssi := 0.0
	retries := 0.0
	for {
		msgType, buf, rssi, err := receiveMessage(r, telemetryTimeout)
		if err != nil {
			log.Fatal(err)
		}
		switch msgType {
		case messageTypeTelemetry:
			log.Print("Received TELEMETRY message:")
			log.Printf("RSSI: %d", rssi)
			tm, err := receiveTelemetryMessage(buf)
			spew.Dump(tm)
			// accumulate stats
			messageCount++
			if rssi != 0 {
				groundRssi = ((groundRssi * float64(messageCount-1)) + float64(rssi)) / float64(messageCount)
			}
			if tm.signalStrength != 0 {
				roverRssi = ((roverRssi * float64(messageCount-1)) + float64(tm.signalStrength)) / float64(messageCount)
			}
			retries = ((retries * float64(messageCount-1)) + float64(tm.retries)) / float64(messageCount)
			log.Printf("***Avg ground RSSI: %.2f   Avg rover RSSI: %.2f   Avg retries: %.2f", groundRssi, roverRssi, retries)
			if err != nil {
				log.Fatal(err)
			}
			// brief pause to give the otehr side time to switch from Tx to Rx
			time.Sleep(msgDelay)
			// send the ack
			ta := telemetryAck{
				timestamp:      currentTimestamp(),
				ack:            true,
				commandWaiting: false,
			}
			log.Print("Sending TELEMETRY_ACK:")
			err = sendMessage(r, messageTypeTelemetryAck, ta)
			if err != nil {
				// TODO retry sending the ACK
				log.Fatal(err)
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
