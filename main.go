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
	types "github.com/0187773933/StreamDeck/v1/types"
	server "github.com/0187773933/StreamDeck/v1/server"
	bolt_api "github.com/boltdb/bolt"
	"github.com/google/gousb"
	"os/exec"
	"regexp"
	"bufio"
	"strings"
	"strconv"
	"bytes"
	// try "github.com/manucorporat/try"
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
		s.UI.Device.Close()
		os.Exit( 0 )
	}()
}
var CONNECTED bool = false
var WATCHING_SERIAL_NUMBER string = ""
// var WATCHING_VENDOR_ID string = ""
// var WATCHING_PRODUCT_ID string = ""
func watch_usb_events( restart_chan chan bool ) {
	known_devices := make( map[string]bool )
	ctx := gousb.NewContext()
	defer ctx.Close()

	for {
		cmd := exec.Command( "lsusb" )
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			fmt.Printf( "Error running lsusb: %v\n" , err )
			time.Sleep( 2 * time.Second )
			continue
		}

		scanner := bufio.NewScanner( &out )
		current_devices := make( map[string]bool )

		deviceRE := regexp.MustCompile( `Bus (\d+) Device (\d+): ID (\w+):(\w+) (.+)` )
		for scanner.Scan() {
			line := scanner.Text()
			if deviceRE.MatchString( line ) {
				matches := deviceRE.FindStringSubmatch( line )

				deviceID := fmt.Sprintf( "%s:%s" , matches[ 3 ] , matches[ 4 ] )
				current_devices[ deviceID ] = true

				if !known_devices[ deviceID ] {
					vendorID , err := strconv.ParseUint( matches[ 3 ] , 16 , 16 )
					if err != nil {
						fmt.Printf( "Error parsing vendor ID: %v\n" , err )
						continue
					}
					productID , err := strconv.ParseUint( matches[ 4 ] , 16 , 16 )
					if err != nil {
						fmt.Printf( "Error parsing product ID: %v\n" , err )
						continue
					}
					fmt.Printf( "Device connected: %s\n" , line )
					dev , err := ctx.OpenDeviceWithVIDPID( gousb.ID( vendorID ) , gousb.ID( productID ) )
					if err != nil {
						fmt.Printf( "Error opening device: %v\n" , err )
						continue
					}
					if dev == nil {
						fmt.Println( "Device not found." )
						continue
					}

					manufacturer , _ := dev.Manufacturer()
					serial_number , _ := dev.SerialNumber()
					product , _ := dev.Product()
					dev.Close()

					fmt.Printf( "Manufacturer: %s , Serial Number: %s , Product: %s\n" , manufacturer , serial_number , product )
					fmt.Println( serial_number , WATCHING_SERIAL_NUMBER )
					if serial_number == WATCHING_SERIAL_NUMBER {
						fmt.Println( "StreamDeck connected , restarting" )
						restart_chan <- true
					}
				}
			}
		}

		// Detect and log disconnected devices
		for deviceID := range known_devices {
			if !current_devices[ deviceID ] {
				// fmt.Printf( "Device disconnected: %s\n" , deviceID )
				parts := strings.Split( deviceID, ":" )
				vendorID := parts[ 0 ]
				productID := parts[ 1 ]
				fmt.Println( vendorID , productID , "disconnected" )
				// restart_chan <- true
			}
		}

		known_devices = current_devices
		time.Sleep( 10 * time.Second )
	}
}

func restart_ui( ui *ui_wrapper.StreamDeckUI , config types.ConfigFile ) {
	if ui.DB != nil {
		if err := ui.DB.Close(); err != nil {
			fmt.Printf( "Error closing database: %v\n" , err )
		}
		ui.DB = nil
	}
	if ui.Device.Serial != "" {
		if err := ui.Device.Close(); err != nil {
			fmt.Printf( "Error closing device: %v\n" , err )
		}
	}
	ui.Connect()
	if ui.Ready {
		fmt.Println( "StreamDeck Connecting" )
		var activePageID string
		if len(os.Args) > 2 {
			activePageID = os.Args[2]
		} else {
			activePageID = "default"
		}
		db, err := bolt_api.Open(config.BoltDBPath, 0600, &bolt_api.Options{Timeout: 3 * time.Second})
		if err != nil {
			fmt.Printf("Failed to open database: %v\n", err)
			return
		}
		ui.DB = db
		ui.SetActivePageID(activePageID)
		fmt.Println(ui)
		ui.Clear()
		ui.Render()
		ui.SetBrightness(ui.Brightness)
		go ui.WatchKeys() // Ensure any previous instances are stopped or managed
	} else {
		fmt.Println("StreamDeck Not Connected")
	}
}

func main() {
	config_file_path := "./config.yaml"
	if len( os.Args ) > 1 { config_file_path , _ = filepath.Abs( os.Args[ 1 ] ) }
	config := utils.ParseConfig( config_file_path )
	fmt.Printf( "Loaded Config File From : %s\n" , config_file_path )
	// WATCHING_SERIAL_NUMBER = config.StreamDeckUI.(map[interface{}]interface{})["serial"].(string)
	WATCHING_SERIAL_NUMBER = config.StreamDeckUI.Serial

	SetupCloseHandler()
	// ui_wrapper.PrintDevices()

	restart_chan := make( chan bool )
	go watch_usb_events( restart_chan )

	ui := ui_wrapper.NewStreamDeckUIFromInterface( &config.StreamDeckUI )
	restart_ui( ui , config )

	s = server.New( config , ui )
	go func() {
		s.Start()
	}()

	for {
		select {
		case <-restart_chan:
			fmt.Println( "Restarting UI" )
			restart_ui( s.UI , config )
		}
	}
}

