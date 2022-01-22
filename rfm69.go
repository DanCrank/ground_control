package main

import (
	"log"
	"time"

	"github.com/DanCrank/devices/sx1231"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
)

const frequency uint32 = 915000000
const bitrate uint32 = 50000
const useEncryption bool = true
const hwDelay time.Duration = 100 * time.Millisecond //generic delay to allow hardware to cope with non-root access
//see: https://forum.up-community.org/discussion/2141/solved-tutorial-gpio-i2c-spi-access-without-root-permissions

// define a Logprintf function to pass in RadioOpts
func logPrintf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func initRadio() *sx1231.Radio {
	//spiPort, err := spireg.Open("")
	spiPort, err := spireg.Open("SPI0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer spiPort.Close()
	time.Sleep(hwDelay)

	interruptPin := gpioreg.ByName("GPIO19")
	if interruptPin == nil {
		log.Fatal("Failed to find pin GPIO19 for radio interrupt")
	}
	if err := interruptPin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}
	time.Sleep(hwDelay)

	syncWords := syncWords()
	radioOpts := sx1231.RadioOpts{
		Sync:    syncWords[:],
		Freq:    frequency,
		Rate:    bitrate,
		PABoost: true,
		Logger:  logPrintf,
	}
	radio, err := sx1231.New(spiPort, interruptPin, radioOpts)
	if err != nil {
		log.Fatal(err)
	}
	if useEncryption {
		encryptionKey := encryptionKey()
		radio.SetEncryptionKey(encryptionKey[:])
	}
	return radio
}
