package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"path/filepath"
	utils "github.com/0187773933/StreamDeck/v1/utils"
	ui_wrapper "github.com/0187773933/StreamDeck/v1/ui"
	server "github.com/0187773933/StreamDeck/v1/server"
)

var s server.Server

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		fmt.Println( "\r- Ctrl+C pressed in Terminal" )
		fmt.Println( "Shutting Down StreamDeck Server" )
		s.FiberApp.Shutdown()
		os.Exit( 0 )
	}()
}

func main() {

	config_file_path := "./config.yaml"
	if len( os.Args ) > 1 { config_file_path , _ = filepath.Abs( os.Args[ 1 ] ) }
	config := utils.ParseConfig( config_file_path )
	fmt.Printf( "Loaded Config File From : %s\n" , config_file_path )

	// 1.) Setup StreamDeck
	ui := ui_wrapper.NewStreamDeckUIFromInterface( &config.StreamDeckUI )
	ui.AddDevice()
	defer ui.Device.Close()
	ui.ActivePageID = "default"
	fmt.Println( ui )
	// // ui.ActivePageID = "spotify-triple"
	ui.Render()
	go ui.WatchKeys()

	// 2.) Start Server
	SetupCloseHandler()
	// utils.GenerateNewKeys()
	s = server.New( config , ui )
	s.Start()

}