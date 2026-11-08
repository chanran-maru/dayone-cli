package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:     "calendar",
	Aliases: []string{"cal"},
	Short:   "Show monthly calendar",

	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()

		year, month, _ := now.Date()
		first := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())

		fmt.Printf("\n      %s %d\n", month, year)
		fmt.Println("Su Mo Tu We Th Fr Sa")

		for i := 0; i < int(first.Weekday()); i++ {
			fmt.Print("   ")
		}

 	last := first.AddDate(0, 1, -1).Day()

 	for day := 1; day <= last; day++ {
 		fmt.Printf("%2d ", day)

 		if (int(first.Weekday())+day-1)%7 == 0 {
 			fmt.Println()
 		}
 	}

 	fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}
