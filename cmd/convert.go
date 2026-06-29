package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"imago/internal/imageio"
)

var (
	convertTo      string
	convertQuality int
)

var convertCmd = &cobra.Command{
	Use:   "convert <input> [more...]",
	Short: "Convert images from one format to another",
	Long: `Convert images between formats (jpg, png, webp, gif, bmp, tiff).

Examples:
  imago convert photo.jpeg --to png
  imago convert pic.png --to jpg --quality 85
  imago convert ./album --to webp --out ./webp_album`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := imageio.NormalizeFormat(convertTo)
		if format == "" {
			return fmt.Errorf("--to is required (e.g. --to png)")
		}
		if !imageio.CanEncode(format) {
			return fmt.Errorf("cannot write format %q", convertTo)
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
			newBase := baseName(in) + imageio.Extension(format)
			out, err := outputPath(in, newBase, batch)
			if err != nil {
				return "", err
			}
			if err := imageio.Save(img, out, format, imageio.EncodeOptions{Quality: convertQuality}); err != nil {
				return "", err
			}
			return fmt.Sprintf("  ✓ %s (%s) → %s (%s)", in, srcFormat, out, format), nil
		})
	},
}

func init() {
	convertCmd.Flags().StringVarP(&convertTo, "to", "t", "", "target format: jpg, png, webp, gif, bmp, tiff (required)")
	convertCmd.Flags().IntVar(&convertQuality, "quality", 90, "quality 1-100 for lossy output formats")
}
