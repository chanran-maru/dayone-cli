package cmd

import (
	"bufio"
	"database/sql"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"DayOneCLI/internal/review"
	"DayOneCLI/internal/storage"
)

func isolatedDB(t *testing.T) *sql.DB {
	t.Helper()
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("config directory isolation is configured for Windows and Linux")
	}
	dir := t.TempDir()
	t.Setenv("AppData", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	db, err := storage.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestReviewWriteBufferedInput(t *testing.T) {
	for _, content := range []string{"오늘도 좋은 하루\n", "줄바꿈 없는 내용", "\n", ""} {
		t.Run(content, func(t *testing.T) {
			db := isolatedDB(t)
			input, err := os.CreateTemp(t.TempDir(), "stdin")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { input.Close() })
			if _, err := io.WriteString(input, "happy\n4\n"+content); err != nil {
				t.Fatal(err)
			}
			if _, err := input.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			previous := os.Stdin
			os.Stdin = input
			t.Cleanup(func() { os.Stdin = previous })
			if err := reviewWriteCmd.RunE(reviewWriteCmd, []string{"2026-11-14"}); err != nil {
				t.Fatal(err)
			}
			got, err := review.NewRepository(db).GetByDate("2026-11-14")
			if err != nil {
				t.Fatal(err)
			}
			if got.Mood != "happy" || got.Rating != 4 || got.Content != strings.TrimSpace(content) {
				t.Fatalf("saved review = %+v", got)
			}
		})
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestReadReviewContentReadError(t *testing.T) {
	want := errors.New("input failed")
	_, err := readReviewContent(bufio.NewReader(io.MultiReader(strings.NewReader("partial"), failingReader{want})))
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestTodayQueryErrors(t *testing.T) {
	for _, table := range []string{"journals", "tasks", "reviews"} {
		t.Run(table, func(t *testing.T) {
			db := isolatedDB(t)
			if _, err := db.Exec("DROP TABLE " + table); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec("CREATE TABLE " + table + " (date TEXT)"); err != nil {
				t.Fatal(err)
			}
			if err := todayCmd.RunE(todayCmd, nil); err == nil {
				t.Fatal("query failure was hidden")
			}
		})
	}
}

func TestTodayEmptyDatabase(t *testing.T) {
	isolatedDB(t)
	if err := todayCmd.RunE(todayCmd, nil); err != nil {
		t.Fatal(err)
	}
}
