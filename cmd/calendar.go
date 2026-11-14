package cmd

import (
	"fmt"
	"time"

	"DayOneCLI/internal/task"

	"github.com/spf13/cobra"
)

var dateFlag string

var calendarCmd = &cobra.Command{
	Use:     "calendar",
	Aliases: []string{"cal"},
	Short:   "Show monthly calendar",

	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()

		year, month, _ := now.Date()
		today := now.Day()
		first := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())

		// 날짜별 작업 조회
		taskMap := make(map[int][]task.Task)
		if err := withTaskRepo(func(r *task.Repository) error {
			for day := 1; day <= first.AddDate(0, 1, -1).Day(); day++ {
				dateStr := fmt.Sprintf("%d-%02d-%02d", year, month, day)
				tasks, err := r.ListByDate(dateStr)
				if err == nil && len(tasks) > 0 {
					taskMap[day] = tasks
				}
			}
			return nil
		}); err != nil {
			fmt.Println("작업 조회 실패:", err)
		}

		fmt.Printf("\n      %s %d\n", month, year)
		fmt.Println("Su Mo Tu We Th Fr Sa")

		for i := 0; i < int(first.Weekday()); i++ {
			fmt.Print("   ")
		}

		last := first.AddDate(0, 1, -1).Day()

		for day := 1; day <= last; day++ {
			// 오늘 날짜 강조
			if day == today {
				fmt.Printf("►%2d◄ ", day)
			} else if _, hasTasks := taskMap[day]; hasTasks {
				fmt.Printf("[%2d] ", day)
			} else {
				fmt.Printf("%2d  ", day)
			}

			if (int(first.Weekday())+day-1)%7 == 0 {
				fmt.Println()
			}
		}

		fmt.Println()

		// 특정 날짜의 작업 표시
		if dateFlag != "" {
			fmt.Printf("\n%s 날짜의 작업:\n", dateFlag)
			err := withTaskRepo(func(r *task.Repository) error {
				tasks, err := r.ListByDate(dateFlag)
				if err != nil {
					return err
				}

				if len(tasks) == 0 {
					fmt.Println("  작업이 없습니다.")
					return nil
				}

				for _, t := range tasks {
					status := "⬜"
					if t.Status == "completed" {
						status = "✅"
					}
					fmt.Printf("  %s [%d] %s\n", status, t.ID, t.Title)
				}

				return nil
			})
			if err != nil {
				fmt.Println("작업 조회 실패:", err)
			}
		}
	},
}

func init() {
	calendarCmd.Flags().StringVar(&dateFlag, "date", "", "날짜별 작업 보기 (YYYY-MM-DD)")
	rootCmd.AddCommand(calendarCmd)
}
