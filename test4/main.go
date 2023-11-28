package main

import (
	"fmt"
	// "os"
	// "os/signal"
	"encoding/json"
	"strings"
	hid "github.com/dh1tw/hid"
	// streamdeck "github.com/dh1tw/streamdeck"
)

type Device struct {
	Name      string
	VendorID  uint16
	ProductID uint16
	Serial string
	Path string
}
func GetAllHIDDevices() ( result []Device ) {
	devices := hid.Enumerate( 0 , 0 )
	devices_map := make( map[ Device ]bool )
	for _ , device := range devices {
		// fmt.Println( i , device )
		// fmt.Printf( "%d === %s %s === %d === %d\n" , i , device.Manufacturer , device.Product , device.VendorID , device.ProductID )
		name := strings.TrimSpace( device.Manufacturer + " " + device.Product )
		if name == "" { continue }
		d := Device{
			Name: name ,
			VendorID: device.VendorID ,
			ProductID: device.ProductID ,
			Serial: device.Serial ,
			Path: device.Path ,
		}
		devices_map[ d ] = true
	}
	result = make( []Device , 0 , len( devices_map ) )
	for device := range devices_map { result = append( result , device ) }
	return
}

// func GetHIDDevice( serial string ) ( result *hid.Device ) {
// 	devices := hid.Enumerate( 0 , 0 )

// 	for _ , device := range devices {
// 		// fmt.Println( i , device )
// 		// fmt.Printf( "%d === %s %s === %d === %d\n" , i , device.Manufacturer , device.Product , device.VendorID , device.ProductID )
// 		name := strings.TrimSpace( device.Manufacturer + " " + device.Product )
// 		if name == "" { continue }
// 		d := Device{
// 			Name: name ,
// 			VendorID: device.VendorID ,
// 			ProductID: device.ProductID ,
// 			Serial: device.Serial ,
// 		}
// 		devices_map[ d ] = true
// 	}
// 	result = make( []Device , 0 , len( devices_map ) )
// 	for device := range devices_map { result = append( result , device ) }
// 	return
// }

func PrintDevices( devices []Device ) {
	jd , _ := json.MarshalIndent( devices , "" , "  " )
	fmt.Println( string( jd ) )
}

// https://github.com/dh1tw/streamdeck/issues/6
// TODO = put server on top of this so we can push new screen states at will
func main() {
	hid_devices := GetAllHIDDevices()
	PrintDevices( hid_devices )

// 	sd , _ := streamdeck.NewStreamDeck( "AL02K2C02319" )
// 	// fmt.Println( sd )
// 	// defer sd.ClearAllBtns()
// 	fmt.Println("using stream deck device with serial number", sd.Serial())


// 	// on_button_event := func( btn_index int , state streamdeck.BtnState ) {
// 	// 	fmt.Printf( "Button: %d , %s\n" , btn_index , state )
// 	// 	// switch state {
// 	// 	// 	case streamdeck.BtnPressed:
// 	// 	// 		// sd.WriteText( btn_index , pressedText )
// 	// 	// 		fmt.Println( btn_index , "pressed" )
// 	// 	// 	case streamdeck.BtnLongPressed:
// 	// 	// 		// sd.WriteText( btn_index , longPressedText )
// 	// 	// 		fmt.Println( btn_index , "long pressed" )
// 	// 	// 	case streamdeck.BtnReleased:
// 	// 	// 		// sd.WriteText( btn_index , releasedText )
// 	// 	// 		fmt.Println( btn_index , "released" )
// 	// 	// }
// 	// }

// 	// sd.SetBtnEventCb( on_button_event )
// 	c := make( chan os.Signal , 1 )
// 	signal.Notify( c , os.Interrupt )
// 	<-c
}