package main

import (
	"fmt"
	"image/color"
	"time"

	streamdeck "github.com/magicmonkey/go-streamdeck"
	"github.com/magicmonkey/go-streamdeck/actionhandlers"
	"github.com/magicmonkey/go-streamdeck/buttons"
	_ "github.com/magicmonkey/go-streamdeck/devices"
)


// https://github.com/magicmonkey/go-streamdeck/blob/v0.0.5/comms.go#L82
func main() {
	sd , err := streamdeck.New()
	if err != nil { panic( err ) }

	// A simple yellow button in position 26
	cButton := buttons.NewColourButton( color.RGBA{ 255 , 255 , 0 , 255 } )
	sd.AddButton( 0 , cButton )

	// A button with text on it in position 2, which echoes to the console when presesd
	myButton := buttons.NewTextButton( "hola" )
	myButton.SetActionHandler( &actionhandlers.TextPrintAction{ Label: "You pressed me" } )
	sd.AddButton( 2 , myButton )

	// A button with text on it which changes when pressed\
	counter := 7
	myNextButton := buttons.NewTextButton( string( counter ) )
	test := func( button streamdeck.Button ) {
		fmt.Println( "here , there" )
	}
	myNextButtonOnPress := actionhandlers.NewCustomAction( test )
	myNextButton.SetActionHandler( myNextButtonOnPress )
	sd.AddButton( 7 , myNextButton )

	// A button which performs multiple actions when pressed
	multiActionButton := buttons.NewColourButton(color.RGBA{255, 0, 255, 255})
	thisActionHandler := &actionhandlers.ChainedAction{}
	thisActionHandler.AddAction(&actionhandlers.TextPrintAction{Label: "Purple press"})
	thisActionHandler.AddAction(&actionhandlers.ColourChangeAction{NewColour: color.RGBA{255, 0, 0, 255}})
	multiActionButton.SetActionHandler(thisActionHandler)
	sd.AddButton(27, multiActionButton)

	time.Sleep(20 * time.Second)
}