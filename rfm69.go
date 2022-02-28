package main

import (
	"log"

	sx1231 "github.com/DanCrank/rfm69-sx1231-rpi"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

const frequency uint32 = 915000000
const bitrate uint32 = 9600
const useEncryption bool = true

//see: https://forum.up-community.org/discussion/2141/solved-tutorial-gpio-i2c-spi-access-without-root-permissions

// define a Logprintf function to pass in RadioOpts
func logPrintf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func initRadio() (*sx1231.Radio, spi.PortCloser) {
	spiPort, err := spireg.Open("/dev/spidev0.1")
	if err != nil {
		log.Fatal(err)
	}

	interruptPin := gpioreg.ByName("GPIO22")
	if interruptPin == nil {
		log.Fatal("Failed to find pin GPIO22 for radio interrupt")
	}
	if err := interruptPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	csPin := gpioreg.ByName("GPIO7")
	if csPin == nil {
		log.Fatal("Failed to find pin GPIO7 for radio chip select")
	}
	if err := csPin.Out(gpio.High); err != nil {
		log.Fatal(err)
	}

	resetPin := gpioreg.ByName("GPIO25")
	if resetPin == nil {
		log.Fatal("Failed to find pin GPIO25 for radio reset")
	}
	if err := resetPin.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	syncWords := syncWords()
	radioOpts := sx1231.RadioOpts{
		Sync:    syncWords[:],
		Freq:    frequency,
		Rate:    bitrate,
		PABoost: true,
		//Logger:  logPrintf,
		Logger: nil,
	}
	radio, err := sx1231.New(spiPort, csPin, resetPin, interruptPin, radioOpts)
	if err != nil {
		log.Fatal(err)
	}
	if useEncryption {
		encryptionKey := encryptionKey()
		radio.SetEncryptionKey(encryptionKey[:])
	}
	radio.SetPower(20)
	return radio, spiPort
}
