package cmd

import (
	"DayOneCLI/internal/tui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dayone",
	Short: "Your personal daily life manager",
	Long:  "DayOne is a CLI and TUI tool for managing schedules, journals, and daily reviews.",

	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunDashboard()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
