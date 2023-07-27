package ui

import (
	"os"
	"io"
	// "path/filepath"
	"fmt"
	"time"
	// "net/url"
	"os/exec"
	"strings"
	streamdeck_wrapper "github.com/muesli/streamdeck"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	resize "github.com/nfnt/resize"
	ioutil "io/ioutil"
	yaml "gopkg.in/yaml.v2"
	http "net/http"
	oto "github.com/hajimehoshi/oto"
	mp3 "github.com/hajimehoshi/go-mp3"
	try "github.com/manucorporat/try"
)

// https://github.com/muesli/streamdeck/blob/master/streamdeck.go#L112
// StreamDeck OriginalV2 = 72x72
// fmt.Println( device.Pixels )
const IMAGE_SIZE uint = 72

func get_image_data( file_path string ) ( result *image.RGBA ) {
	imgFile , err := os.Open( file_path )
	if err != nil { fmt.Println( "Error:" , err ); return }
	defer imgFile.Close()
	img , _ , err := image.Decode( imgFile )
	if err != nil { fmt.Println( "Error:" , err ); return }
	resizedImg := resize.Resize( IMAGE_SIZE , IMAGE_SIZE , img , resize.Lanczos3 )
	rgba := image.NewRGBA( image.Rect( 0 , 0 , int( IMAGE_SIZE ) , int( IMAGE_SIZE ) ) )
	draw.Draw( rgba , rgba.Bounds() , resizedImg , resizedImg.Bounds().Min , draw.Src )
	result = rgba
	return
}

func get_json( url string ) ( result string ) {
	req , req_err := http.NewRequest( "GET" , url , nil )
	if req_err != nil { fmt.Println( req_err ); return }
	req.Header.Set( "Content-Type" , "application/json" )
	client := &http.Client{}
	resp , resp_err := client.Do( req )
	if resp_err != nil { fmt.Println( resp_err ); return }
	body , body_err := ioutil.ReadAll( resp.Body )
	if body_err != nil { fmt.Println( body_err ); return }
	result = string( body )
	resp.Body.Close()
	return
}

type Button struct {
	PressCount int `yaml:"-"`
	LastPressTime time.Time `yaml:"-"`
	// Toggled bool `yamm:"toggled"`
	Timer *time.Timer `yaml:"-"`
	Index uint8 `yaml:"index"`
	Image string `yaml:"image"`
	MP3 string `yaml:"mp3"`
	Id string `yaml:"_"`
	SingleClick string `yaml:"single_click"`
	DoubleClick string `yaml:"double_click"`
	TripleClick string `yaml:"triple_click"`
	Toggle string `yaml:"toggle"`
}

type PageButton struct {
	Index uint8 `yaml:"index"`
	Id string `yaml:"id"`
}
type StreamDeckUIPage struct {
	Buttons []PageButton
}
type StreamDeckUI struct {
	Device streamdeck_wrapper.Device `yaml:"-"`
	ActivePageID string `yaml:"-"`
	Serial string `yaml:"serial"`
	Brightness int `yaml:"brightness"`
	GlobalCooldownMilliseconds int `yaml:"global_cooldown_milliseconds"`
	EndpointHostName string `yaml:"endpoint_hostname"`
	EndpointToken string `yaml:"endpoint_token"`
	Pages map[string]StreamDeckUIPage `yaml:"pages"`
	Buttons map[string]Button `yaml:"buttons"`
	LoadedButtonImages map[uint8]string `yaml:"-"`
}

func ( ui *StreamDeckUI ) AddDevice() {
	devs , _ := streamdeck_wrapper.Devices()
	if len( devs ) < 1 {
		fmt.Println( "No Devices Found" )
		os.Exit( 1 )
	}
	ui.Device = devs[ 0 ]
	open_error := ui.Device.Open()
	if open_error != nil {
		fmt.Printf( "can't open device: %s" , open_error )
		os.Exit( 1 )
	}
	// ui.Device.Clear()
}
func ( ui *StreamDeckUI ) set_image( button_index uint8 , file_path string ) {
	if ui.LoadedButtonImages[ button_index ] != file_path {
		ui.LoadedButtonImages[ button_index ] = file_path
	} else {
		return
	}
	image_data := get_image_data( file_path )
	err := ui.Device.SetImage( button_index , image_data )
	if err != nil {
		fmt.Printf( "Cannot set image: %s" , err )
		os.Exit( 1 )
	}
}

func ( ui *StreamDeckUI ) isPageID( test string ) ( result bool ) {
	_ , exists := ui.Pages[ test ]
	result = exists
	return
}

func ( ui *StreamDeckUI ) is_endpoint_url( input_url string ) ( result bool ) {
	// _ , err := url.ParseRequestURI( input_url )
	// return err == nil
	result = false
	if strings.Contains( input_url , ui.EndpointHostName ) { result = true }
	return
}


func ( ui *StreamDeckUI ) btn_num_to_page_button( button_index uint8 ) ( result Button ) {
	for _ , button := range ui.Pages[ ui.ActivePageID ].Buttons {
		if button.Index == button_index {
			result = ui.Buttons[ button.Id ]
			result.Id = button.Id
			return
		}
	}
	return
}

func ( ui *StreamDeckUI ) Clear() { ui.Device.Clear() }

func ( ui *StreamDeckUI ) Render() {
	// ui.Device.Clear()
	for _ , button := range ui.Pages[ ui.ActivePageID ].Buttons {
		btn := ui.Buttons[ button.Id ]
		ui.set_image( button.Index , btn.Image )
		// Initialize Button State
		btn.PressCount = 0
		btn.LastPressTime = time.Now()
		ui.Buttons[ button.Id ] = btn
	}
}

