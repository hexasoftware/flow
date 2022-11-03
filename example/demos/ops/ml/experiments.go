package ml

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"log"
	"math"

	"github.com/hexasoftware/flow/flowserver"
	"gonum.org/v1/gonum/mat"
)

// imageToMat create a grayscaled matrix of the image
func imageToGrayMatrix(im image.Image) (mat.Matrix, error) {
	dims := im.Bounds().Size()
	fdata := make([]float64, dims.X*dims.Y)

	for y := 0; y < dims.Y; y++ {
		for x := 0; x < dims.X; x++ {
			oldPixel := im.At(x, y)
			r, g, b, _ := oldPixel.RGBA()
			lum := (19595*r + 38470*g + 7471*b + 1<<15) >> 24
			fdata[x+y*dims.X] = float64(lum) / 256
		}
	}
	m := mat.NewDense(dims.Y, dims.X, fdata)
	return m, nil
}

func displayImg(img image.Image) (flowserver.Base64Data, error) {
	pngEncoded := bytes.NewBuffer(nil)
	err := png.Encode(pngEncoded, img)
	if err != nil {
		return flowserver.Base64Data([]byte{}), err
	}
	base64enc := base64.StdEncoding.EncodeToString(pngEncoded.Bytes())
	out := bytes.NewBuffer(nil)
	out.WriteString("data:image/png;base64,")
	out.WriteString(base64enc)

	return flowserver.Base64Data(out.String()), nil

}
func displayGrayMat(m mat.Matrix) (flowserver.Base64Data, error) {
	r, c := m.Dims()

	img := image.NewGray(image.Rect(0, 0, c, r))
	for y := 0; y < r; y++ {
		for x := 0; x < c; x++ {
			img.Pix[x+y*c] = byte(m.At(y, x) * 255)
		}
	}
	pngEncoded := bytes.NewBuffer(nil)

	err := png.Encode(pngEncoded, img)
	if err != nil {
		return flowserver.Base64Data([]byte{}), err
	}
	base64enc := base64.StdEncoding.EncodeToString(pngEncoded.Bytes())
	out := bytes.NewBuffer(nil)
	out.WriteString("data:image/png;base64,")
	out.WriteString(base64enc)

	return flowserver.Base64Data(out.String()), nil

}

// Test
func toGrayImage(data []byte, w, h int) (flowserver.Base64Data, error) {
	// Convert matrix to byte 0-255

	/*bdata := make([]byte, w*h)
	for i, v := range data {
		bdata[i] = byte(v * 255)
	}*/

	img := image.NewGray(image.Rect(0, 0, w, h))
	for i, v := range data {
		img.Pix[i] = v
	}
	pngEncoded := bytes.NewBuffer(nil)

	err := png.Encode(pngEncoded, img)
	if err != nil {
		return flowserver.Base64Data([]byte{}), err
	}

	base64enc := base64.StdEncoding.EncodeToString(pngEncoded.Bytes())

	out := bytes.NewBuffer(nil)
	out.WriteString("data:image/png;base64,")
	out.WriteString(base64enc)

	return flowserver.Base64Data(out.String()), nil

	/*for y:=0;y<h;y++ {
		for x:=0;x<w;x++ {
			bdata[x + y *w] = data[
		}
	}*/
}

// Convolution matrix
func matConv(a mat.Matrix, conv mat.Matrix) (mat.Matrix, error) {
	convR, convC := conv.Dims()

	if convR&1 == 0 || convC&1 == 0 {
		return nil, errors.New("kernel matrix should have odd columns and odd rows")
	}
	midR := int(math.Floor(float64(convR) / 2))
	midC := int(math.Floor(float64(convC) / 2))
	log.Println("Middle:", midR, midC)

	norm := float64(0)
	for cy := 0; cy < convR; cy++ {
		for cx := 0; cx < convC; cx++ {
			norm += conv.At(cy, cx)
		}
	}
	if norm == 0 {
		norm = 1.0
	}

	rows, cols := a.Dims()
	ret := mat.NewDense(rows, cols, nil)
	for y := 0; y < rows; y++ { // Matrix loop
		for x := 0; x < cols; x++ {
			acc := float64(0) //accumulator
			for cy := 0; cy < convR; cy++ {
				matY := y + cy - midR
				if matY < 0 || matY >= rows {
					continue
				}
				for cx := 0; cx < convC; cx++ {
					matX := x + (cx - midC)
					if matX < 0 || matX >= cols {
						continue
					}
					acc += a.At(matY, matX) * (conv.At(cy, cx) / norm)
				}
			}
			//acc /= 9
			if acc > 1.0 {
				acc = 1
			}
			if acc < 0.0 {
				acc = 0.0
			}

			ret.Set(y, x, acc)
		}
	}
	return ret, nil

}
