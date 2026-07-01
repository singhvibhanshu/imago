package tui

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/singhvibhanshu/imago/internal/imageio"
)

// makeTestPNG writes a small PNG and returns its path.
func makeTestPNG(t *testing.T, dir string) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 120, 90))
	for y := 0; y < 90; y++ {
		for x := 0; x < 120; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 2), uint8(y * 2), 100, 255})
		}
	}
	data, err := imageio.Encode(img, "png", imageio.EncodeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "sample.png")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestOperations(t *testing.T) {
	dir := t.TempDir()
	in := makeTestPNG(t, dir)

	if _, err := runStrip(in); err != nil {
		t.Fatalf("strip: %v", err)
	}
	assertExists(t, filepath.Join(dir, "sample.stripped.png"))

	if _, err := runConvert(in, "jpg"); err != nil {
		t.Fatalf("convert: %v", err)
	}
	assertExists(t, filepath.Join(dir, "sample.jpg"))

	if _, err := runCompress(in, false, "50"); err != nil {
		t.Fatalf("compress quality: %v", err)
	}
	assertExists(t, filepath.Join(dir, "sample.min.png"))

	if _, err := runCompress(in, true, "3KB"); err != nil {
		t.Fatalf("compress target: %v", err)
	}

	if _, err := runResize(in, 60, 0, 0); err != nil {
		t.Fatalf("resize: %v", err)
	}
	assertExists(t, filepath.Join(dir, "sample.60x45.png"))
}

func TestValidators(t *testing.T) {
	if err := validQuality("80"); err != nil {
		t.Errorf("valid quality rejected: %v", err)
	}
	if err := validQuality("0"); err == nil {
		t.Error("quality 0 should be rejected")
	}
	if err := validSize("50KB"); err != nil {
		t.Errorf("valid size rejected: %v", err)
	}
	if err := validSize(""); err == nil {
		t.Error("empty size should be rejected")
	}
	if err := validOptInt(""); err != nil {
		t.Error("blank width should be allowed")
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected output %s to exist: %v", path, err)
	}
}
