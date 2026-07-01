package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/singhvibhanshu/imago/internal/imageio"
)

var stripCmd = &cobra.Command{
	Use:   "strip <input> [more...]",
	Short: "Remove hidden metadata (EXIF, GPS location, etc.) from images",
	Long: `Remove hidden metadata from images without changing the picture itself.

Photos from phones and cameras embed EXIF metadata, including the exact GPS
location where the photo was taken, the date/time, and your device model. This
can leak private information when you share a photo. 'strip' removes it.

For JPEG and PNG the removal is lossless: the image data is left byte-for-byte
identical, only the metadata is deleted. (Note: convert, compress and resize
already produce metadata-free output as a side effect.)

Examples:
  imago strip photo.jpg
  imago strip ./photos --out ./clean        # whole folder`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputs, err := gatherInputs(args)
		if err != nil {
			return err
		}
		batch := len(inputs) > 1

		return runBatch(inputs, func(in string) (string, error) {
			data, err := os.ReadFile(in)
			if err != nil {
				return "", err
			}

			stripped, lossless, err := imageio.StripMetadata(data)
			if err != nil {
				return "", err
			}

			ext := filepath.Ext(in)
			newBase := baseName(in) + ".stripped" + ext
			out, err := outputPath(in, newBase, batch)
			if err != nil {
				return "", err
			}

			if lossless {
				if err := imageio.WriteBytes(out, stripped); err != nil {
					return "", err
				}
				removed := int64(len(data) - len(stripped))
				if removed <= 0 {
					return fmt.Sprintf("  ✓ %s → %s (no metadata found)", in, out), nil
				}
				return fmt.Sprintf("  ✓ %s → %s (removed %s of metadata)", in, out, imageio.HumanSize(removed)), nil
			}

			// Fallback for non-JPEG/PNG: decode and re-encode, which drops metadata.
			img, format, err := imageio.Load(in)
			if err != nil {
				return "", err
			}
			if err := imageio.Save(img, out, format, imageio.EncodeOptions{Quality: 95}); err != nil {
				return "", err
			}
			return fmt.Sprintf("  ✓ %s → %s (re-encoded; metadata removed)", in, out), nil
		})
	},
}
