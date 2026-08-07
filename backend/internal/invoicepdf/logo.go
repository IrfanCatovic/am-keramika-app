package invoicepdf

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

//go:embed assets/logo-stampa-racuni.svg
var logoSVG []byte

const logoRasterWidth = 400

var (
	logoPNG     []byte
	logoPNGOnce sync.Once
)

func logoPNGBytes() []byte {
	logoPNGOnce.Do(func() {
		pngBytes, err := rasterizeSVG(logoSVG, logoRasterWidth)
		if err == nil {
			logoPNG = pngBytes
		}
	})
	return logoPNG
}

func rasterizeSVG(svg []byte, targetWidth int) ([]byte, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svg))
	if err != nil {
		return nil, err
	}

	vbW := icon.ViewBox.W
	vbH := icon.ViewBox.H
	if vbW <= 0 || vbH <= 0 {
		vbW, vbH = 1618, 848
	}

	width := targetWidth
	height := int(float64(width) * vbH / vbW)
	if height < 1 {
		height = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
	raster := rasterx.NewDasher(width, height, scanner)
	icon.SetTarget(0, 0, float64(width), float64(height))
	icon.Draw(raster, 1.0)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
