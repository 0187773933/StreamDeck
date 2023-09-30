package main

import (
	"fmt"
	"encoding/json"
	"strings"

	"sync"
	"time"
	"context"

	"image"
	"image/color"
	"image/jpeg"
	"bytes"
	"golang.org/x/image/draw"

	hid "github.com/dh1tw/hid"
	// streamdeck_wrapper "github.com/muesli/streamdeck"
)

type Device struct {
	ID     string
	Serial string

	Columns uint8
	Rows    uint8
	Keys    uint8
	Pixels  uint
	DPI     uint
	Padding uint

	featureReportSize   int
	firmwareOffset      int
	keyStateOffset      int
	translateKeyIndex   func(index, columns uint8) uint8
	imagePageSize       int
	imagePageHeaderSize int
	flipImage           func(image.Image) image.Image
	toImageFormat       func(image.Image) ([]byte, error)
	imagePageHeader     func(pageIndex int, keyIndex uint8, payloadLength int, lastPage bool) []byte

	getFirmwareCommand   []byte
	resetCommand         []byte
	setBrightnessCommand []byte

	keyState []byte

	device *hid.Device
	info   hid.DeviceInfo

	lastActionTime time.Time
	asleep         bool
	sleepCancel    context.CancelFunc
	sleepMutex     *sync.RWMutex
	fadeDuration   time.Duration

	brightness         uint8
	preSleepBrightness uint8
}

func (d *Device) Open() error {
	var err error
	d.device, err = d.info.Open()
	d.lastActionTime = time.Now()
	d.sleepMutex = &sync.RWMutex{}
	return err
}

