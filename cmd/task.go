package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"DayOneCLI/internal/storage"
	"DayOneCLI/internal/task"
	"github.com/spf13/cobra"
)

// 최상위 task 명령어
var taskMgmtCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage your tasks",
}

// DB를 열고 Repository를 생성한 뒤 작업 실행
func withTaskRepo(
	action func(*task.Repository) error,
) error {
	db, err := storage.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	repo := task.NewRepository(db)
	return action(repo)
}

// 터미널에서 작업 제목 입력
func readTaskTitle() (string, error) {
	fmt.Print("작업 제목 > ")

	reader := bufio.NewReader(os.Stdin)
	title, err := reader.ReadString('\n')

	if err != nil && len(title) == 0 {
		return "", err
	}

	title = strings.TrimSpace(title)

	if title == "" {
		return "", fmt.Errorf("작업 제목은 비워둘 수 없습니다")
	}

	return title, nil
}

// CREATE: 작업 추가
var taskAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.TrimSpace(args[0])
		if title == "" {
			return fmt.Errorf("작업 제목은 비워둘 수 없습니다")
		}

		err := withTaskRepo(func(r *task.Repository) error {
			return r.Create(title)
		})
		if err != nil {
			return err
		}

		fmt.Println("작업이 추가되었습니다:", title)
		return nil
	},
}

// READ: 작업 목록
var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return withTaskRepo(func(r *task.Repository) error {
			tasks, err := r.ListAll()
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				fmt.Println("작업이 없습니다.")
				return nil
			}

			for _, t := range tasks {
				status := "⬜"
				if t.Status == "completed" {
					status = "✅"
				}
				fmt.Printf("%s [%d] %s\n", status, t.ID, t.Title)
			}

			return nil
		})
	},
}

// READ: 진행 중 작업
var taskPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending tasks",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return withTaskRepo(func(r *task.Repository) error {
			tasks, err := r.List("pending")
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				fmt.Println("진행 중인 작업이 없습니다.")
				return nil
			}

			for _, t := range tasks {
				fmt.Printf("⬜ [%d] %s\n", t.ID, t.Title)
			}

			return nil
		})
	},
}

// READ: 완료된 작업
var taskDoneCmd = &cobra.Command{
	Use:   "done",
	Short: "List completed tasks",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return withTaskRepo(func(r *task.Repository) error {
			tasks, err := r.List("completed")
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				fmt.Println("완료된 작업이 없습니다.")
				return nil
			}

			for _, t := range tasks {
				fmt.Printf("✅ [%d] %s\n", t.ID, t.Title)
			}

			return nil
		})
	},
}

// UPDATE: 작업 완료
var taskCompleteCmd = &cobra.Command{
	Use:   "complete [id]",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ID는 숫자여야 합니다")
		}

		return withTaskRepo(func(r *task.Repository) error {
			err := r.Complete(id)

			if task.IsNotFound(err) {
				return fmt.Errorf("완료할 작업이 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("작업이 완료되었습니다.")
			return nil
		})
	},
}

// DELETE: 작업 삭제
var taskDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ID는 숫자여야 합니다")
		}

		return withTaskRepo(func(r *task.Repository) error {
			err := r.Delete(id)

			if task.IsNotFound(err) {
				return fmt.Errorf("삭제할 작업이 없습니다")
			}
			if err != nil {
				return err
			}

			fmt.Println("작업이 삭제되었습니다.")
			return nil
		})
	},
}

func init() {
	taskMgmtCmd.AddCommand(
		taskAddCmd,
		taskListCmd,
		taskPendingCmd,
		taskDoneCmd,
		taskCompleteCmd,
		taskDeleteCmd,
	)

	rootCmd.AddCommand(taskMgmtCmd)
}
