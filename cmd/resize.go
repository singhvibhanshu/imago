package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"imago/internal/imageio"
	"imago/internal/process"
)

var (
	resizeWidth   int
	resizeHeight  int
	resizePercent float64
	resizeStretch bool
	resizeTo      string
	resizeQuality int
)

var resizeCmd = &cobra.Command{
	Use:   "resize <input> [more...]",
	Short: "Resize images by pixel dimensions or percentage",
	Long: `Resize images to exact pixel dimensions or by a percentage.

Aspect ratio is preserved by default. When both --width and --height are given,
the image is fit inside that box; pass --stretch to force exact dimensions.

Examples:
  imago resize photo.jpg --width 600
  imago resize photo.jpg --width 600 --height 800
  imago resize photo.jpg --percent 50
  imago resize ./photos --width 1024 --out ./resized`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if resizeWidth == 0 && resizeHeight == 0 && resizePercent == 0 {
			return fmt.Errorf("specify --width, --height and/or --percent")
		}

		params := process.ResizeParams{
			Width:      resizeWidth,
			Height:     resizeHeight,
			Percent:    resizePercent,
			KeepAspect: !resizeStretch,
		}

		inputs, err := gatherInputs(args)
		if err != nil {
			return err
		}
		batch := len(inputs) > 1

		return runBatch(inputs, func(in string) (string, error) {
			img, srcFormat, err := imageio.Load(in)
			if err != nil {
				return "", err
			}

			outFormat := srcFormat
			if resizeTo != "" {
				outFormat = imageio.NormalizeFormat(resizeTo)
			}
			if !imageio.CanEncode(outFormat) {
				return "", fmt.Errorf("cannot write format %q", outFormat)
			}

			resized := process.Resize(img, params)
			b := resized.Bounds()

			newBase := fmt.Sprintf("%s.%dx%d%s", baseName(in), b.Dx(), b.Dy(), imageio.Extension(outFormat))
			out, err := outputPath(in, newBase, batch)
			if err != nil {
				return "", err
			}
			if err := imageio.Save(resized, out, outFormat, imageio.EncodeOptions{Quality: resizeQuality}); err != nil {
				return "", err
			}
			ob := img.Bounds()
			return fmt.Sprintf("  ✓ %s (%dx%d) → %s (%dx%d)", in, ob.Dx(), ob.Dy(), out, b.Dx(), b.Dy()), nil
		})
	},
}

func init() {
	resizeCmd.Flags().IntVarP(&resizeWidth, "width", "w", 0, "target width in pixels")
	resizeCmd.Flags().IntVar(&resizeHeight, "height", 0, "target height in pixels")
	resizeCmd.Flags().Float64VarP(&resizePercent, "percent", "p", 0, "scale by percentage (e.g. 50 = half size)")
	resizeCmd.Flags().BoolVar(&resizeStretch, "stretch", false, "ignore aspect ratio and force exact width x height")
	resizeCmd.Flags().StringVarP(&resizeTo, "to", "t", "", "output format (default: same as input)")
	resizeCmd.Flags().IntVar(&resizeQuality, "quality", 90, "quality 1-100 for lossy output formats")
}
