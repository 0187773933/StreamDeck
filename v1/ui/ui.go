package ui

import (
	"os"
	"io"
	// "path/filepath"
	"fmt"
	"time"
	// "net/url"
	"os/exec"
	"sync"
	// "reflect"
	"strings"
	"bytes"
	"net/url"
	slug "github.com/gosimple/slug"
	"encoding/json"
	"path/filepath"
	// streamdeck_wrapper "github.com/muesli/streamdeck"
	streamdeck_wrapper "github.com/0187773933/StreamDeck/v1/streamdeck"
	// types "github.com/0187773933/StreamDeck/v1/types"
	// utils "github.com/0187773933/StreamDeck/v1/utils"
	"image"
	"image/draw"
	// _ "image/jpeg"
	// _ "image/png"
	"image/jpeg"
	"image/png"
	resize "github.com/nfnt/resize"
	ioutil "io/ioutil"
	yaml "gopkg.in/yaml.v2"
	http "net/http"
	oto "github.com/hajimehoshi/oto"
	mp3 "github.com/hajimehoshi/go-mp3"
	try "github.com/manucorporat/try"
	bolt_api "github.com/boltdb/bolt"
	twilio "github.com/twilio/twilio-go"
	twilio_api "github.com/twilio/twilio-go/rest/api/v2010"
)

// https://github.com/muesli/streamdeck/blob/master/streamdeck.go#L112
// StreamDeck OriginalV2 = 72x72
// fmt.Println( device.Pixels )
// const IMAGE_SIZE uint = 72

// func get_image_data( file_path string ) ( result *image.RGBA ) {
// 	imgFile , err := os.Open( file_path )
// 	if err != nil { fmt.Println( "Error:" , err ); return }
// 	defer imgFile.Close()
// 	img , _ , err := image.Decode( imgFile )
// 	if err != nil { fmt.Println( "Error:" , err ); return }
// 	resizedImg := resize.Resize( IMAGE_SIZE , IMAGE_SIZE , img , resize.Lanczos3 )
// 	rgba := image.NewRGBA( image.Rect( 0 , 0 , int( IMAGE_SIZE ) , int( IMAGE_SIZE ) ) )
// 	draw.Draw( rgba , rgba.Bounds() , resizedImg , resizedImg.Bounds().Min , draw.Src )
// 	result = rgba
// 	return
// }

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

func post_json( url string , headers map[string]string , payload interface{} ) ( result interface{} ) {
	client := &http.Client{}
	payload_bytes , err := json.Marshal( payload )
	if err != nil { fmt.Println( err ); return }
	req , err := http.NewRequest( "POST" , url , bytes.NewBuffer( payload_bytes ) )
	if err != nil { fmt.Println( err ); return }
	req.Header.Set( "Content-Type" , "application/json" )
	for key, value := range headers {
		req.Header.Set( key , value )
	}
	resp , err := client.Do( req )
	if err != nil { fmt.Println( err ); return }
	defer resp.Body.Close()
	body , err := ioutil.ReadAll( resp.Body )
	if err != nil { fmt.Println( err ); return }
	json.Unmarshal( body , &result )
	return
}

type Button struct {
	PressCount int `yaml:"press_count"`
	LastPressTime time.Time `yaml:"last_press_time"`
	// Toggled bool `yamm:"toggled"`
	Timer *time.Timer `yaml:"-"`
	Index uint8 `yaml:"index"`
	Image string `yaml:"image"`
	MP3 string `yaml:"mp3"`
	Id string `yaml:"id"`
	SingleClick string `yaml:"single_click"`
	DoubleClick string `yaml:"double_click"`
	TripleClick string `yaml:"triple_click"`
	Toggle string `yaml:"toggle"`
	ReturnPage string `yaml:"return_page"`
	Options map[string]string `yaml:"options"`
}

type PageButton struct {
	Index uint8 `yaml:"index"`
	Id string `yaml:"id"`
}

type StreamDeckUIPage struct {
	Buttons []PageButton
}

