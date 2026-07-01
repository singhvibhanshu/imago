package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/singhvibhanshu/imago/internal/imageio"
	"github.com/singhvibhanshu/imago/internal/process"
)

// The run* functions each perform one operation on a file and return a
// human-readable summary. They write output alongside the input and never
// overwrite the original. They drive the same internal packages as the CLI.

func runStrip(in string) (string, error) {
	data, err := os.ReadFile(in)
	if err != nil {
		return "", err
	}
	out := derive(in, ".stripped", filepath.Ext(in))
	stripped, lossless, err := imageio.StripMetadata(data)
	if err != nil {
		return "", err
	}
	if lossless {
		if err := imageio.WriteBytes(out, stripped); err != nil {
			return "", err
		}
		removed := int64(len(data) - len(stripped))
		if removed <= 0 {
			return fmt.Sprintf("No metadata found.  Wrote %s", out), nil
		}
		return fmt.Sprintf("Removed %s of metadata (pixels untouched).  Wrote %s",
			imageio.HumanSize(removed), out), nil
	}
	img, format, err := imageio.Load(in)
	if err != nil {
		return "", err
	}
	if err := imageio.Save(img, out, format, imageio.EncodeOptions{Quality: 95}); err != nil {
		return "", err
	}
	return fmt.Sprintf("Re-encoded to drop metadata.  Wrote %s", out), nil
}

func runConvert(in, format string) (string, error) {
	format = imageio.NormalizeFormat(format)
	img, srcFormat, err := imageio.Load(in)
	if err != nil {
		return "", err
	}
	out := derive(in, "", imageio.Extension(format))
	if sameName(out, in) {
		out = derive(in, ".converted", imageio.Extension(format))
	}
	if err := imageio.Save(img, out, format, imageio.EncodeOptions{Quality: 90}); err != nil {
		return "", err
	}
	sz, _ := fileSize(out)
	return fmt.Sprintf("Converted %s → %s.  Wrote %s (%s)",
		srcFormat, format, out, imageio.HumanSize(sz)), nil
}

func runCompress(in string, targetMode bool, value string) (string, error) {
	img, format, err := imageio.Load(in)
	if err != nil {
		return "", err
	}
	out := derive(in, ".min", imageio.Extension(format))
	origSize, _ := fileSize(in)
	value = strings.TrimSpace(value)

	if targetMode {
		maxBytes, err := imageio.ParseSize(value)
		if err != nil {
			return "", fmt.Errorf("invalid target size %q", value)
		}
		res, err := process.CompressToTarget(img, format, maxBytes)
		if err != nil {
			return "", err
		}
		if err := imageio.WriteBytes(out, res.Data); err != nil {
			return "", err
		}
		note := ""
		if res.Scale < 1.0 {
			note = fmt.Sprintf(", resized to %dx%d", res.Width, res.Height)
		}
		status := "fits target"
		if !res.Met {
			status = "smallest achievable (target not reachable)"
		}
		return fmt.Sprintf("%s → %s%s [%s].  Wrote %s", imageio.HumanSize(origSize),
			imageio.HumanSize(int64(len(res.Data))), note, status, out), nil
	}

	q := 80
	if value != "" {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 100 {
			return "", fmt.Errorf("quality must be a number 1-100")
		}
		q = n
	}
	data, err := process.CompressToQuality(img, format, q)
	if err != nil {
		return "", err
	}
	if err := imageio.WriteBytes(out, data); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s → %s (quality %d).  Wrote %s", imageio.HumanSize(origSize),
		imageio.HumanSize(int64(len(data))), q, out), nil
}

func runResize(in string, w, h int, pct float64) (string, error) {
	if w == 0 && h == 0 && pct == 0 {
		return "", fmt.Errorf("enter a width, height, and/or percent")
	}
	img, format, err := imageio.Load(in)
	if err != nil {
		return "", err
	}
	resized := process.Resize(img, process.ResizeParams{
		Width: w, Height: h, Percent: pct, KeepAspect: true,
	})
	b := resized.Bounds()
	ob := img.Bounds()
	out := derive(in, fmt.Sprintf(".%dx%d", b.Dx(), b.Dy()), imageio.Extension(format))
	if err := imageio.Save(resized, out, format, imageio.EncodeOptions{Quality: 90}); err != nil {
		return "", err
	}
	return fmt.Sprintf("%dx%d → %dx%d.  Wrote %s", ob.Dx(), ob.Dy(), b.Dx(), b.Dy(), out), nil
}

// --- helpers ---

func derive(in, suffix, ext string) string {
	dir := filepath.Dir(in)
	base := filepath.Base(in)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join(dir, base+suffix+ext)
}

func sameName(a, b string) bool {
	ap, _ := filepath.Abs(a)
	bp, _ := filepath.Abs(b)
	return ap == bp
}

func fileSize(p string) (int64, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
