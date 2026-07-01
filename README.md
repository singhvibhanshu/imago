# imago

[![npm version](https://img.shields.io/npm/v/@singhvibhanshu/imago.svg)](https://www.npmjs.com/package/@singhvibhanshu/imago)
[![license](https://img.shields.io/npm/l/@singhvibhanshu/imago.svg)](LICENSE)
![platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Windows-blue)

**A fully offline image toolkit: convert, compress and resize images right on your own machine. Your photos are never uploaded.**

Most "convert to PNG" or "compress to 50 KB" websites work by uploading your
photo to *their* servers, where it might be stored, cached, or misused. `imago`
does the exact same jobs as a local command-line tool, so your images never
leave your computer. No internet, no accounts, no privacy worries.

It ships as a single self-contained binary per platform: no runtime, no system
libraries, no cgo.

> Installing needs internet (to fetch the package once). The tool itself runs
> 100% offline afterwards.

## Install

```bash
# npm (global)
npm install -g @singhvibhanshu/imago

# or run on demand, no install
npx @singhvibhanshu/imago --help

# Bun (uses the npm registry too)
bun add -g @singhvibhanshu/imago
bunx @singhvibhanshu/imago --help

# Go developers
go install github.com/singhvibhanshu/imago@latest
```

## Features

| Command    | What it does |
|------------|--------------|
| `convert`  | Convert between `jpg`, `png`, `webp`, `gif`, `bmp`, `tiff` |
| `compress` | Shrink by quality, or down to a **target file size** (e.g. under 50 KB) |
| `resize`   | Resize by exact pixel dimensions or percentage, aspect ratio preserved |
| `strip`    | Remove hidden metadata (EXIF, **GPS location**, timestamps), losslessly |
| `tui`      | Interactive terminal UI: run `imago` with no arguments |
| *(batch)*  | Point any command at a folder to process every image at once |

## Interactive mode (TUI)

Prefer not to remember flags? Just run imago with no arguments to open an
interactive terminal UI: browse for an image, pick an operation, and see the
result, all with the arrow keys:

```bash
imago            # or: imago tui
```

```
🖼  imago  ·  interactive mode
  What do you want to do?
  › Convert format   Change to jpg, png, webp, gif, bmp, tiff
    Compress         By quality or down to a target file size
    Resize           Change pixel dimensions or scale by percent
    Strip metadata   Remove EXIF / GPS / hidden data (lossless)
```

The full flag-based CLI below still works exactly the same, great for scripts
and batch jobs.

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
imago compress photo.jpg --target 1.5MB
imago compress ./photos --target 100KB --out ./small    # whole folder
```

When you pass `--target`, imago first lowers JPEG quality to fit; if that isn't
enough (or the format is lossless, like PNG), it automatically scales the
dimensions down until the file fits. This is exactly what competitive-exam and
government form uploads demand: *"photo must be under 50 KB"*.

### Resize
```bash
imago resize photo.jpg --width 600                      # height auto (keep ratio)
imago resize photo.jpg --width 600 --height 800         # fit inside 600x800 box
imago resize photo.jpg --width 600 --height 800 --stretch   # exact, ignore ratio
imago resize photo.jpg --percent 50                     # half size
imago resize ./photos --width 1024 --out ./resized      # whole folder
```

### Strip metadata (privacy)
```bash
imago strip photo.jpg                                   # remove EXIF/GPS, lossless
imago strip ./photos --out ./clean                      # whole folder
```

Photos from phones embed EXIF metadata, including the **exact GPS coordinates**
where the picture was taken. `strip` removes it without changing the image: for
JPEG and PNG the pixel data is left byte-for-byte identical, only the metadata is
deleted.

> `convert`, `compress` and `resize` already produce metadata-free output as a
> side effect. `strip` is for when you want to scrub metadata *without*
> otherwise altering the image.

Run `imago <command> --help` to see every flag for a command.

## Global flags

| Flag | Description |
|------|-------------|
| `-o, --out` | Output file or directory (default: alongside the input) |
| `-r, --recursive` | Recurse into subdirectories when given a folder |
| `--overwrite` | Allow overwriting an existing file |
| `-q, --quiet` | Only print errors |

## Output file names

By default, outputs are written next to the input and the original is **never**
overwritten:

- **convert** → same name, new extension (`photo.jpg` → `photo.png`)
- **compress** → `.min` suffix (`photo.jpg` → `photo.min.jpg`)
- **resize** → dimensions in the name (`photo.jpg` → `photo.600x400.jpg`)

Use `--out` to send results to a specific file or folder.

## How it works

This package ships prebuilt native binaries written in Go. On install, your
package manager downloads only the small binary matching your operating system
and CPU (via per-platform `optionalDependencies`). No Go toolchain or
compilation is required to use it.

## Build from source

```bash
git clone https://github.com/singhvibhanshu/imago.git
cd imago
go build -o imago .
./imago --help
```

Requires Go 1.25+ to build from source.

## Releasing (maintainers)

1. Bump the version in `npm/imago/package.json`.
2. Commit, then tag and push:
   ```bash
   git tag v0.1.1
   git push --tags
   ```
3. GitHub Actions cross-compiles every platform and publishes to npm
   automatically (see `.github/workflows/publish.yml`).

To publish manually instead: `bash scripts/build-npm.sh && bash scripts/publish-npm.sh`.

## Notes

- WebP output uses a pure-Go **lossless** encoder, so `--quality` is ignored for
  WebP; size targeting for WebP relies on dimension scaling.
- Decoding works for all listed formats and auto-detects the format regardless
  of file extension.

## License

[MIT](LICENSE) © Vibhanshu Singh