type TwilioConfig struct {
	SID string `yaml:"sid"`
	Token string `yaml:"token"`
	From string `yaml:"from"`
	APIKeySID string `yaml:"api_key_sid"`
	APIKeySecret string `yaml:"api_key_secret"`
}

type PushOverConfig struct {
	Token string `yaml:"token"`
	GlobalNotify bool `yaml:"global_notify"`
	GlobalNotifyTo string `yaml:"global_notify_to"`
	GlobalNotifySound string `yaml:"global_notify_sound"`
}

type StreamDeckUI struct {
	Device streamdeck_wrapper.Device `yaml:"-"`
	Ready bool `yaml:"-"`
	Muted bool `yaml:"muted"`
	Fresh bool `yaml:"-"`
	SettingsMode bool `yaml:"-"`
	PlayBackMutex sync.Mutex `yaml:"-"`
	TwilioCallMutex sync.Mutex `yaml:"-"`
	ActivePageID string `yaml:"-"`
	Sleep bool `yaml:"-"`
	Serial string `yaml:"serial"`
	VendorID string `yaml:"vendor_id"`
	ProductID string `yaml:"product_id"`
	IconSize uint `yaml:"icon_size"`
	Brightness uint8 `yaml:"brightness"`
	XSize int `yaml:"x_size"`
	YSize int `yaml:"y_size"`
	GlobalCooldownMilliseconds int64 `yaml:"global_cooldown_milliseconds"`
	LastPressTime time.Time `yaml:"-"`
	EndpointHostName string `yaml:"endpoint_hostname"`
	EndpointToken string `yaml:"endpoint_token"`
	Twilio TwilioConfig `yaml:"twilio"`
	PushOver PushOverConfig `yaml:"push_over"`
	TwilioLocked bool `yaml:"-"`
	Pages map[string]StreamDeckUIPage `yaml:"pages"`
	Buttons map[string]Button `yaml:"buttons"`
	LoadedButtonImages map[uint8]string `yaml:"-"`
	DB *bolt_api.DB `yaml:"-"`
}

func ( ui *StreamDeckUI ) Connect() {
	devs , error := streamdeck_wrapper.Devices()
	if error != nil { panic( error ) }
	if len( devs ) < 1 {
		fmt.Println( "No Devices Found" )
		ui.Ready = false
		// os.Exit( 1 )
		return
	}
	for _ , dev := range devs {
		if dev.Serial == ui.Serial {
			ui.Device = dev
			break
		}
	}
	ui.Device = devs[ 0 ]
	open_error := ui.Device.Open()
	if open_error != nil {
		fmt.Printf( "can't open device: %s" , open_error )
		// os.Exit( 1 )
		ui.Ready = false
		return
	}
	ui.Ready = true
	// ui.Device.Clear()
}

func ( ui *StreamDeckUI ) LoadDB() {
	fmt.Println( "LoadDB" )
}

func ( ui *StreamDeckUI ) StoreDB() {
	fmt.Println( "StoreDB" )
	// Buttons map[string]Button `yaml:"buttons"`
	// LoadedButtonImages map[uint8]string `yaml:"-"`
	for _ , button := range ui.Buttons {
		btn := ui.Buttons[ button.Id ]
		fmt.Println( btn )
		// image_data := ui.get_image_data( btn.Image )
	}
}

func GetDevices() ( result []streamdeck_wrapper.Device ) {
	result , error := streamdeck_wrapper.Devices()
	if error != nil { panic( error ) }
	if len( result ) < 1 {
		fmt.Println( "No Devices Found" )
		os.Exit( 1 )
	}
	return
}

func PrintDevices() ( result []streamdeck_wrapper.Device ) {
	result , error := streamdeck_wrapper.Devices()
	if error != nil { panic( error ) }
	if len( result ) < 1 {
		fmt.Println( "No Devices Found" )
		os.Exit( 1 )
	}
	for i , dev := range result {
		fmt.Printf( "%d === %s\n" , i , dev.Serial )
	}
	return
}


