package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"DayOneCLI/internal/journal"
	"DayOneCLI/internal/storage"
	"github.com/spf13/cobra"
)

// 최상위 journal 명령어
var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Manage your daily journal",
}

// DB를 열고 Repository를 생성한 뒤 작업 실행
func withJournalRepo(
	action func(*journal.Repository) error,
) error {
	db, err := storage.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	repo := journal.NewRepository(db)
	return action(repo)
}

// 터미널에서 일기 내용 입력
func readContent() (string, error) {
	fmt.Print("일기 내용 > ")

	reader := bufio.NewReader(os.Stdin)
	content, err := reader.ReadString('\n')

	if err != nil && len(content) == 0 {
		return "", err
	}

	content = strings.TrimSpace(content)

	if content == "" {
		return "", fmt.Errorf("일기 내용은 비워둘 수 없습니다")
	}

	return content, nil
}

// CREATE
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write today's journal",

	RunE: func(cmd *cobra.Command, args []string) error {
		date := time.Now().Format("2006-01-02")

		content, err := readContent()
		if err != nil {
			return err
		}

		err = withJournalRepo(func(r *journal.Repository) error {
			return r.Create(date, content)
		})
		if err != nil {
			return err
		}

		fmt.Println("일기가 저장되었습니다:", date)
		return nil
	},
}

// READ
var readCmd = &cobra.Command{
	Use:   "read [date]",
	Short: "Read a journal by date",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return withJournalRepo(func(r *journal.Repository) error {
			j, err := r.GetByDate(args[0])

			if journal.IsNotFound(err) {
				return fmt.Errorf("해당 날짜의 일기가 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("날짜:", j.Date)
			fmt.Println("내용:", j.Content)
			return nil
		})
	},
}

// LIST
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all journals",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return withJournalRepo(func(r *journal.Repository) error {
			journals, err := r.List()
			if err != nil {
				return err
			}

			if len(journals) == 0 {
				fmt.Println("작성된 일기가 없습니다.")
				return nil
			}

			for _, j := range journals {
				fmt.Println(j.Date, "-", j.Content)
			}

			return nil
		})
	},
}

// UPDATE
var editCmd = &cobra.Command{
	Use:   "edit [date]",
	Short: "Edit a journal",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := readContent()
		if err != nil {
			return err
		}

		return withJournalRepo(func(r *journal.Repository) error {
			err := r.Update(args[0], content)

			if journal.IsNotFound(err) {
				return fmt.Errorf("수정할 일기가 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("일기가 수정되었습니다.")
			return nil
		})
	},
}

// DELETE
var deleteCmd = &cobra.Command{
	Use:   "delete [date]",
	Short: "Delete a journal",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return withJournalRepo(func(r *journal.Repository) error {
			err := r.Delete(args[0])

			if journal.IsNotFound(err) {
				return fmt.Errorf("삭제할 일기가 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("일기가 삭제되었습니다.")
			return nil
		})
	},
}

func init() {
	journalCmd.AddCommand(
		writeCmd,
		readCmd,
		listCmd,
		editCmd,
		deleteCmd,
	)

	rootCmd.AddCommand(journalCmd)
}
