package main

import (
	"fmt"
	"log"

	"github.com/DanCrank/rfm69"
)

const frequency uint32 = 915000000
const bitrate uint32 = 9600

func initRadioEcc1() *rfm69.Radio {
	r := rfm69.Open()
	if r.Error() != nil {
		log.Fatal(r.Error())
	}

	log.Print("Resetting radio")
	r.Reset()

	config := rfm69.DefaultConfiguration()
	config[rfm69.RegDataModul] = rfm69.PacketMode | rfm69.ModulationTypeFSK | 0<<rfm69.ModulationShapingShift
	// Use PA1 with 17 dBm output power.
	config[rfm69.RegPaLevel] = rfm69.Pa1On | rfm69.Pa2On | 0x1F<<rfm69.OutputPowerShift
	// Default != reset value
	config[rfm69.RegLna] = /* rfm69.LnaZin | */ 1<<rfm69.LnaCurrentGainShift | 0<<rfm69.LnaGainSelectShift
	// Interrupt on DIO0 when Sync word is seen.
	// Cleared when leaving Rx or FIFO is emptied.
	//config[rfm69.RegDioMapping1] = 2 << rfm69.Dio0MappingShift
	config[rfm69.RegDioMapping1] = 0
	// Default != reset value.
	config[rfm69.RegDioMapping2] = 5 << rfm69.ClkOutShift
	// Default != reset value.
	//config[rfm69.RegRssiThresh] = 0xE4
	// Make sure enough preamble bytes are sent.
	config[rfm69.RegPreambleMsb] = 0x00
	config[rfm69.RegPreambleLsb] = 0x04
	// Use 4 bytes for Sync word.
	config[rfm69.RegSyncConfig] = rfm69.SyncOn | 1<<rfm69.SyncSizeShift
	// Sync word.
	sw := syncWords()
	config[rfm69.RegSyncValue1] = sw[0]
	config[rfm69.RegSyncValue2] = sw[1]
	config[rfm69.RegSyncValue3] = 0
	config[rfm69.RegSyncValue4] = 0
	config[rfm69.RegSyncValue5] = 0
	config[rfm69.RegSyncValue6] = 0
	config[rfm69.RegSyncValue7] = 0
	config[rfm69.RegSyncValue8] = 0
	// encryption key
	ec := encryptionKey()
	config[rfm69.RegAesKey1] = ec[0]
	config[rfm69.RegAesKey2] = ec[1]
	config[rfm69.RegAesKey3] = ec[2]
	config[rfm69.RegAesKey4] = ec[3]
	config[rfm69.RegAesKey5] = ec[4]
	config[rfm69.RegAesKey6] = ec[5]
	config[rfm69.RegAesKey7] = ec[6]
	config[rfm69.RegAesKey8] = ec[7]
	config[rfm69.RegAesKey9] = ec[8]
	config[rfm69.RegAesKey10] = ec[9]
	config[rfm69.RegAesKey11] = ec[10]
	config[rfm69.RegAesKey12] = ec[11]
	config[rfm69.RegAesKey13] = ec[12]
	config[rfm69.RegAesKey14] = ec[13]
	config[rfm69.RegAesKey15] = ec[14]
	config[rfm69.RegAesKey16] = ec[15]
	config[rfm69.RegPacketConfig1] = rfm69.VariableLength | 2<<rfm69.DcFreeShift | rfm69.CrcOn
	config[rfm69.RegPayloadLength] = 64
	config[rfm69.RegFifoThresh] = rfm69.TxStartFifoNotEmpty | 15<<rfm69.FifoThresholdShift
	config[rfm69.RegPacketConfig2] = rfm69.AutoRxRestartOn | rfm69.AesOn | 0<<rfm69.InterPacketRxDelayShift
	// misc settings to match what the rust code did
	config[rfm69.RegFdevMsb] = 0x01
	config[rfm69.RegFdevLsb] = 0x38 // rover has 0x3B, working rust code had 0x38...maybe just a typo?
	//config[rfm69.RegRxBw] = 2<<rfm69.DccFreqShift | rfm69.RxBwMant20 | 4<<rfm69.RxBwExpShift  // RxBwRxBwMant = 0b01 / 20, RxBwExp = 4, this is what the rust code had
	//config[rfm69.RegAfcBw] = 2<<rfm69.DccFreqShift | rfm69.RxBwMant20 | 4<<rfm69.RxBwExpShift // other sx1231 driver used 0xF4 0xF4
	config[rfm69.RegRxBw] = 0xEC  // rover has 0xF4, working rust code had 0xEC
	config[rfm69.RegAfcBw] = 0xEC // rover has 0xF4, working rust code had 0xEC
	config[rfm69.RegRssiThresh] = 0xFF
	//r.SetChannelBW(100000) // didn't figure out what this works out to be
	r.WriteConfiguration(config, true)
	r.SetFrequency(frequency)
	r.SetBitrate(bitrate)
	// Default != reset value.
	//r.hw.WriteRegister(rfm69.RegTestDagc, 0x30)

	log.Println("")
	dumpRegisters(r)
	log.Print("Sleeping")
	r.Sleep()

	return r
}

func dumpRegisters(r *rfm69.Radio) {
	regs := r.ReadConfiguration(true)
	log.Println("     0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F")
	for i := 0; i < len(regs); i += 16 {
		line := fmt.Sprintf("%02x:", i)
		for j := 0; j < 16 && i+j < len(regs); j++ {
			line += fmt.Sprintf(" %02x", regs[i+j])
		}
		log.Println(line)
	}
}
