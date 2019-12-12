package face

import (
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"io/ioutil"
	"log"

	pigo "github.com/esimov/pigo/core"
	"github.com/markbates/pkger"
)

var (
	classifier *pigo.Pigo = nil
)

func loadClassifier() {
	model, err := pkger.Open("/face/facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %v", err)
	}
	cascadeFile, err := ioutil.ReadAll(model)
	if err != nil {
		log.Fatalf("Error reading the cascade file: %v", err)
	}
	p := pigo.NewPigo()
	classifier, err = p.Unpack(cascadeFile)
	if err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
	}
}

// Draw detects faces draw rectangles around them.
func Draw(img draw.Image) {

	if classifier == nil {
		loadClassifier()
	}

	src := pigo.ImgToNRGBA(img)
	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y
	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	angle := 0.0
	dets := classifier.RunCascade(cParams, angle)

	// Calculate the intersection over union (IoU) of two clusters.
	iouThreshold := 0.2
	faces := classifier.ClusterDetections(dets, iouThreshold)

	drawRectangles(img, faces)
}

// Draw a rectangle for each detection.
func drawRectangles(img draw.Image, faces []pigo.Detection) {
	col := color.RGBA{0, 255, 0, 255} // Green

	// hLine draws a horizontal line
	hLine := func(x1, y, x2 int) {
		for ; x1 <= x2; x1++ {
			img.Set(x1, y, col)
		}
	}

	// vLine draws a veritcal line
	vLine := func(x, y1, y2 int) {
		for ; y1 <= y2; y1++ {
			img.Set(x, y1, col)
		}
	}

	// rect draws a rectangle utilizing HLine() and VLine()
	rect := func(x1, y1, x2, y2 int) {
		hLine(x1, y1, x2)
		hLine(x1, y2, x2)
		vLine(x1, y1, y2)
		vLine(x2, y1, y2)
	}

	var qThresh float32 = 5.0

	for _, face := range faces {
		if face.Q <= qThresh {
			continue
		}
		rect(
			face.Col-face.Scale/2,
			face.Row-face.Scale/2,
			face.Col-face.Scale/2+face.Scale,
			face.Row-face.Scale/2+face.Scale,
		)
	}
}
