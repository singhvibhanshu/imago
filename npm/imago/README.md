# imago

**A fully offline image toolkit — convert, compress and resize images right on your own machine. Nothing is ever uploaded.**

Most "convert to PNG" or "compress to 50 KB" websites upload your photo to their
servers. `imago` does the same jobs locally, so your images never leave your
computer.

```bash
npm install -g @singhvibhanshu/imago
imago --help

# or with npx / bun, no global install:
npx @singhvibhanshu/imago --help
bunx @singhvibhanshu/imago --help
```

> Installation needs internet (to fetch the package), but the tool itself runs
> 100% offline afterwards.

## Features

- **Convert** between `jpg`, `png`, `webp`, `gif`, `bmp`, `tiff`
- **Compress** by quality, or down to a **target file size** (e.g. under 50 KB)
- **Resize** by pixel dimensions or percentage, aspect ratio preserved
- **Strip** hidden metadata (EXIF, GPS location, timestamps) — losslessly
- **Batch** process whole folders

## Examples

```bash
imago convert photo.jpeg --to png
imago compress photo.jpg --target 50KB        # great for exam/form uploads
imago resize photo.jpg --width 600 --height 800
imago strip photo.jpg                          # remove EXIF/GPS metadata
imago compress ./photos --target 100KB --out ./small   # whole folder
```

See `imago <command> --help` for all options.

## How it works

This package ships prebuilt native binaries (written in Go). On install, npm
downloads only the small binary matching your operating system and CPU. No Go
toolchain or compilation required.

## License

MIT
