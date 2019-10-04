// Package main is a terminal viewer for video device.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/lugu/qiloop/app"
	"github.com/lugu/qiloop/type/value"

	"image"
	"image/color"

	_ "image/jpeg"
	_ "image/png"

	"github.com/qeesung/image2ascii/convert"
)

var (
	id = "ascii" // video device subscriber id
)

type imageRGB struct {
	pixels       []byte
	width, heigh int
}

func (i *imageRGB) ColorModel() color.Model {
	return color.RGBAModel
}
func (i *imageRGB) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: 0, Y: 0,
		},
		Max: image.Point{
			X: int(i.width), Y: int(i.heigh),
		},
	}
}
func (i *imageRGB) At(x, y int) color.Color {
	var c color.RGBA
	c.R = i.pixels[3*y*i.width+3*x]
	c.G = i.pixels[3*y*i.width+3*x+1]
	c.B = i.pixels[3*y*i.width+3*x+2]
	c.A = 0xff
	return c
}

const (
	topCam    = 0
	bottomCam = 1
	depthCam  = 2
	stereoCam = 3

	qvga = 1
	vga  = 2
	vga4 = 3

	yuv  = 10
	rgb  = 11
	hsv  = 12
	dist = 21

	fps = 10
)

func printImage(img value.Value) {

	// GetImageRemote returns an value, let's cast it into a list
	// of values:
	values, ok := img.(value.ListValue)
	if !ok {
		log.Fatalf("invalid return type: %#v", img)
	}
	var image imageRGB
	// Let's extract the image data.
	image.width = int(values[0].(value.IntValue).Value())
	image.heigh = int(values[1].(value.IntValue).Value())
	image.pixels = values[6].(value.RawValue).Value()

	//log.Printf("camera resolution: %dx%d\n", image.width, image.heigh)

	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = 100
	convertOptions.FixedHeight = 40

	// Create the image converter
	converter := convert.NewImageConverter()
	fmt.Print(converter.Image2ASCIIString(&image, &convertOptions))
}

func main() {
	var cameraName string
	flag.StringVar(&cameraName, "camera", "top", "possible values: top, bottom, depth, stereo")

	flag.Parse()

	var camera int32 = topCam
	switch cameraName {
	case "top":
		camera = topCam
	case "bottom":
		camera = bottomCam
	case "depth":
		camera = depthCam
	case "stereo":
		camera = stereoCam
	default:
		log.Fatal("invalid camera argument")
	}

	// A Session object is used to connect the service directory.
	sess, err := app.SessionFromFlag()
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}

	// Using this session, let's instanciate our service
	// constructor.
	services := Services(sess)

	// Using the constructor, we request a proxy to ALVideoDevice
	videoDevice, err := services.ALVideoDevice()
	if err != nil {
		log.Fatalf("failed to create video device: %s", err)
	}

	// Configure the camera
	id, err := videoDevice.SubscribeCamera(id, camera, qvga, rgb, fps)
	if err != nil {
		log.Fatalf("failed to initialize camera: %s", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	timer := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case s := <-interrupt:
			log.Printf("%v: quitting.", s)
			ok, err := videoDevice.Unsubscribe(id)
			if !ok || err != nil {
				log.Fatalf("failed to unsubscribe: %s", err)
			}
			return
		case <-timer.C:
			img, err := videoDevice.GetImageRemote(id)
			if err != nil {
				log.Fatalf("failed to retrieve image: %s", err)
			}
			printImage(img)
		}
	}

}
