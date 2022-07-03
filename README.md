# Prototype rover ground station code

## About

This is part of a second iteration of a rover vehicle (the first
being https://github.com/DanCrank/tank-bot-mk1). This repo will
eventually grow into a ground station that can send commands to
the rover, receive telemetry from the rover, and present some kind
of user interface.

The rover code can be found at https://github.com/DanCrank/rover.

## Hardware references
This code targets the following hardware...

Raspberry Pi (any)

Adafruit RFM69HCW Transceiver Radio Bonnet - 868 / 915 MHz <br>
https://learn.adafruit.com/adafruit-radio-bonnets <br>
https://cdn-shop.adafruit.com/product-files/3076/RFM69HCW-V1.1.pdf

Notes on converting to LoRa:
- Uses RFM9x rather than RFM69
- Uses SX127x rather than SX1231
- Extra sensitivity/range comes from using "spread-spectrum" rather than FSK/OOK
- Pinouts for the bonnet / featherwing should be same as RFM69
- Datasheet: https://github.com/SeeedDocument/RFM95-98_LoRa_Module/blob/master/RFM95_96_97_98_DataSheet.pdf
- Alternate: https://cdn.sparkfun.com/assets/learn_tutorials/8/0/4/RFM95_96_97_98W.pdf
- Possible Go driver: https://pkg.go.dev/git.morran.xyz/morran/go-sx127x
- Found a Go driver for sx126x at https://github.com/tinygo-org/drivers/blob/release/sx126x/sx126x.go
- TinyGo sx127x driver grinding along at https://github.com/tinygo-org/drivers/pull/60
- Arduino library at https://jgromes.github.io/LoRaLib/class_r_f_m96.html
- Adafruit CircuitPython here: https://github.com/adafruit/Adafruit_CircuitPython_RFM9x
- RFM9x apparently expands to RFM95/6/7/8. 433MHz version is RFM96/98
- RFM96 and RFM98 are the same, or at least no one knows the difference (see https://jgromes.github.io/LoRaLib/class_r_f_m98.html)
- Supports packet lengths up to 2k bytes (!) but does _not_ support encryption in hardware (!)