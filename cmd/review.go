package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"DayOneCLI/internal/review"
	"DayOneCLI/internal/statistics"
	"DayOneCLI/internal/storage"
	"github.com/spf13/cobra"
)

// 최상위 review 명령어
var reviewMgmtCmd = &cobra.Command{
	Use:   "review",
	Short: "Manage your daily reviews",
}

// DB를 열고 Repository를 생성한 뒤 작업 실행
func withReviewRepo(
	action func(*review.Repository) error,
) error {
	db, err := storage.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	repo := review.NewRepository(db)
	return action(repo)
}

// 터미널에서 리뷰 내용 입력
func readReviewContent(reader *bufio.Reader) (string, error) {
	fmt.Print("리뷰 내용 (선택사항, 엔터로 스킵) > ")

	content, err := reader.ReadString('\n')

	if err != nil && err != io.EOF {
		return "", err
	}

	content = strings.TrimSpace(content)

	return content, nil
}

// CREATE: 일일 리뷰 작성
var reviewWriteCmd = &cobra.Command{
	Use:   "write [date]",
	Short: "Write a daily review",
	Args:  cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		date := time.Now().Format("2006-01-02")
		if len(args) > 0 {
			date = args[0]
		}

		// 감정 입력
		fmt.Print("감정 (😊/😐/😢/😡/😴) > ")
		reader := bufio.NewReader(os.Stdin)
		mood, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		mood = strings.TrimSpace(mood)

		// 평점 입력
		fmt.Print("평점 (1~5) > ")
		ratingStr, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		ratingStr = strings.TrimSpace(ratingStr)
		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			return fmt.Errorf("평점은 숫자여야 합니다")
		}

		// 내용 입력
		content, err := readReviewContent(reader)
		if err != nil {
			return err
		}

		err = withReviewRepo(func(r *review.Repository) error {
			return r.Create(date, mood, rating, content)
		})
		if err != nil {
			return err
		}

		fmt.Printf("리뷰가 저장되었습니다: %s %s %d점\n", date, mood, rating)
		return nil
	},
}

// READ: 특정 날짜 리뷰 조회
var reviewReadCmd = &cobra.Command{
	Use:   "read [date]",
	Short: "Read a daily review",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return withReviewRepo(func(r *review.Repository) error {
			rev, err := r.GetByDate(args[0])

			if review.IsNotFound(err) {
				return fmt.Errorf("해당 날짜의 리뷰가 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("날짜:", rev.Date)
			fmt.Println("감정:", rev.Mood)
			fmt.Println("평점:", rev.Rating)
			if rev.Content != "" {
				fmt.Println("내용:", rev.Content)
			}
			return nil
		})
	},
}

// LIST: 전체 리뷰 목록
var reviewListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all daily reviews",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return withReviewRepo(func(r *review.Repository) error {
			reviews, err := r.List()
			if err != nil {
				return err
			}

			if len(reviews) == 0 {
				fmt.Println("작성된 리뷰가 없습니다.")
				return nil
			}

			for _, rev := range reviews {
				stars := statistics.FormatStars(float64(rev.Rating))
				fmt.Printf("%s %s [%s]\n", rev.Date, rev.Mood, stars)
				if rev.Content != "" {
					fmt.Printf("  %s\n", rev.Content)
				}
			}

			return nil
		})
	},
}

// DELETE: 리뷰 삭제
var reviewDeleteCmd = &cobra.Command{
	Use:   "delete [date]",
	Short: "Delete a daily review",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return withReviewRepo(func(r *review.Repository) error {
			err := r.Delete(args[0])

			if review.IsNotFound(err) {
				return fmt.Errorf("삭제할 리뷰가 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("리뷰가 삭제되었습니다.")
			return nil
		})
	},
}

func init() {
	reviewMgmtCmd.AddCommand(
		reviewWriteCmd,
		reviewReadCmd,
		reviewListCmd,
		reviewDeleteCmd,
	)

	rootCmd.AddCommand(reviewMgmtCmd)
}
