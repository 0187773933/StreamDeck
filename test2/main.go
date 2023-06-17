package main

import (
	"fmt"
	utils "github.com/0187773933/StreamDeck/v1/utils"
	streamdeck "github.com/magicmonkey/go-streamdeck"
)


func WrapGetStreamDeck() {
	hid_device := utils.GetHIDDevice( "AL02K2C02319" )
	defer hid_device.Close()


}


// https://github.com/magicmonkey/go-streamdeck/blob/v0.0.5/comms.go#L82
func main() {
	WrapGetStreamDeck()
}