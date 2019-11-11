// Package main is a terminal viewer for video device.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/lugu/qiloop/app"
	"github.com/lugu/qiloop/type/value"
	tb "github.com/nsf/termbox-go"
)

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

	screenWidth  = 640
	screenHeight = 480
)

var (
	id          = "ascii" // video device subscriber id
	fps         = 15
	cameraName  = "top"
	videoDevice ALVideoDeviceProxy

	errQuit    = errors.New("Quitting...")
	firstFrame = true
)

func getImage() (image.Image, error) {

	img, err := videoDevice.GetImageRemote(id)
	if err != nil {
		return nil, fmt.Errorf("GetImageRemote: %s", err)
	}

	// GetImageRemote returns an value, let's cast it into a list
	// of values:
	values, ok := img.(value.ListValue)
	if !ok {
		return nil, fmt.Errorf("Invalid type (not a list): %#v", img)
	}
	image := &imageRGB{
		width:  int(values[0].(value.IntValue).Value()),
		heigh:  int(values[1].(value.IntValue).Value()),
		pixels: values[6].(value.RawValue).Value(),
	}
	return image, nil
}

func ascii() error {
	if err := tb.Init(); err != nil {
		return err
	}

	tb.SetInputMode(tb.InputEsc)
	tb.SetOutputMode(tb.Output256)

	go func() {
		for {
			tb.Interrupt()
			time.Sleep((1000 / time.Duration(fps)) * time.Millisecond)
		}
	}()

	for {
		e := tb.PollEvent()
		switch e.Type {
		case tb.EventInterrupt, tb.EventResize:
			image, err := getImage()
			if err != nil {
				tb.Close()
				return err
			}

			view := NewView(image)
			view.Print()
		case tb.EventKey:
			if e.Key == tb.KeyCtrlC || e.Ch == 'q' || e.Key == tb.KeyEsc {
				tb.Close()
				return nil
			}
		}

	}
}

func update(screen *ebiten.Image) error {

	fullscreen := ebiten.IsFullscreen()
	cursorVisible := ebiten.IsCursorVisible()

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return errQuit
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		fullscreen = !fullscreen
		cursorVisible = !cursorVisible
	}

	ebiten.SetFullscreen(fullscreen)
	ebiten.SetCursorVisible(cursorVisible)

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	if firstFrame {
		screen.Fill(color.RGBA{0xeb, 0xeb, 0xeb, 0xff})
		firstFrame = false
		return nil
	}

	image, err := getImage()
	if err != nil {
		return err
	}

	// screenWidth, screenHeight := screen.Size()
	size := image.Bounds().Size()
	ebiten.SetScreenSize(size.X, size.Y)

	op := &ebiten.DrawImageOptions{}
	img, err := ebiten.NewImageFromImage(image, ebiten.FilterDefault)
	if err != nil {
		return err
	}
	screen.DrawImage(img, op)

	return nil
}

func gui() error {

	ebiten.SetRunnableInBackground(true)
	ebiten.SetMaxTPS(fps)

	// TODO: resize not yet supported
	// call ebiten.SetWindowResizable(true)

	err := ebiten.Run(update, screenWidth, screenHeight, 1.0, "QiView")
	if err != nil && err != errQuit {
		return err
	}
	return nil
}

func main() {
	var is_ascii bool = false
	flag.StringVar(&cameraName, "camera", cameraName, "possible values: top, bottom, depth, stereo")
	flag.IntVar(&fps, "fps", fps, "framerate")
	flag.BoolVar(&is_ascii, "ascii", is_ascii, "ascii mode")

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

	sess, err := app.SessionFromFlag()
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}

	videoDevice, err = Services(sess).ALVideoDevice()
	if err != nil {
		log.Fatalf("failed to create video device: %s", err)
	}

	id, err := videoDevice.SubscribeCamera(id, camera, qvga, rgb, int32(fps))
	if err != nil {
		videoDevice.Unsubscribe(id)
		log.Fatalf("failed to initialize camera: %s", err)
	}
	defer videoDevice.Unsubscribe(id)

	if is_ascii {
		err = ascii()
	} else {
		err = gui()
	}
	if err != nil {
		videoDevice.Unsubscribe(id)
		log.Fatal(err)
	}
}
