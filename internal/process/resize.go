// Package process holds the image transformation logic: resizing and the
// size-targeting compressor.
package process

import (
	"image"

	"golang.org/x/image/draw"
)

// ResizeParams describes a desired resize. Any field left zero is inferred.
type ResizeParams struct {
	Width      int     // target width in pixels (0 = auto)
	Height     int     // target height in pixels (0 = auto)
	Percent    float64 // scale by percentage, e.g. 50 = half (overrides W/H)
	KeepAspect bool    // preserve aspect ratio when only constraining one side
}

// TargetDimensions computes the final pixel dimensions for a source image given
// the requested parameters, without doing the actual scaling.
func TargetDimensions(src image.Rectangle, p ResizeParams) (int, int) {
	sw, sh := src.Dx(), src.Dy()
	if sw == 0 || sh == 0 {
		return sw, sh
	}

	if p.Percent > 0 {
		w := int(float64(sw) * p.Percent / 100.0)
		h := int(float64(sh) * p.Percent / 100.0)
		return atLeastOne(w), atLeastOne(h)
	}

	w, h := p.Width, p.Height

	switch {
	case w > 0 && h > 0:
		if p.KeepAspect {
			// Fit within the box without distortion.
			ratio := float64(sw) / float64(sh)
			if float64(w)/float64(h) > ratio {
				w = int(float64(h) * ratio)
			} else {
				h = int(float64(w) / ratio)
			}
		}
	case w > 0 && h == 0:
		h = int(float64(w) * float64(sh) / float64(sw))
	case h > 0 && w == 0:
		w = int(float64(h) * float64(sw) / float64(sh))
	default:
		w, h = sw, sh // nothing requested
	}

	return atLeastOne(w), atLeastOne(h)
}

// Resize returns a new image scaled to the requested dimensions using a
// high-quality (Catmull-Rom) kernel.
func Resize(img image.Image, p ResizeParams) image.Image {
	w, h := TargetDimensions(img.Bounds(), p)
	if w == img.Bounds().Dx() && h == img.Bounds().Dy() {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// ScaleByFactor returns the image scaled by a factor (0 < f <= 1 typically),
// used by the size-targeting compressor when it needs to shrink dimensions.
func ScaleByFactor(img image.Image, f float64) image.Image {
	b := img.Bounds()
	w := atLeastOne(int(float64(b.Dx()) * f))
	h := atLeastOne(int(float64(b.Dy()) * f))
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}

func atLeastOne(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