func ( ui *StreamDeckUI ) get_image_data( file_path string ) ( result *image.RGBA ) {
	imgFile , err := os.Open( file_path )
	if err != nil { fmt.Println( "Error:" , err ); return }
	defer imgFile.Close()
	img , _ , err := image.Decode( imgFile )
	if err != nil { fmt.Println( "Error:" , err ); return }
	resizedImg := resize.Resize( ui.IconSize , ui.IconSize , img , resize.Lanczos3 )
	rgba := image.NewRGBA( image.Rect( 0 , 0 , int( ui.IconSize ) , int( ui.IconSize ) ) )
	draw.Draw( rgba , rgba.Bounds() , resizedImg , resizedImg.Bounds().Min , draw.Src )
	result = rgba
	return
}

func ( ui *StreamDeckUI ) set_image( button_index uint8 , file_path string ) {
	image_data := ui.get_image_data( file_path )
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

func ( ui *StreamDeckUI ) SetActivePageID( page_id string ) ( result string ) {
	ui.DB.Update( func( tx *bolt_api.Tx ) error {
		tmp2_bucket , _ := tx.CreateBucketIfNotExists( []byte( "tmp2" ) )
		tmp2_bucket.Put( []byte( "active-page-id" ) , []byte( page_id ) )
		return nil
	})
	return
}

func ( ui *StreamDeckUI ) GetActivePageID() ( result string ) {
	ui.DB.View( func( tx *bolt_api.Tx ) error {
		tmp2_bucket := tx.Bucket( []byte( "tmp2" ) )
		bucket_result := tmp2_bucket.Get( []byte( "active-page-id" ) )
		result = string( bucket_result )
		return nil
	})
	return
}

// doesn't persist
func ( ui *StreamDeckUI ) AddPage( page_id string , page StreamDeckUIPage ) ( result string ) {
	ui.Pages[ page_id ] = page
	return
}

// doesn't persist
func ( ui *StreamDeckUI ) AddButton( button_id string , button Button ) ( result string ) {
	ui.Buttons[ button_id ] = button
	return
}

func ( ui *StreamDeckUI ) BtnNumToPageButton( button_index uint8 ) ( result Button ) {
	page_id := ui.GetActivePageID()
	for _ , button := range ui.Pages[ page_id ].Buttons {
		if button.Index == button_index {
			result = ui.Buttons[ button.Id ]
			result.Id = button.Id
			return
		}
	}
	return
}

func ( ui *StreamDeckUI ) BtnIdToPageButton( button_id string ) ( result Button ) {
	page_id := ui.GetActivePageID()
	for _ , button := range ui.Pages[ page_id ].Buttons {
		if button.Id == button_id {
			result = ui.Buttons[ button.Id ]
			result.Id = button.Id
			return
		}
	}
	return
}

func ( ui *StreamDeckUI ) SetBrightness( brightness_level uint8 ) {
	ui.Device.SetBrightness( brightness_level )
}

func ( ui *StreamDeckUI ) DecreaseBrightness() {
	ui.Brightness = ( ui.Brightness - 10 )
	if ui.Brightness <= 0 {
		ui.Brightness = 0
		ui.Sleep = true
	}
	ui.Device.SetBrightness( ui.Brightness )
}

func ( ui *StreamDeckUI ) IncreaseBrightness() {
	ui.Brightness = ( ui.Brightness + 10 )
	if ui.Brightness > 100 { ui.Brightness = 100 }
	ui.Device.SetBrightness( ui.Brightness )
}

func ( ui *StreamDeckUI ) Clear() { ui.Device.Clear() }

func ( ui *StreamDeckUI ) Show() {
	ui.Sleep = false
	// ui.Device.Wake()
	ui.Device.SetBrightness( 100 )
	ui.Brightness = 100
}
func ( ui *StreamDeckUI ) Hide() {
	ui.Sleep = true
	// ui.Device.Sleep()
	ui.Device.SetBrightness( 0 )
	ui.Brightness = 0
}

func ( ui *StreamDeckUI ) Mute() {
	ui.Muted = true
}

func ( ui *StreamDeckUI ) UnMute() {
	ui.Muted = false
}

func ( ui *StreamDeckUI ) RenderSoft() {
	// ui.Device.Clear()
	page_id := ui.GetActivePageID()
	if strings.HasPrefix( page_id , "settings" ) {
		ui.SettingsMode = true
	} else {
		ui.SettingsMode = false
	}
	for _ , button := range ui.Pages[ page_id ].Buttons {
		btn := ui.Buttons[ button.Id ]
		// ????
		if ui.LoadedButtonImages[ button.Index ] != btn.Image {
			ui.LoadedButtonImages[ button.Index ] = btn.Image
		} else {
			break;
		}
		ui.set_image( button.Index , btn.Image )
		// Initialize Button State
		btn.PressCount = 0
		btn.LastPressTime = time.Now()
		ui.Buttons[ button.Id ] = btn
	}
}
func ( ui *StreamDeckUI ) Render() {
	// ui.Device.Clear()
	page_id := ui.GetActivePageID()
	if strings.HasPrefix( page_id , "settings" ) {
		ui.SettingsMode = true
	} else {
		ui.SettingsMode = false
	}
	fmt.Println( "Active Page ID ===" , page_id )
	for _ , button := range ui.Pages[ page_id ].Buttons {
		btn := ui.Buttons[ button.Id ]
		ui.set_image( button.Index , btn.Image )
		// Initialize Button State
		btn.PressCount = 0
		btn.LastPressTime = time.Now()
		ui.Buttons[ button.Id ] = btn
	}
}

func ( ui *StreamDeckUI ) SingleClickNumber( button_num uint8 ) {
	fmt.Println( "Single Click" )
	button := ui.BtnNumToPageButton( button_num )
	if button.SingleClick == "" { fmt.Println( "Single Click not Registered" ); return; }
	fmt.Println( button.Index , "Single Click" , button.SingleClick )
	if ui.isPageID( button.SingleClick ) {
		ui.SetActivePageID( button.SingleClick )
		ui.Clear()
		ui.Render()
		return
	} else if ui.is_endpoint_url( button.SingleClick ) {
		if button.MP3 != "" {
			CWD , _ := os.Getwd()
			if ui.Muted == false {
				go ui.PlayMP3( fmt.Sprintf( "%s/%s" , CWD , button.MP3 ) )
			}
		}
		x_url := ""
		if strings.Contains( button.SingleClick , "?" ) {
			x_url = fmt.Sprintf( "%s&%s" , button.SingleClick , ui.EndpointToken )
		} else {
			x_url = fmt.Sprintf( "%s?%s" , button.SingleClick , ui.EndpointToken )
		}
		get_json( x_url )
	} else {
		fmt.Printf( "exec-ing: %s\n" , button.SingleClick )
		cmd := exec.Command( "bash" , "-c" , button.SingleClick )
		cmd.Start()
	}
}

func ( ui *StreamDeckUI ) SingleClickId( button_id string ) {
	fmt.Println( "Single Click" )
	button := ui.BtnIdToPageButton( button_id )
	if button.SingleClick == "" { fmt.Println( "Single Click not Registered" ); return; }
	fmt.Println( button.Index , "Single Click" , button.SingleClick )
	if ui.isPageID( button.SingleClick ) {
		ui.SetActivePageID( button.SingleClick )
		ui.Clear()
		ui.Render()
		return
	} else if ui.is_endpoint_url( button.SingleClick ) {
		if button.MP3 != "" {
			CWD , _ := os.Getwd()
			if ui.Muted == false {
				go ui.PlayMP3( fmt.Sprintf( "%s/%s" , CWD , button.MP3 ) )
			}
		}
		x_url := ""
		if strings.Contains( button.SingleClick , "?" ) {
			x_url = fmt.Sprintf( "%s&%s" , button.SingleClick , ui.EndpointToken )
		} else {
			x_url = fmt.Sprintf( "%s?%s" , button.SingleClick , ui.EndpointToken )
		}
		get_json( x_url )
	} else {
		fmt.Printf( "exec-ing: %s\n" , button.SingleClick )
		cmd := exec.Command( "bash" , "-c" , button.SingleClick )
		cmd.Start()
	}
}

func ( ui *StreamDeckUI ) TwilioCall( to string , url string ) {
	ui.TwilioCallMutex.Lock()
	ui.TwilioLocked = true
	client := twilio.NewRestClientWithParams( twilio.ClientParams{
		Username: ui.Twilio.SID ,
		Password: ui.Twilio.Token ,
	})
	params := &twilio_api.CreateCallParams{}
	params.SetTo( to )
	params.SetFrom( ui.Twilio.From )
	params.SetUrl( url )
	resp , err := client.Api.CreateCall( params )
	var call_sid string
	if err != nil {
		fmt.Println( err.Error() )
	} else {
		fmt.Println( "Call Status: " + *resp.Status )
		fmt.Println( "Call Sid: " + *resp.Sid )
		fmt.Println( "Call Direction: " + *resp.Direction )
		call_sid = *resp.Sid
		fetch_call_params := &twilio_api.FetchCallParams{}
		answered := false
		completed := false
		for i := 0; i < 120; i++ {
			time.Sleep( 1 * time.Second )
			call , err := client.Api.FetchCall( call_sid , fetch_call_params )
			if err != nil {
				fmt.Println( err.Error() )
			} else {
				status := *call.Status
				fmt.Printf( "Call Status: %s\n" , status )
				if status == "in-progress" {
					answered = true
				} else if status == "completed" {
					completed = true
					break
				}
			}
		}
		fmt.Println( "done waiting for call" , answered , completed )
	}
	ui.TwilioLocked = false
	ui.TwilioCallMutex.Unlock()
}


func ( ui *StreamDeckUI ) PushOverSend( to string , message string , sound string ) {
	headers := map[string]string{
		"Accept": "application/json, text/plain, */*" ,
	}
	params := map[string]string{
		"token": ui.PushOver.Token ,
		"user": to ,
		"message": message ,
		"sound": sound ,
	}
	post_json( "https://api.pushover.net/1/messages.json" , headers , params )
}

func ( ui *StreamDeckUI ) ButtonAction( button Button , action_type string , action string , mp3_path string ) {
	fmt.Println( button.Index , action_type , action )

	if strings.HasPrefix( action , "settings" ) == false && ui.SettingsMode == false {
		if button.Options[ "push_over_to" ] != "" {
			go ui.PushOverSend( button.Options[ "push_over_to" ] , button.Options[ "push_over_message" ] , button.Options[ "push_over_sound" ] )
		} else if ui.PushOver.GlobalNotify == true {
			message := fmt.Sprintf( "Pressed %s - %s - %s\n" , button.Id , action_type , action )
			go ui.PushOverSend( ui.PushOver.GlobalNotifyTo , message , ui.PushOver.GlobalNotifySound )
		}
	}

	if action == "twilio-call" {
		if ui.TwilioLocked == false {
			go ui.TwilioCall( button.Options[ "to" ] , button.Options[ "url" ] )
		} else {
			fmt.Println( "Already a Twilio Call in Progress , Not Queing Another One" )
		}
	} else if action == "settings-brightness-increase" {
		ui.IncreaseBrightness()
	} else if action == "settings-brightness-decrease" {
		ui.DecreaseBrightness()
	} else if action == "settings-show" {
		ui.Show()
	} else if action == "settings-hide" {
		ui.Hide()
	} else if action == "settings-mute" {
		ui.Mute()
	} else if action == "settings-unmute" {
		ui.UnMute()
	} else if ui.isPageID( action ) {
		ui.SetActivePageID( action )
		ui.Clear()
		ui.Render()
	} else if ui.is_endpoint_url( action ) {
		if mp3_path != "" {
			CWD, _ := os.Getwd()
			if ui.Muted == false {
				go ui.PlayMP3( fmt.Sprintf( "%s/%s" , CWD , mp3_path ) )
			}
		}
		x_url := ""
		if strings.Contains( button.SingleClick , "?" ) {
			x_url = fmt.Sprintf( "%s&%s" , action , ui.EndpointToken )
		} else {
			x_url = fmt.Sprintf( "%s?%s" , action , ui.EndpointToken )
		}
		go get_json( x_url )
	} else {
		fmt.Printf( "exec-ing: %s\n", action )
		cmd := exec.Command( "bash", "-c" , action )
		go cmd.Start()
	}
	if button.ReturnPage != "" {
		ui.SetActivePageID( button.ReturnPage )
		ui.Clear()
		ui.Render()
	}
}

func ( ui *StreamDeckUI ) WatchKeys() {
	key_channel , err := ui.Device.ReadKeys()
	if err != nil {
		fmt.Printf( "Error reading keys: %v\n" , err )
		os.Exit( 1 )
	}
	for key := range key_channel {
		button := ui.BtnNumToPageButton( key.Index )
		if key.Pressed {
			now := time.Now()

			if now.Sub( button.LastPressTime ) > time.Second {
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

			if ui.Sleep {
				fmt.Println( "here again" )
				ui.Show()
				ui.Sleep = false
				ui.Muted = false
				ui.Brightness = 100
				ui.Device.SetBrightness( 100 )
				ui.Fresh = true
				ui.SetActivePageID( "default" )
				ui.Clear()
				ui.Render()
			}

			button.Timer = time.AfterFunc( ( time.Millisecond * 500 ) , func() {
				if ui.Fresh {
					ui.Fresh = false
					return
				}
				switch button.PressCount {
					case 1:
						if ui.SettingsMode == false {
							if now.Sub( ui.LastPressTime ).Milliseconds() < ui.GlobalCooldownMilliseconds {
								fmt.Println( "pressed too soon , waiting" )
								break
							}
						}
						if button.SingleClick == "" {
							fmt.Println( "Single Click not Registered" )
							break
						}
						ui.ButtonAction( button , "Single Click" , button.SingleClick , button.MP3 )
						ui.LastPressTime = now
					case 2:
						if ui.SettingsMode == false {
							if now.Sub( ui.LastPressTime ).Milliseconds() < ui.GlobalCooldownMilliseconds {
								fmt.Println( "pressed too soon , waiting" )
								break
							}
						}
						if button.DoubleClick == "" {
							fmt.Println( "Double Click not Registered" )
							if button.SingleClick == "" {
								fmt.Println( "Single Click not Registered" )
							} else {
								fmt.Println( "Rolling Back to Single Click" )
								ui.ButtonAction( button , "Single Click" , button.SingleClick , button.MP3 )
								ui.LastPressTime = now
							}
							break
						}
						ui.ButtonAction( button , "Double Click", button.DoubleClick , button.MP3 )
						ui.LastPressTime = now
					case 3:
						if ui.SettingsMode == false {
							if now.Sub( ui.LastPressTime ).Milliseconds() < ui.GlobalCooldownMilliseconds {
								fmt.Println( "pressed too soon , waiting" )
								break
							}
						}
						if button.TripleClick == "" {
							fmt.Println( "Triple Click not Registered" )
							if button.DoubleClick == "" {
								fmt.Println( "Double Click not Registered" )
								if button.SingleClick == "" {
									fmt.Println( "Single Click not Registered" )
								} else {
									fmt.Println( "Rolling Back to Single Click" )
									ui.ButtonAction( button , "Single Click" , button.SingleClick , button.MP3 )
									ui.LastPressTime = now
								}
								break
							} else {
								fmt.Println( "Rolling Back to Double Click" )
								ui.ButtonAction( button , "Double Click", button.DoubleClick , button.MP3 )
								ui.LastPressTime = now
							}
							break
						}
						ui.ButtonAction( button, "Triple Click", button.TripleClick , button.MP3 )
						ui.LastPressTime = now
				}
				button.PressCount = 0
			})

			if button.Toggle != "" {
				pageID := ui.GetActivePageID()
				for i , xButton := range ui.Pages[ ui.ActivePageID ].Buttons {
					if xButton.Index == key.Index {
						ui.Pages[ pageID ].Buttons[ i ].Id = button.Toggle
						break
					}
				}
				ui.RenderSoft()
			}
			ui.Buttons[ button.Id ] = button
		}
	}
}

func ( ui *StreamDeckUI ) PlayMP3( file_path string ) {
	if ui.Muted { return }
	ui.PlayBackMutex.Lock()
	defer ui.PlayBackMutex.Unlock()
	try.This( func() {
		f , err := os.Open( file_path )
		defer f.Close()
		if err != nil { fmt.Println( err ); return }
		d , err := mp3.NewDecoder( f )
		if err != nil { fmt.Println( err ); return }
		c , err := oto.NewContext( d.SampleRate() , 2 , 2 , 8192 )
		defer c.Close()
		if err != nil { fmt.Println( err ); return }
		p := c.NewPlayer()
		defer p.Close()
		if _ , err := io.Copy( p , d ); err != nil { fmt.Println( err ) }
	}).Catch( func( e try.E ) {
		fmt.Println( e )
	})
}

// doesn't persist
func ( ui *StreamDeckUI ) AddImageAsTiledButton( file_path string , button Button ) string {

	fmt.Println( button )

	// image output prep
	input_file , err := os.Open( file_path )
	if err != nil { fmt.Println( err ); return "" }
	defer input_file.Close()
	img , format , err := image.Decode( input_file )
	if err != nil { fmt.Println( err ); return "" }
	// Extract the file name stem (without extension) for the output directory
	// original_dir , original_file_name := filepath.Split( file_path )
	_ , original_file_name := filepath.Split( file_path )
	file_stem := strings.TrimSuffix( original_file_name , filepath.Ext( file_path ) )
	file_stem = slug.Make( file_stem )
	cwd , _ := os.Getwd()
	output_dir := filepath.Join( cwd , "images" , file_stem )
	err = os.MkdirAll( output_dir , os.ModePerm );
	if err != nil { fmt.Println( err ); return "" }

	tile_size_uint := ui.IconSize
	tile_size_int := int( tile_size_uint )
	x_size_int := ui.XSize
	y_size_int := ui.YSize
	x_size_uint := uint( ui.XSize )
	y_size_uint := uint( ui.YSize )

	// Calculate the new dimensions
	new_width := x_size_uint * tile_size_uint
	new_height := y_size_uint * tile_size_uint

	// Resize the image
	resized_img := resize.Resize( new_width , new_height , img , resize.Lanczos3 )

	var page_buttons []PageButton


	// Iterate over the tiles and save each one in the designated directory
	for y := 0; y < y_size_int; y++ {
		for x := 0; x < x_size_int; x++ {
			// Define the rectangle for the current tile
			rect := image.Rect( x*tile_size_int , y*tile_size_int , (x+1)*tile_size_int , (y+1)*tile_size_int )
			tile := image.NewRGBA( image.Rect( 0 , 0 , tile_size_int , tile_size_int ) )
			draw.Draw( tile , tile.Bounds() , resized_img , rect.Min , draw.Src )

			// Construct the file name based on the position in the grid
			file_name_part := y*x_size_int+x+1
			file_name := fmt.Sprintf( "%d.%s" , file_name_part , format )
			tile_path := filepath.Join( output_dir , file_name )

			btn_id := fmt.Sprintf( "%s-%d" , file_stem , file_name_part )
			btn := button
			btn.Image = tile_path
			// fmt.Println( btn )
			ui.AddButton( btn_id , btn )
			page_btn := PageButton{
				Index: uint8( file_name_part - 1 ) ,
				Id: btn_id ,
			}
			page_buttons = append( page_buttons , page_btn )

			// Save the tile using the original format
			out_file , err := os.Create( tile_path )
			if err != nil { fmt.Println( err ); return "" }
			switch format {
				case "jpeg":
					jpeg.Encode( out_file , tile , nil )
				case "png":
					png.Encode( out_file , tile )
				default:
					fmt.Println( "Unsupported image format:" , format )
			}
			out_file.Close()
		}
	}

	ui.AddPage( file_stem , StreamDeckUIPage{
		Buttons: page_buttons ,
	})
	return file_stem
}

func ( ui *StreamDeckUI ) AddImageAsTiledButtonCustom( file_path string , button Button , x_size_int int , y_size_int int , tile_size_int int ) string {

	fmt.Println( button )

	// image output prep
	input_file , err := os.Open( file_path )
	if err != nil { fmt.Println( err ); return "" }
	defer input_file.Close()
	img , format , err := image.Decode( input_file )
	if err != nil { fmt.Println( err ); return "" }
	// Extract the file name stem (without extension) for the output directory
	// original_dir , original_file_name := filepath.Split( file_path )
	_ , original_file_name := filepath.Split( file_path )
	file_stem := strings.TrimSuffix( original_file_name , filepath.Ext( file_path ) )
	file_stem = slug.Make( file_stem )
	cwd , _ := os.Getwd()
	output_dir := filepath.Join( cwd , "images" , file_stem )
	err = os.MkdirAll( output_dir , os.ModePerm );
	if err != nil { fmt.Println( err ); return "" }

	tile_size_uint := uint( tile_size_int )
	x_size_uint := uint( x_size_int )
	y_size_uint := uint( y_size_int )

	// Calculate the new dimensions
	new_width := x_size_uint * tile_size_uint
	new_height := y_size_uint * tile_size_uint

	// Resize the image
	resized_img := resize.Resize( new_width , new_height , img , resize.Lanczos3 )

	var page_buttons []PageButton


	// Iterate over the tiles and save each one in the designated directory
	for y := 0; y < y_size_int; y++ {
		for x := 0; x < x_size_int; x++ {
			// Define the rectangle for the current tile
			rect := image.Rect( x*tile_size_int , y*tile_size_int , (x+1)*tile_size_int , (y+1)*tile_size_int )
			tile := image.NewRGBA( image.Rect( 0 , 0 , tile_size_int , tile_size_int ) )
			draw.Draw( tile , tile.Bounds() , resized_img , rect.Min , draw.Src )

			// Construct the file name based on the position in the grid
			file_name_part := y*x_size_int+x+1
			file_name := fmt.Sprintf( "%d.%s" , file_name_part , format )
			tile_path := filepath.Join( output_dir , file_name )

			btn_id := fmt.Sprintf( "%s-%d" , file_stem , file_name_part )
			btn := button
			btn.Image = tile_path
			// fmt.Println( btn )
			ui.AddButton( btn_id , btn )
			page_btn := PageButton{
				Index: uint8( file_name_part - 1 ) ,
				Id: btn_id ,
			}
			page_buttons = append( page_buttons , page_btn )

			// Save the tile using the original format
			out_file , err := os.Create( tile_path )
			if err != nil { fmt.Println( err ); return "" }
			switch format {
				case "jpeg":
					jpeg.Encode( out_file , tile , nil )
				case "png":
					png.Encode( out_file , tile )
				default:
					fmt.Println( "Unsupported image format:" , format )
			}
			out_file.Close()
		}
	}

	ui.AddPage( file_stem , StreamDeckUIPage{
		Buttons: page_buttons ,
	})
	return file_stem
}

func NewStreamDeckUI( file_path string ) ( result *StreamDeckUI ) {
	file , _ := ioutil.ReadFile( file_path )
	error := yaml.Unmarshal( file , &result )
	if error != nil { panic( error ) }
	result.LoadedButtonImages = make(map[uint8]string)
	result.Brightness = 100
	return
}

func NewStreamDeckUIFromInterface( config interface{} ) ( result *StreamDeckUI ) {
	intermediate , _ := yaml.Marshal( config )
	error := yaml.Unmarshal( intermediate , &result )
	result.LoadedButtonImages = make(map[uint8]string)
	result.Brightness = 100
	if error != nil { panic( error ) }
	return
}