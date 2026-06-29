package imageio

import (
	"bytes"
	"image"

	"github.com/HugoSmits86/nativewebp"
)

// encodeWebP writes a WebP image using a pure-Go encoder, so the final binary
// stays free of cgo and works without any system libraries installed.
//
// Note: this encoder is lossless, so the quality argument is currently ignored.
// Size targeting for WebP therefore relies on dimension scaling (see process).
func encodeWebP(img image.Image, _ int) ([]byte, error) {
	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
