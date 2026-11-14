package cmd

import (
	"fmt"
	"time"

	"DayOneCLI/internal/journal"
	"DayOneCLI/internal/review"
	"DayOneCLI/internal/storage"
	"DayOneCLI/internal/task"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's overview",

	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		date := now.Format("2006-01-02")

		db, err := storage.OpenDB()
		if err != nil {
			return err
		}
		defer db.Close()

		fmt.Println("=== DayOne Today ===")
		fmt.Printf("Date: %s %s\n\n", date, now.Format("Monday"))

		// 일기 확인
		journalRepo := journal.NewRepository(db)
		j, err := journalRepo.GetByDate(date)
		if err == nil {
			fmt.Printf("📝 Journal: %s\n", j.Content)
		} else if journal.IsNotFound(err) {
			fmt.Println("📝 Journal: Not written")
		} else {
			return fmt.Errorf("load today's journal: %w", err)
		}

		// 작업 확인
		taskRepo := task.NewRepository(db)
		pendingTasks, err := taskRepo.List("pending")
		if err != nil {
			return fmt.Errorf("load pending tasks: %w", err)
		}
		if len(pendingTasks) > 0 {
			fmt.Printf("\n✅ Pending Tasks (%d):\n", len(pendingTasks))
			for _, t := range pendingTasks {
				fmt.Printf("  ⬜ [%d] %s\n", t.ID, t.Title)
			}
		} else {
			fmt.Println("\n✅ Pending Tasks: None")
		}

		// 리뷰 확인
		reviewRepo := review.NewRepository(db)
		rev, err := reviewRepo.GetByDate(date)
		if err == nil {
			fmt.Printf("\n📊 Review: %s %d/5\n", rev.Mood, rev.Rating)
			if rev.Content != "" {
				fmt.Printf("  %s\n", rev.Content)
			}
		} else if review.IsNotFound(err) {
			fmt.Println("\n📊 Review: Not written")
		} else {
			return fmt.Errorf("load today's review: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
