package cmd

import (
	"fmt"

	"DayOneCLI/internal/storage"
	"DayOneCLI/internal/statistics"
	"github.com/spf13/cobra"
)

// 최상위 statistics 명령어
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics",
}

// DB를 열고 Service를 생성한 뒤 작업 실행
func withStatsService(
	action func(*statistics.StatisticsService) error,
) error {
	db, err := storage.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	svc := statistics.NewService(db)
	return action(svc)
}

// 전체 통계
var statsOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show all statistics",

	RunE: func(cmd *cobra.Command, args []string) error {
		return withStatsService(func(s *statistics.StatisticsService) error {
			fmt.Println("=== DayOne Statistics ===")
			fmt.Println()

			// 일기 통계
			journalStats, err := s.GetJournalStats()
			if err != nil {
				return err
			}
			fmt.Println("📝 Journal")
			fmt.Printf("  Total: %d\n", journalStats.TotalCount)
			fmt.Printf("  This month: %d\n", journalStats.ThisMonth)
			fmt.Printf("  This week: %d\n", journalStats.ThisWeek)
			fmt.Printf("  Current streak: %d days\n", journalStats.CurrentStreak)
			fmt.Printf("  Longest streak: %d days\n\n", journalStats.LongestStreak)

			// 작업 통계
			taskStats, err := s.GetTaskStats()
			if err != nil {
				return err
			}
			fmt.Println("✅ Tasks")
			fmt.Printf("  Total: %d\n", taskStats.Total)
			fmt.Printf("  Pending: %d\n", taskStats.Pending)
			fmt.Printf("  Completed: %d\n\n", taskStats.Completed)

			// 리뷰 통계
			reviewStats, err := s.GetReviewStats()
			if err != nil {
				return err
			}
			fmt.Println("📊 Reviews")
			fmt.Printf("  Total: %d\n", reviewStats.TotalCount)
			if reviewStats.TotalCount > 0 {
				fmt.Printf("  Average rating: %.1f/5 %s\n", 
					reviewStats.AvgRating,
					statistics.FormatStars(reviewStats.AvgRating))
				fmt.Println("  Mood distribution:")
				for mood, count := range reviewStats.MoodCounts {
					fmt.Printf("    %s: %d\n", mood, count)
				}
			} else {
				fmt.Println("  No reviews yet")
			}

			return nil
		})
	},
}

// 일기 통계
var statsJournalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Show journal statistics",

	RunE: func(cmd *cobra.Command, args []string) error {
		return withStatsService(func(s *statistics.StatisticsService) error {
			stats, err := s.GetJournalStats()
			if err != nil {
				return err
			}

			fmt.Println("📝 Journal Statistics")
			fmt.Printf("  Total: %d\n", stats.TotalCount)
			fmt.Printf("  This month: %d\n", stats.ThisMonth)
			fmt.Printf("  This week: %d\n", stats.ThisWeek)
			fmt.Printf("  Current streak: %d days\n", stats.CurrentStreak)
			fmt.Printf("  Longest streak: %d days\n", stats.LongestStreak)

			return nil
		})
	},
}

// 작업 통계
var statsTaskCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Show task statistics",

	RunE: func(cmd *cobra.Command, args []string) error {
		return withStatsService(func(s *statistics.StatisticsService) error {
			stats, err := s.GetTaskStats()
			if err != nil {
				return err
			}

			fmt.Println("✅ Task Statistics")
			fmt.Printf("  Total: %d\n", stats.Total)
			fmt.Printf("  Pending: %d\n", stats.Pending)
			fmt.Printf("  Completed: %d\n", stats.Completed)

			return nil
		})
	},
}

// 리뷰 통계
var statsReviewCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Show review statistics",

	RunE: func(cmd *cobra.Command, args []string) error {
		return withStatsService(func(s *statistics.StatisticsService) error {
			stats, err := s.GetReviewStats()
			if err != nil {
				return err
			}

			fmt.Println("📊 Review Statistics")
			fmt.Printf("  Total: %d\n", stats.TotalCount)
			if stats.TotalCount > 0 {
				fmt.Printf("  Average rating: %.1f/5 %s\n", 
					stats.AvgRating,
					statistics.FormatStars(stats.AvgRating))
				fmt.Println("  Mood distribution:")
				for mood, count := range stats.MoodCounts {
					fmt.Printf("    %s: %d\n", mood, count)
				}
			} else {
				fmt.Println("  No reviews yet")
			}

			return nil
		})
	},
}

func init() {
	statsCmd.AddCommand(
		statsOverviewCmd,
		statsJournalCmd,
		statsTaskCmd,
		statsReviewCmd,
	)

	// stats 명령어 없이 실행하면 overview 표시
	statsCmd.RunE = statsOverviewCmd.RunE

	rootCmd.AddCommand(statsCmd)
}
