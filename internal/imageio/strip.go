package imageio

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// StripMetadata removes metadata (EXIF, XMP, IPTC, comments, text chunks) from
// raw image bytes WITHOUT touching the actual image data: the pixels are left
// byte-for-byte identical. It sniffs the format from the content.
//
// The second return value reports whether a lossless strip was performed. For
// formats other than JPEG and PNG it returns (nil, false, nil); the caller can
// then fall back to decode-and-re-encode, which also drops metadata.
func StripMetadata(data []byte) ([]byte, bool, error) {
	switch {
	case len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8:
		out, err := stripJPEG(data)
		return out, true, err
	case len(data) >= 8 && bytes.Equal(data[:8], pngSignature):
		out, err := stripPNG(data)
		return out, true, err
	default:
		return nil, false, nil
	}
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// stripJPEG walks the JPEG marker segments and drops the metadata-carrying ones
// (EXIF/XMP in APP1, IPTC/Photoshop in APP13, other APPn, and COM comments),
// while keeping JFIF (APP0), ICC color profiles (APP2) and the Adobe marker
// (APP14) so the image still renders correctly. The entropy-coded scan data is
// copied verbatim, so the decoded pixels are unchanged.
func stripJPEG(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return nil, fmt.Errorf("not a valid JPEG")
	}

	out := make([]byte, 0, len(data))
	out = append(out, 0xFF, 0xD8) // SOI
	i := 2

	for i+1 < len(data) {
		if data[i] != 0xFF {
			// Shouldn't happen in a well-formed stream; copy the remainder.
			out = append(out, data[i:]...)
			break
		}
		marker := data[i+1]

		// Padding 0xFF bytes between segments.
		if marker == 0xFF {
			out = append(out, 0xFF)
			i++
			continue
		}
		// Start of Scan: image data follows with no length field; copy the rest.
		if marker == 0xDA {
			out = append(out, data[i:]...)
			break
		}
		// End of Image.
		if marker == 0xD9 {
			out = append(out, 0xFF, 0xD9)
			i += 2
			break
		}
		// Standalone markers with no payload (TEM, RSTn).
		if marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
			out = append(out, 0xFF, marker)
			i += 2
			continue
		}
		// Length-prefixed segment.
		if i+3 >= len(data) {
			out = append(out, data[i:]...)
			break
		}
		segLen := int(data[i+2])<<8 | int(data[i+3]) // includes the 2 length bytes
		end := i + 2 + segLen
		if end > len(data) {
			end = len(data)
		}

		if !keepJPEGSegment(marker) {
			i = end // drop this segment
			continue
		}
		out = append(out, data[i:end]...)
		i = end
	}
	return out, nil
}

// keepJPEGSegment reports whether a marker's segment should be preserved.
func keepJPEGSegment(marker byte) bool {
	switch {
	case marker == 0xFE: // COM comment
		return false
	case marker >= 0xE0 && marker <= 0xEF: // APPn
		// Keep JFIF (APP0), ICC profile (APP2) and Adobe (APP14); drop the rest
		// (EXIF/XMP in APP1, IPTC/Photoshop in APP13, etc.).
		return marker == 0xE0 || marker == 0xE2 || marker == 0xEE
	default:
		return true // DQT, DHT, SOF, etc.: essential, keep.
	}
}

// stripPNG drops ancillary text/metadata chunks (tEXt, zTXt, iTXt, eXIf, tIME)
// while preserving every chunk needed to render the image. IDAT (the pixel
// data) is copied verbatim.
func stripPNG(data []byte) ([]byte, error) {
	if len(data) < 8 || !bytes.Equal(data[:8], pngSignature) {
		return nil, fmt.Errorf("not a valid PNG")
	}

	drop := map[string]bool{
		"tEXt": true, "zTXt": true, "iTXt": true, // text metadata
		"eXIf": true, // EXIF
		"tIME": true, // last-modified timestamp
	}

	out := make([]byte, 0, len(data))
	out = append(out, data[:8]...) // signature
	i := 8

	for i+8 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[i : i+4]))
		ctype := string(data[i+4 : i+8])
		end := i + 12 + length // 4 (len) + 4 (type) + length (data) + 4 (crc)
		if end > len(data) {
			end = len(data)
		}

		if !drop[ctype] {
			out = append(out, data[i:end]...)
		}
		if ctype == "IEND" {
			break
		}
		i = end
	}
	return out, nil
}
