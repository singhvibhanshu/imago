package tui

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/singhvibhanshu/imago/internal/imageio"
)

var imageExts = []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff", ".tif"}

var (
	okStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	errStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203"))
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// Run launches the guided, step-by-step wizard.
func Run() error {
	fmt.Println(titleStyle.Render("🖼  imago") + subtle.Render("  ·  interactive wizard  ·  nothing is uploaded"))

	for {
		cwd, _ := os.Getwd()
		var (
			filePath, action, format, cmode string
			quality, target, wStr, hStr, pStr string
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewFilePicker().
					Title("Select an image").
					Description("↑/↓ move · enter open folder / choose file · esc quit").
					CurrentDirectory(cwd).
					AllowedTypes(imageExts).
					Value(&filePath),
			),

			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What do you want to do?").
					Options(
						huh.NewOption("Convert format", "convert"),
						huh.NewOption("Compress", "compress"),
						huh.NewOption("Resize", "resize"),
						huh.NewOption("Strip metadata (EXIF / GPS)", "strip"),
					).
					Value(&action),
			),

			// Convert options
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Convert to which format?").
					Options(huh.NewOptions("jpg", "png", "webp", "gif", "bmp", "tiff")...).
					Value(&format),
			).WithHideFunc(func() bool { return action != "convert" }),

			// Compress: mode
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Compress by…").
					Options(
						huh.NewOption("Quality (pick 1-100)", "quality"),
						huh.NewOption("Target file size (e.g. under 50 KB)", "target"),
					).
					Value(&cmode),
			).WithHideFunc(func() bool { return action != "compress" }),

			// Compress: quality value
			huh.NewGroup(
				huh.NewInput().
					Title("Quality (1-100)").
					Placeholder("80").
					Value(&quality).
					Validate(validQuality),
			).WithHideFunc(func() bool { return !(action == "compress" && cmode == "quality") }),

			// Compress: target value
			huh.NewGroup(
				huh.NewInput().
					Title("Target size").
					Placeholder("e.g. 50KB, 1.5MB").
					Value(&target).
					Validate(validSize),
			).WithHideFunc(func() bool { return !(action == "compress" && cmode == "target") }),

			// Resize options
			huh.NewGroup(
				huh.NewNote().Description("Leave a field blank to auto-calc. Aspect ratio is preserved."),
				huh.NewInput().Title("Width (px)").Placeholder("blank = auto").Value(&wStr).Validate(validOptInt),
				huh.NewInput().Title("Height (px)").Placeholder("blank = auto").Value(&hStr).Validate(validOptInt),
				huh.NewInput().Title("Percent (%)").Placeholder("blank = none").Value(&pStr).Validate(validOptFloat),
			).WithHideFunc(func() bool { return action != "resize" }),
		).WithTheme(huh.ThemeCharm())

		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				break
			}
			return err
		}

		var result string
		var err error
		switch action {
		case "convert":
			result, err = runConvert(filePath, format)
		case "compress":
			if cmode == "target" {
				result, err = runCompress(filePath, true, target)
			} else {
				result, err = runCompress(filePath, false, quality)
			}
		case "resize":
			result, err = runResize(filePath, atoiOr(wStr, 0), atoiOr(hStr, 0), atofOr(pStr, 0))
		case "strip":
			result, err = runStrip(filePath)
		}

		fmt.Println()
		if err != nil {
			fmt.Println(errStyle.Render("✗ " + err.Error()))
		} else {
			fmt.Println(okStyle.Render("✓ Done") + "  " + result)
		}
		fmt.Println()

		var again bool
		confirm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().Title("Process another image?").Value(&again),
			),
		).WithTheme(huh.ThemeCharm())
		if err := confirm.Run(); err != nil || !again {
			break
		}
	}

	fmt.Println(subtle.Render("Thanks for using imago 👋"))
	return nil
}

// --- validators & parsing ---

func validQuality(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil // defaults to 80
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 100 {
		return errors.New("enter a number from 1 to 100")
	}
	return nil
}

func validSize(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("enter a size like 50KB or 1.5MB")
	}
	if _, err := imageio.ParseSize(s); err != nil {
		return errors.New("try something like 50KB, 200KB or 1.5MB")
	}
	return nil
}

func validOptInt(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if n, err := strconv.Atoi(s); err != nil || n < 0 {
		return errors.New("enter a whole number of pixels, or leave blank")
	}
	return nil
}

func validOptFloat(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if f, err := strconv.ParseFloat(s, 64); err != nil || f < 0 {
		return errors.New("enter a percentage, or leave blank")
	}
	return nil
}

func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}

func atofOr(s string, def float64) float64 {
	if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return f
	}
	return def
}
