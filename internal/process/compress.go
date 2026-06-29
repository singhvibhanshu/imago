package process

import (
	"image"

	"github.com/singhvibhanshu/imago/internal/imageio"
)

// CompressResult reports what the compressor actually produced.
type CompressResult struct {
	Data    []byte
	Quality int     // quality used (lossy formats only; 0 otherwise)
	Scale   float64 // 1.0 means original dimensions were kept
	Width   int
	Height  int
	Met     bool // true if the output is within the requested size budget
}

// CompressToQuality is the simple path: encode once at a fixed quality.
func CompressToQuality(img image.Image, format string, quality int) ([]byte, error) {
	return imageio.Encode(img, format, imageio.EncodeOptions{Quality: quality})
}

// CompressToTarget produces an encoding that fits within maxBytes.
//
// Strategy:
//  1. For lossy formats (JPEG), binary-search the highest quality that fits.
//  2. If even the lowest quality is too large — or the format is lossless —
//     progressively shrink the dimensions and try again.
//
// This two-stage approach is what makes "get this photo under 50 KB at roughly
// these dimensions" reliable, which is exactly what exam-form uploads demand.
func CompressToTarget(img image.Image, format string, maxBytes int64) (CompressResult, error) {
	format = imageio.NormalizeFormat(format)
	lossy := format == imageio.FormatJPEG

	var best CompressResult // smallest output seen, used as best-effort fallback
	best.Scale = 1.0

	scale := 1.0
	for i := 0; i < 16; i++ {
		cur := img
		if scale < 1.0 {
			cur = ScaleByFactor(img, scale)
		}
		b := cur.Bounds()
		w, h := b.Dx(), b.Dy()

		if lossy {
			if data, q, ok := searchQuality(cur, format, maxBytes); ok {
				return CompressResult{Data: data, Quality: q, Scale: scale, Width: w, Height: h, Met: true}, nil
			}
			// Floor quality still too big: remember it, then shrink.
			if data, err := imageio.Encode(cur, format, imageio.EncodeOptions{Quality: 1}); err == nil {
				best = keepSmaller(best, CompressResult{Data: data, Quality: 1, Scale: scale, Width: w, Height: h})
			}
		} else {
			data, err := imageio.Encode(cur, format, imageio.EncodeOptions{})
			if err != nil {
				return CompressResult{}, err
			}
			if int64(len(data)) <= maxBytes {
				return CompressResult{Data: data, Scale: scale, Width: w, Height: h, Met: true}, nil
			}
			best = keepSmaller(best, CompressResult{Data: data, Scale: scale, Width: w, Height: h})
		}

		if w <= 16 || h <= 16 {
			break // don't shrink into oblivion
		}
		scale *= 0.8
	}

	// Couldn't hit the target; hand back the closest we managed, flagged.
	best.Met = false
	return best, nil
}

// searchQuality finds the highest JPEG/WebP quality whose output fits maxBytes.
func searchQuality(img image.Image, format string, maxBytes int64) ([]byte, int, bool) {
	lo, hi := 1, 100
	var best []byte
	bestQ := -1
	for lo <= hi {
		mid := (lo + hi) / 2
		data, err := imageio.Encode(img, format, imageio.EncodeOptions{Quality: mid})
		if err != nil {
			return nil, 0, false
		}
		if int64(len(data)) <= maxBytes {
			best, bestQ = data, mid
			lo = mid + 1 // fits — try for better quality
		} else {
			hi = mid - 1 // too big — lower quality
		}
	}
	if bestQ >= 0 {
		return best, bestQ, true
	}
	return nil, 0, false
}

func keepSmaller(a, b CompressResult) CompressResult {
	if a.Data == nil {
		return b
	}
	if len(b.Data) < len(a.Data) {
		return b
	}
	return a
}
