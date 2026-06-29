// Package imageio handles loading and saving images, format detection,
// and a few small helpers shared across commands. Everything here is local:
// no image ever leaves the machine.
package imageio

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Blank imports register decoders with image.Decode so we can read
	// these formats without caring which one a file actually is.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/bmp"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp" // decode only; encoding handled in encode_webp.go
)

// Supported output formats.
const (
	FormatJPEG = "jpeg"
	FormatPNG  = "png"
	FormatGIF  = "gif"
	FormatBMP  = "bmp"
	FormatTIFF = "tiff"
	FormatWEBP = "webp"
)

// NormalizeFormat maps user-facing aliases (jpg, JPEG, .png) to a canonical
// format string used everywhere else.
func NormalizeFormat(s string) string {
	s = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), "."))
	switch s {
	case "jpg", "jpeg", "jpe":
		return FormatJPEG
	case "png":
		return FormatPNG
	case "gif":
		return FormatGIF
	case "bmp":
		return FormatBMP
	case "tif", "tiff":
		return FormatTIFF
	case "webp":
		return FormatWEBP
	default:
		return s
	}
}

// Extension returns the conventional file extension for a canonical format.
func Extension(format string) string {
	switch NormalizeFormat(format) {
	case FormatJPEG:
		return ".jpg"
	case FormatPNG:
		return ".png"
	case FormatGIF:
		return ".gif"
	case FormatBMP:
		return ".bmp"
	case FormatTIFF:
		return ".tiff"
	case FormatWEBP:
		return ".webp"
	default:
		return "." + format
	}
}

// CanEncode reports whether we can write a given format.
func CanEncode(format string) bool {
	switch NormalizeFormat(format) {
	case FormatJPEG, FormatPNG, FormatGIF, FormatBMP, FormatTIFF, FormatWEBP:
		return true
	default:
		return false
	}
}

// IsImageFile is a cheap extension check used when scanning folders for batch
// processing. It does not open the file.
func IsImageFile(path string) bool {
	switch NormalizeFormat(filepath.Ext(path)) {
	case FormatJPEG, FormatPNG, FormatGIF, FormatBMP, FormatTIFF, FormatWEBP:
		return true
	default:
		return false
	}
}

// Load reads an image from disk and returns it along with its detected format.
func Load(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", fmt.Errorf("decoding %s: %w", path, err)
	}
	return img, NormalizeFormat(format), nil
}

// EncodeOptions controls lossy encoders.
type EncodeOptions struct {
	// Quality is 1-100 for JPEG and WebP. Ignored by lossless formats.
	Quality int
	// PNGCompression maps to png.CompressionLevel (0 = default).
	PNGCompression png.CompressionLevel
}

// Encode writes an image into the requested format and returns the raw bytes.
// Keeping this in-memory lets the compressor try several settings without
// touching the disk until it is happy with the result.
func Encode(img image.Image, format string, opts EncodeOptions) ([]byte, error) {
	var buf bytes.Buffer
	format = NormalizeFormat(format)

	switch format {
	case FormatJPEG:
		q := opts.Quality
		if q <= 0 {
			q = 90
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, err
		}
	case FormatPNG:
		enc := png.Encoder{CompressionLevel: opts.PNGCompression}
		if err := enc.Encode(&buf, img); err != nil {
			return nil, err
		}
	case FormatGIF:
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, err
		}
	case FormatBMP:
		if err := bmp.Encode(&buf, img); err != nil {
			return nil, err
		}
	case FormatTIFF:
		if err := tiff.Encode(&buf, img, nil); err != nil {
			return nil, err
		}
	case FormatWEBP:
		b, err := encodeWebP(img, opts.Quality)
		if err != nil {
			return nil, err
		}
		return b, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q", format)
	}
	return buf.Bytes(), nil
}

// Save encodes an image and writes it to path, creating parent dirs as needed.
func Save(img image.Image, path, format string, opts EncodeOptions) error {
	data, err := Encode(img, format, opts)
	if err != nil {
		return err
	}
	return WriteBytes(path, data)
}

// WriteBytes writes raw bytes to path, creating parent directories.
func WriteBytes(path string, data []byte) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

// ParseSize converts a human-friendly size like "50KB", "1.5MB", "500k" or a
// raw byte count "204800" into bytes.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	var mult float64 = 1
	switch {
	case strings.HasSuffix(s, "KB"), strings.HasSuffix(s, "K"):
		mult = 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "KB"), "K")
	case strings.HasSuffix(s, "MB"), strings.HasSuffix(s, "M"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "MB"), "M")
	case strings.HasSuffix(s, "GB"), strings.HasSuffix(s, "G"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "GB"), "G")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}

	s = strings.TrimSpace(s)
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	if val <= 0 {
		return 0, fmt.Errorf("size must be positive")
	}
	return int64(val * mult), nil
}

// HumanSize formats a byte count for display.
func HumanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
