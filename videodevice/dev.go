package main

import (
	"fmt"
	"image"

	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/yuyv"
	"github.com/lugu/qiloop/bus"
	"github.com/lugu/qiloop/type/value"
)

// videoDev implements ALVideoDeviceImplementor
type videoDev struct {
	cfg v4l.DeviceConfig
	cam *v4l.Device
	img yuyv.Image
}

func NewVideoDevice() (bus.Actor, error) {
	devs := v4l.FindDevices()
	if len(devs) != 1 {
		return nil, fmt.Errorf("no device found")
	}
	d := devs[0].Path

	cam, err := v4l.Open(d)
	if err != nil {
		return nil, fmt.Errorf("open %s: %s", d, err)
	}

	cfg, err := cam.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config %s: %s", d, err)
	}

	cfg.Format = yuyv.FourCC

	err = cam.SetConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("set config %s: %s", d, err)
	}

	cfg, err = cam.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config %s: %s", d, err)
	}

	if cfg.Format != yuyv.FourCC {
		return nil, fmt.Errorf("invalid color format")
	}

	info, err := cam.BufferInfo()
	if err != nil {
		return nil, fmt.Errorf("invalid color format: %s", err)
	}

	img := yuyv.Image{
		Pix:    make([]byte, info.BufferSize),
		Stride: info.ImageStride,
		Rect:   image.Rect(0, 0, cfg.Width, cfg.Height),
	}

	return ALVideoDeviceObject(&videoDev{
		cfg: cfg,
		cam: cam,
		img: img,
	}), nil
}

func (d *videoDev) Activate(activation bus.Activation,
	helper ALVideoDeviceSignalHelper) error {
	return nil
}

func (d *videoDev) OnTerminate() {
}

func (d *videoDev) SubscribeCamera(name string, cameraIndex int32,
	resolution int32, colorSpace int32, fps int32) (string, error) {

	err := d.cam.TurnOn()
	if err != nil {
		return "", fmt.Errorf("read config %v: %s", d, err)
	}

	return "singleton", nil
}

func (d *videoDev) GetImageRemote(name string) (value.Value, error) {

	buf, err := d.cam.Capture()
	if err != nil {
		return value.Void(), fmt.Errorf("Capture: %w", err)
	}

	width, height := d.cfg.Width, d.cfg.Height

	buf.ReadAt(d.img.Pix, 0)

	rgba := &image.RGBA{
		Pix:    make([]byte, width*height*4),
		Stride: width * 4,
		Rect:   d.img.Rect,
	}

	yuyv.ToRGBA(rgba, rgba.Rect, &d.img, rgba.Rect.Min)

	pixels := make([]byte, width*height*3)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			col := rgba.At(x, y)
			r, g, b, _ := col.RGBA()
			pixels[y*width*3+x*3] = byte(r)
			pixels[y*width*3+x*3+1] = byte(g)
			pixels[y*width*3+x*3+2] = byte(b)
		}
	}

	values := make([]value.Value, 7)
	values[0] = value.Int(int32(width))  // width
	values[1] = value.Int(int32(height)) // height
	values[2] = value.Void()
	values[3] = value.Void()
	values[4] = value.Void()
	values[5] = value.Void()
	values[6] = value.Raw(pixels) // pixels
	return value.List(values), nil

}

func (d *videoDev) Unsubscribe(nameId string) (bool, error) {
	d.cam.TurnOff()
	return true, nil
}
