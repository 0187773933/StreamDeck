package main

import (
	"fmt"
	"os"
	"time"
	"os/signal"
	"syscall"
	"path/filepath"
	utils "github.com/0187773933/StreamDeck/v1/utils"
	ui_wrapper "github.com/0187773933/StreamDeck/v1/ui"
	server "github.com/0187773933/StreamDeck/v1/server"
	bolt_api "github.com/boltdb/bolt"
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
		s.UI.DB.Close()
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
	ui.Connect()
	defer ui.Device.Close()
	var active_page_id string
	if len( os.Args ) > 2 {
		active_page_id = os.Args[ 2 ]
	} else {
		active_page_id = "default"
	}
	db , _ := bolt_api.Open( config.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	ui.DB = db
	ui.SetActivePageID( active_page_id )
	fmt.Println( ui )
	ui.Clear()
	ui.Render()
	ui.SetBrightness( ui.Brightness )
	go ui.WatchKeys()

	// 2.) Start Server
	SetupCloseHandler()
	// utils.GenerateNewKeys()
	s = server.New( config , ui )
	s.Start()

}
