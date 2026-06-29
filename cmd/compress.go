package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/singhvibhanshu/imago/internal/imageio"
	"github.com/singhvibhanshu/imago/internal/process"
)

var (
	compressQuality int
	compressTarget  string
	compressTo      string
)

var compressCmd = &cobra.Command{
	Use:   "compress <input> [more...]",
	Short: "Shrink images by quality or to a target file size",
	Long: `Compress images, either by a fixed quality or down to a target file size.

The --target mode is ideal for online forms that demand, say, a photo under
50 KB: imago lowers quality first and, if needed, scales the dimensions until
the file fits.

Examples:
  imago compress photo.jpg --quality 70
  imago compress photo.jpg --target 50KB
  imago compress ./photos --target 100KB --out ./small`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var maxBytes int64
		if compressTarget != "" {
			var err error
			maxBytes, err = imageio.ParseSize(compressTarget)
			if err != nil {
				return err
			}
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

			// Output format defaults to the source format.
			outFormat := srcFormat
			if compressTo != "" {
				outFormat = imageio.NormalizeFormat(compressTo)
			}
			if !imageio.CanEncode(outFormat) {
				return "", fmt.Errorf("cannot write format %q", outFormat)
			}

			newBase := baseName(in) + ".min" + imageio.Extension(outFormat)
			out, err := outputPath(in, newBase, batch)
			if err != nil {
				return "", err
			}

			origSize := int64(-1)
			if info, statErr := statSize(in); statErr == nil {
				origSize = info
			}

			if maxBytes > 0 {
				res, err := process.CompressToTarget(img, outFormat, maxBytes)
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
				if !res.Met {
					return fmt.Sprintf("  ⚠ %s → %s (%s, smallest achievable%s; target %s not reachable)",
						in, out, imageio.HumanSize(int64(len(res.Data))), note, compressTarget), nil
				}
				return fmt.Sprintf("  ✓ %s (%s) → %s (%s%s)",
					in, sizeStr(origSize), out, imageio.HumanSize(int64(len(res.Data))), note), nil
			}

			// Fixed-quality path.
			data, err := process.CompressToQuality(img, outFormat, compressQuality)
			if err != nil {
				return "", err
			}
			if err := imageio.WriteBytes(out, data); err != nil {
				return "", err
			}
			return fmt.Sprintf("  ✓ %s (%s) → %s (%s, q%d)",
				in, sizeStr(origSize), out, imageio.HumanSize(int64(len(data))), compressQuality), nil
		})
	},
}

func init() {
	compressCmd.Flags().IntVar(&compressQuality, "quality", 80, "quality 1-100 (ignored when --target is set)")
	compressCmd.Flags().StringVar(&compressTarget, "target", "", "target max file size, e.g. 50KB, 1.5MB")
	compressCmd.Flags().StringVarP(&compressTo, "to", "t", "", "output format (default: same as input)")
}
