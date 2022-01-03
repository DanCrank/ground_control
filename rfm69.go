package main

import (
	"log"

	"github.com/ecc1/rfm69"
)

const frequency uint32 = 915000000
const bandwidth uint32 = 100000
const bitrate uint32 = 9600

const ackTimeout uint = 1000 // millis to wait for an ack meg
const msgDelay uint = 100    // millis to wait between Rx and Tx, to give the other side time to switch from Tx to Rx
const listenDelay uint = 50  // millis to wait between checks of the receive buffer when receiving

func initRadio() *rfm69.Radio {
	r := rfm69.Open()
	if r.Error() != nil {
		log.Fatal(r.Error())
	}
	r.Reset()
	r.InitRF(frequency)
	r.SetChannelBW(bandwidth)
	r.SetBitrate(bitrate)
	dumpRF(r)
	r.Sleep()
	return r
}

func dumpRF(r *rfm69.Radio) {
	if r.Error() != nil {
		log.Fatal(r.Error())
	}
	log.Printf("Mode: %s", r.State())
	log.Printf("Frequency: %d Hz", r.Frequency())
	mod := r.ReadModulationType()
	switch mod {
	case rfm69.ModulationTypeFSK:
		log.Printf("Modulation type: FSK")
	case rfm69.ModulationTypeOOK:
		log.Printf("Modulation type: OOK")
	default:
		log.Panicf("Unknown modulation mode %X", mod)
	}
	log.Printf("Bitrate: %d baud", r.Bitrate())
	log.Printf("Channel BW: %d Hz", r.ChannelBW())
}