func (d Device) Clear() error {
	img := image.NewRGBA(image.Rect(0, 0, int(d.Pixels), int(d.Pixels)))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	for i := uint8(0); i <= d.Columns*d.Rows; i++ {
		err := d.SetImage(i, img)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

// imageData allows to access raw image data in a byte array through pages of a
// given size.
type imageData struct {
	image    []byte
	pageSize int
}

// Page returns the page with the given index and an indication if this is the
// last page.
func (d imageData) Page(pageIndex int) ([]byte, bool) {
	offset := pageIndex * d.pageSize
	if offset >= len(d.image) {
		return []byte{}, true
	}

	length := d.pageLength(pageIndex)
	if offset+length > len(d.image) {
		length = len(d.image) - offset
	}

	return d.image[offset : offset+length], pageIndex == d.PageCount()-1
}

func (d imageData) pageLength(pageIndex int) int {
	remaining := len(d.image) - (pageIndex * d.pageSize)
	if remaining > d.pageSize {
		return d.pageSize
	}
	if remaining > 0 {
		return remaining
	}
	return 0
}

// PageCount returns the total number of pages.
func (d imageData) PageCount() int {
	count := len(d.image) / d.pageSize
	if len(d.image)%d.pageSize != 0 {
		return count + 1
	}
	return count
}

// Length of the raw image data in bytes.
func (d imageData) Length() int {
	return len(d.image)
}

// toJPEG returns the raw bytes of the given image in JPEG format.
func toJPEG(img image.Image) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	opts := jpeg.Options{
		Quality: 100,
	}
	err := jpeg.Encode(buffer, img, &opts)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}

func (d Device) SetImage(index uint8, img image.Image) error {
	if img.Bounds().Dy() != int(d.Pixels) ||
		img.Bounds().Dx() != int(d.Pixels) {
		return fmt.Errorf("supplied image has wrong dimensions, expected %[1]dx%[1]d pixels", d.Pixels)
	}

	imageBytes, err := toJPEG(d.flipImage(img))
	if err != nil {
		return fmt.Errorf("cannot convert image data: %v", err)
	}
	imageData := imageData{
		image:    imageBytes,
		pageSize: d.imagePageSize - d.imagePageHeaderSize,
	}

	data := make([]byte, d.imagePageSize)

	var page int
	var lastPage bool
	for !lastPage {
		var payload []byte
		payload, lastPage = imageData.Page(page)
		header := d.imagePageHeader(page, d.translateKeyIndex(index, d.Columns), len(payload), lastPage)

		copy(data, header)
		copy(data[len(header):], payload)

		_, err := d.device.Write(data)
		if err != nil {
			return fmt.Errorf("cannot write image page %d of %d (%d image bytes) %d bytes: %v",
				page, imageData.PageCount(), imageData.Length(), len(data), err)
		}

		page++
	}

	return nil
}

type HIDDevice struct {
	Name      string
	VendorID  uint16
	ProductID uint16
	Serial string
	Path string
}
func GetAllHIDDevices() ( result []HIDDevice ) {
	devices := hid.Enumerate( 0 , 0 )
	devices_map := make( map[ HIDDevice ]bool )
	for _ , device := range devices {
		// fmt.Println( i , device )
		// fmt.Printf( "%d === %s %s === %d === %d\n" , i , device.Manufacturer , device.Product , device.VendorID , device.ProductID )
		name := strings.TrimSpace( device.Manufacturer + " " + device.Product )
		if name == "" { continue }
		d := HIDDevice{
			Name: name ,
			VendorID: device.VendorID ,
			ProductID: device.ProductID ,
			Serial: device.Serial ,
			Path: device.Path ,
		}
		devices_map[ d ] = true
	}
	result = make( []HIDDevice , 0 , len( devices_map ) )
	for device := range devices_map { result = append( result , device ) }
	return
}

func PrintDevices( devices []HIDDevice ) {
	jd , _ := json.MarshalIndent( devices , "" , "  " )
	fmt.Println( string( jd ) )
}

func identity(index, _ uint8) uint8 {
	return index
}

func miniImagePageHeader(pageIndex int, keyIndex uint8, payloadLength int, lastPage bool) []byte {
	var lastPageByte byte
	if lastPage {
		lastPageByte = 1
	}
	return []byte{
		0x02, 0x01,
		byte(pageIndex), 0x00,
		lastPageByte,
		keyIndex + 1,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func rotateCounterclockwise(img image.Image) image.Image {
	flipped := image.NewRGBA(img.Bounds())
	draw.Copy(flipped, image.Point{}, img, img.Bounds(), draw.Src, nil)
	for y := 0; y < flipped.Bounds().Dy(); y++ {
		for x := y + 1; x < flipped.Bounds().Dx(); x++ {
			c := flipped.RGBAAt(x, y)
			flipped.SetRGBA(x, y, flipped.RGBAAt(y, x))
			flipped.SetRGBA(y, x, c)
		}
	}
	for y := 0; y < flipped.Bounds().Dy()/2; y++ {
		yy := flipped.Bounds().Max.Y - y - 1
		for x := 0; x < flipped.Bounds().Dx(); x++ {
			c := flipped.RGBAAt(x, y)
			flipped.SetRGBA(x, y, flipped.RGBAAt(x, yy))
			flipped.SetRGBA(x, yy, c)
		}
	}
	return flipped
}

func toRGBA(img image.Image) *image.RGBA {
	switch img := img.(type) {
	case *image.RGBA:
		return img
	}
	out := image.NewRGBA(img.Bounds())
	draw.Copy(out, image.Pt(0, 0), img, img.Bounds(), draw.Src, nil)
	return out
}


func toBMP(img image.Image) ([]byte, error) {
	rgba := toRGBA(img)

	// this is a BMP file header followed by a BPM bitmap info header
	// find more information here: https://en.wikipedia.org/wiki/BMP_file_format
	header := []byte{
		0x42, 0x4d, 0xf6, 0x3c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x36, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x48, 0x00, 0x00, 0x00, 0x48, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x18, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xc0, 0x3c, 0x00, 0x00, 0xc4, 0x0e,
		0x00, 0x00, 0xc4, 0x0e, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	buffer := make([]byte, len(header)+rgba.Bounds().Dx()*rgba.Bounds().Dy()*3)
	copy(buffer, header)

	i := len(header)
	for y := 0; y < rgba.Bounds().Dy(); y++ {
		for x := 0; x < rgba.Bounds().Dx(); x++ {
			c := rgba.RGBAAt(x, y)
			buffer[i] = c.B
			buffer[i+1] = c.G
			buffer[i+2] = c.R
			i += 3
		}
	}
	return buffer, nil
}

var (
	c_REV1_FIRMWARE   = []byte{0x04}
	c_REV1_RESET      = []byte{0x0b, 0x63}
	c_REV1_BRIGHTNESS = []byte{0x05, 0x55, 0xaa, 0xd1, 0x01}

	// c_REV2_FIRMWARE   = []byte{0x05}
	// c_REV2_RESET      = []byte{0x03, 0x02}
	// c_REV2_BRIGHTNESS = []byte{0x03, 0x08}
)


// https://github.com/muesli/streamdeck/blob/v0.4.0/streamdeck.go#L128
func main() {
	// hid_devices := GetAllHIDDevices()
	hid_devices := GetAllHIDDevices()
	PrintDevices( hid_devices )

	dev := Device{
		ID:                   "DevSrvsID:4296782515",
		Serial:               "AL02K2C02319",
		Columns:              3,
		Rows:                 2,
		Keys:                 6,
		Pixels:               80,
		DPI:                  138,
		Padding:              16,
		featureReportSize:    17,
		firmwareOffset:       5,
		keyStateOffset:       1,
		translateKeyIndex:    identity,
		imagePageSize:        1024,
		imagePageHeaderSize:  16,
		imagePageHeader:      miniImagePageHeader,
		flipImage:            rotateCounterclockwise,
		toImageFormat:        toBMP,
		getFirmwareCommand:   c_REV1_FIRMWARE,
		resetCommand:         c_REV1_RESET,
		setBrightnessCommand: c_REV1_BRIGHTNESS,
	}

	fmt.Println( dev )
	dev.Open()
	dev.Clear()
}