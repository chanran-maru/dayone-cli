package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's overview",

	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()

		fmt.Println("=== DayOne Today ===")
		fmt.Println("Date:", now.Format("2006-01-02 Monday"))
		fmt.Println("Events: 0")
		fmt.Println("Tasks: 0")
		fmt.Println("Journal: Not written")
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
