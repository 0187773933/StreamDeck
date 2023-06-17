package main

import (
	"fmt"
	utils "github.com/0187773933/StreamDeck/v1/utils"
	ui_wrapper "github.com/0187773933/StreamDeck/v1/ui"
)

func main() {
	// utils.GenerateNewKeys()
	fmt.Println( "here" )
	config := utils.ParseConfig( "./config.yaml" )
	ui := ui_wrapper.NewStreamDeckUIFromInterface( &config.StreamDeckUI )
	ui.AddDevice()
	defer ui.Device.Close()
	ui.ActivePageID = "default"
	// // ui.ActivePageID = "spotify-triple"
	ui.Render()
	ui.WatchKeys()
}