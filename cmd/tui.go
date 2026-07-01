package cmd

import (
	"github.com/spf13/cobra"

	"github.com/singhvibhanshu/imago/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"wizard", "interactive"},
	Short:   "Launch the interactive guided wizard",
	Long: `Launch imago's guided wizard: it walks you through choosing a file, an
operation, and its options one step at a time, with no flags to remember.

Running imago with no arguments opens the same wizard.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}
