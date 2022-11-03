package decodeops

import (
	"bytes"
	"image"
	// Image decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/hexasoftware/flow/registry"
)

//New decoding ops
func New() *registry.R {
	r := registry.New()
	r.Add(DecodeImage).Tags("experiment-decode")
	return r
}

// DecodeImage from a byte array
func DecodeImage(in []byte) (image.Image, error) {
	br := bytes.NewReader(in)
	im, _, err := image.Decode(br)
	return im, err
}

/*func decodePNG(in []byte) (image.Image, error) {
	br := bytes.NewReader(in)
	im, err := png.Decode(br)
	if err != nil {
		return nil, err
	}
	return im, nil
}

func decodeJPG(in []byte) (image.Image, error) {
	br := bytes.NewReader(in)
	im, err := jpeg.Decode(br)
	if err != nil {
		return nil, err
	}

	return im, nil
}*/
