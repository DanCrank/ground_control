module github.com/dancrank/ground_control

go 1.17

require (
	github.com/davecgh/go-spew v1.1.1
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	periph.io/x/conn/v3 v3.6.10
	periph.io/x/devices/v3 v3.6.13
	periph.io/x/host/v3 v3.7.2
)

//replace github.com/DanCrank/rfm69 => /home/ubuntu/rfm69

require github.com/DanCrank/rfm69 v0.0.4

require (
	github.com/ecc1/gpio v0.0.0-20200212231225-d40e43fcf8f5 // indirect
	github.com/ecc1/radio v0.0.0-20200419171134-0864efbcd270 // indirect
	github.com/ecc1/spi v0.0.0-20200419165236-942b6408d3f6 // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
)
