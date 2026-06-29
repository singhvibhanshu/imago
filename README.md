# imago

**A fully offline image toolkit. Convert, compress and resize images entirely on your own machine — nothing is ever uploaded.**

Most "convert JPG to PNG" or "compress image to 50KB" websites work by uploading
your photo to *their* servers. `imago` does the same jobs as a local command-line
tool, so your images never leave your computer. No internet required, no accounts,
no privacy worries.

It ships as a single self-contained binary (no runtime, no system libraries, no cgo).

## Features

- **Convert** between `jpg`, `jpeg`, `png`, `webp`, `gif`, `bmp`, `tiff`
- **Compress** by quality, or down to a **target file size** (e.g. "under 50 KB")
- **Resize** by exact pixel dimensions or percentage, with aspect-ratio lock
- **Batch** process an entire folder in one command

## Install

```bash
git clone <your-repo-url>
cd imago
go build -o imago .
# optional: move it onto your PATH
mv imago /usr/local/bin/
```

Requires Go 1.25+ to build. Once built, the binary needs nothing else.

## Usage

### Convert formats
```bash
imago convert photo.jpeg --to png
imago convert pic.png --to jpg --quality 85
imago convert ./album --to webp --out ./webp_album      # whole folder
```

### Compress
```bash
imago compress photo.jpg --quality 70                   # fixed quality
imago compress photo.jpg --target 50KB                  # fit under 50 KB
imago compress photo.jpg --target 1.5MB --to jpg
imago compress ./photos --target 100KB --out ./small    # whole folder
```

When you give a `--target`, imago first lowers JPEG quality to fit; if that isn't
enough (or the format is lossless, like PNG), it automatically scales the
dimensions down until the file fits. This is exactly what competitive-exam and
government form uploads usually require ("photo must be under 50 KB").

### Resize
```bash
imago resize photo.jpg --width 600                      # height auto (keep ratio)
imago resize photo.jpg --width 600 --height 800         # fit inside 600x800 box
imago resize photo.jpg --width 600 --height 800 --stretch   # exact, ignore ratio
imago resize photo.jpg --percent 50                     # half size
imago resize ./photos --width 1024 --out ./resized      # whole folder
```

## Global flags

| Flag | Description |
|------|-------------|
| `-o, --out` | Output file or directory (default: alongside the input) |
| `-r, --recursive` | Recurse into subdirectories when given a folder |
| `--overwrite` | Allow overwriting an existing file |
| `-q, --quiet` | Only print errors |

## How output files are named

- **convert** → same name, new extension (`photo.jpg` → `photo.png`)
- **compress** → `.min` suffix (`photo.jpg` → `photo.min.jpg`)
- **resize** → dimensions in the name (`photo.jpg` → `photo.600x400.jpg`)

By default outputs are written next to the input and the original is never
overwritten. Use `--out` to send results to a specific file or folder.

## Notes

- WebP output uses a pure-Go **lossless** encoder, so `--quality` is ignored for
  WebP; size targeting for WebP falls back to dimension scaling.
- Decoding supports all listed formats plus auto-detection regardless of extension.

## License

MIT (or your choice).
