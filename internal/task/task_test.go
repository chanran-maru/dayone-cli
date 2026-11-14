package task

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestListByDateLocalDay(t *testing.T) {
	previous := time.Local
	time.Local = time.FixedZone("KST", 9*60*60)
	t.Cleanup(func() { time.Local = previous })
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT, status TEXT, created_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	for i, timestamp := range []string{
		"2026-11-13T23:59:59+09:00",
		"2026-11-14T00:00:00+09:00",
		"2026-11-13T15:30:00Z",
		"2026-11-14T23:59:59+09:00",
		"2026-11-15T00:00:00+09:00",
	} {
		status := "pending"
		if i == 3 {
			status = "completed"
		}
		if _, err := db.Exec("INSERT INTO tasks VALUES (?, ?, ?, ?)", i+1, "task", status, timestamp); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(db)
	tasks, err := repo.ListByDate("2026-11-14")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3: %+v", len(tasks), tasks)
	}
	ids := map[int]bool{}
	for _, task := range tasks {
		ids[task.ID] = true
	}
	for _, id := range []int{2, 3, 4} {
		if !ids[id] {
			t.Errorf("missing task %d", id)
		}
	}
	for _, date := range []string{"", "2026-02-30", "2026-11-4"} {
		if _, err := repo.ListByDate(date); err == nil {
			t.Errorf("accepted invalid date %q", date)
		}
	}
}