func (ui *StreamDeckUI) WatchKeys() {
	key_channel , err := ui.Device.ReadKeys()
	if err != nil {
		fmt.Printf( "Error reading keys: %v\n" , err )
		os.Exit( 1 )
	}

	for key := range key_channel {
		button := ui.btn_num_to_page_button(key.Index)
		if key.Pressed {
			now := time.Now()
			if now.Sub(button.LastPressTime) > time.Second {
				button.PressCount = 0
			}
			button.PressCount++
			button.LastPressTime = now

			if button.PressCount > 3 {
				button.PressCount = 0
			}

			if button.Timer != nil {
				button.Timer.Stop()
			}

			button.Timer = time.AfterFunc( ( time.Millisecond * 500 ) , func() {
				buttonPressCount := button.PressCount
				switch buttonPressCount {
					case 1:
						// fmt.Println( "Single Click" )
						if button.SingleClick == "" { fmt.Println( "Single Click not Registered" ); break }
						fmt.Println( button.Index , "Single Click" , button.SingleClick )
						if ui.isPageID( button.SingleClick ) {
							ui.ActivePageID = button.SingleClick
							ui.Render()
							break;
						} else if ui.is_endpoint_url( button.SingleClick ) {
							if button.MP3 != "" {
								CWD , _ := os.Getwd()
								go ui.PlayMP3( fmt.Sprintf( "%s/%s" , CWD , button.MP3 ) )
							}
							get_json( fmt.Sprintf( "%s/%s?%s" , ui.EndpointHostName , button.SingleClick , ui.EndpointToken ) )
						} else {
							fmt.Println( "we are exec-ing this ????" )
							fmt.Println( button.SingleClick )
							cmd := exec.Command( "bash" , "-c" , button.SingleClick )
							cmd.Start()
						}
					case 2:
						// fmt.Println( "Double Click" )
						if button.DoubleClick == "" { fmt.Println( "Double Click not Registered" ); break }
						fmt.Println( button.Index , "Double Click" , button.DoubleClick )
						if ui.isPageID( button.DoubleClick ) {
							ui.ActivePageID = button.DoubleClick
							ui.Render()
							break;
						} else if ui.is_endpoint_url( button.DoubleClick ) {
							if button.MP3 != "" {
								CWD , _ := os.Getwd()
								go ui.PlayMP3( fmt.Sprintf( "%s/%s" , CWD , button.MP3 ) )
							}
							get_json( fmt.Sprintf( "%s/%s?%s" , ui.EndpointHostName , button.DoubleClick , ui.EndpointToken ) )
						} else {
							fmt.Println( "we are exec-ing this ????" )
							fmt.Println( button.DoubleClick )
							cmd := exec.Command( "bash" , "-c" , button.DoubleClick )
							cmd.Start()
						}
					case 3:
						// fmt.Println( "Triple Click" )
						if button.TripleClick == "" { fmt.Println( "Triple Click not Registered" ); break }
						fmt.Println( button.Index , "Triple Click" , button.TripleClick )
						if ui.isPageID( button.TripleClick ) {
							ui.ActivePageID = button.TripleClick
							ui.Render()
							break;
						} else if ui.is_endpoint_url( button.TripleClick ) {
							if button.MP3 != "" {
								CWD , _ := os.Getwd()
								go ui.PlayMP3( fmt.Sprintf( "%s/%s" , CWD , button.MP3 ) )
							}
							get_json( fmt.Sprintf( "%s/%s?%s" , ui.EndpointHostName , button.TripleClick , ui.EndpointToken ) )
						} else {
							fmt.Println( "we are exec-ing this ????" )
							fmt.Println( button.TripleClick )
							cmd := exec.Command( "bash" , "-c" , button.TripleClick )
							cmd.Start()
						}
				}
				button.PressCount = 0
			})

			// Update the button in the ui.Buttons map
			if button.Toggle != "" {
				for i , x_button := range ui.Pages[ ui.ActivePageID ].Buttons {
					if x_button.Index == key.Index {
						ui.Pages[ ui.ActivePageID ].Buttons[ i ].Id = button.Toggle
						break
					}
				}
				ui.Render()
			}
			ui.Buttons[button.Id] = button
		} else {
			// when the key is released
			// Just ignore this event
		}
	}
}

func ( ui *StreamDeckUI ) PlayMP3( file_path string ) {
	try.This( func() {
		f, err := os.Open( file_path )
		defer f.Close()
		if err != nil { fmt.Println( err ) }
		d , err := mp3.NewDecoder( f )
		if err != nil { fmt.Println( err ) }
		c , err := oto.NewContext( d.SampleRate() , 2 , 2 , 8192 )
		defer c.Close()
		if err != nil { fmt.Println( err ) }
		p := c.NewPlayer()
		defer p.Close()
		if _ , err := io.Copy( p , d ); err != nil { fmt.Println( err ) }
	}).Catch( func( e try.E ) {
		fmt.Println( e )
	})
}

func NewStreamDeckUI( file_path string ) ( result *StreamDeckUI ) {
	file , _ := ioutil.ReadFile( file_path )
	error := yaml.Unmarshal( file , &result )
	if error != nil { panic( error ) }
	return
}

func NewStreamDeckUIFromInterface( config *interface{} ) ( result *StreamDeckUI ) {
	intermediate , _ := yaml.Marshal( config )
	error := yaml.Unmarshal( intermediate , &result )
	result.LoadedButtonImages = make(map[uint8]string)
	if error != nil { panic( error ) }
	return
}