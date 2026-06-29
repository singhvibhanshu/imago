// Package cmd wires up the imago command-line interface.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/singhvibhanshu/imago/internal/imageio"
)

// Shared flags available to every subcommand.
var (
	flagOut       string
	flagRecursive bool
	flagOverwrite bool
	flagQuiet     bool
)

var rootCmd = &cobra.Command{
	Use:   "imago",
	Short: "Local, private image conversion, compression and resizing",
	Long: `imago is a fully offline image toolkit.

Convert formats, compress to a target file size, and resize images for forms —
all on your own machine. Your photos are never uploaded anywhere.

Examples:
  imago convert photo.jpg --to png
  imago compress photo.jpg --target 50KB
  imago resize photo.png --width 600 --height 800
  imago compress ./photos --target 100KB --out ./compressed   # whole folder`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the CLI and returns a process exit code.
func Execute() int {
	rootCmd.PersistentFlags().StringVarP(&flagOut, "out", "o", "", "output file or directory (default: alongside the input)")
	rootCmd.PersistentFlags().BoolVarP(&flagRecursive, "recursive", "r", false, "recurse into subdirectories when given a folder")
	rootCmd.PersistentFlags().BoolVar(&flagOverwrite, "overwrite", false, "allow overwriting an existing file")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "only print errors")

	rootCmd.AddCommand(convertCmd, compressCmd, resizeCmd, stripCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

// gatherInputs expands the positional args into a concrete list of image files.
// An arg may be a single file or a directory (scanned for images).
func gatherInputs(args []string) ([]string, error) {
	var files []string
	for _, a := range args {
		info, err := os.Stat(a)
		if err != nil {
			return nil, fmt.Errorf("cannot access %q: %w", a, err)
		}
		if !info.IsDir() {
			files = append(files, a)
			continue
		}
		// Directory: collect images within it.
		walkFn := func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != a && !flagRecursive {
					return filepath.SkipDir
				}
				return nil
			}
			if imageio.IsImageFile(path) {
				files = append(files, path)
			}
			return nil
		}
		if err := filepath.WalkDir(a, walkFn); err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no image files found in the given path(s)")
	}
	return files, nil
}

// outputPath decides where a processed file should be written.
//
//	newBase  - desired filename (e.g. "photo.png" or "photo.min.jpg")
//	batch    - true when processing more than one input
func outputPath(input, newBase string, batch bool) (string, error) {
	var out string
	switch {
	case flagOut == "":
		out = filepath.Join(filepath.Dir(input), newBase)
	case batch || isDir(flagOut) || endsWithSep(flagOut):
		out = filepath.Join(flagOut, newBase)
	default:
		out = flagOut // single input, explicit file path
	}

	// Never clobber the source unless explicitly allowed.
	if sameFile(out, input) && !flagOverwrite {
		ext := filepath.Ext(out)
		out = strings.TrimSuffix(out, ext) + ".out" + ext
	}
	if fileExists(out) && !flagOverwrite && !sameFile(out, input) {
		return out, nil // allowed: writing a new derivative next to existing ones
	}
	return out, nil
}

// runBatch applies fn to each input and prints a tidy summary.
func runBatch(inputs []string, fn func(in string) (string, error)) error {
	batch := len(inputs) > 1
	var ok, failed int
	for _, in := range inputs {
		msg, err := fn(in)
		if err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", in, err)
			continue
		}
		ok++
		if !flagQuiet {
			fmt.Println(msg)
		}
	}
	if batch && !flagQuiet {
		fmt.Printf("\nDone: %d succeeded, %d failed\n", ok, failed)
	}
	if failed > 0 && ok == 0 {
		return fmt.Errorf("all %d file(s) failed", failed)
	}
	return nil
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func endsWithSep(p string) bool {
	return strings.HasSuffix(p, string(os.PathSeparator)) || strings.HasSuffix(p, "/")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func sameFile(a, b string) bool {
	ap, err1 := filepath.Abs(a)
	bp, err2 := filepath.Abs(b)
	return err1 == nil && err2 == nil && ap == bp
}

// baseName returns the filename without its extension.
func baseName(path string) string {
	b := filepath.Base(path)
	return strings.TrimSuffix(b, filepath.Ext(b))
}

// statSize returns the size of a file in bytes.
func statSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// sizeStr renders a byte count, or "?" when unknown.
func sizeStr(n int64) string {
	if n < 0 {
		return "?"
	}
	return imageio.HumanSize(n)
}
