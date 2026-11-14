package tui

import (
	"database/sql"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"DayOneCLI/internal/storage"
	"DayOneCLI/internal/task"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestInputBackspace(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"", ""}, {"abc", "ab"}, {"오늘", "오"}, {"한글😊", "한글"},
	} {
		m := initialModel()
		m.state, m.input = stateInput, tt.input
		next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		got := next.(model).input
		if got != tt.want || !utf8.ValidString(got) {
			t.Errorf("backspace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func taskScreenDB(t *testing.T) *sql.DB {
	t.Helper()
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("requires isolated config directory")
	}
	dir := t.TempDir()
	t.Setenv("AppData", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	db, err := storage.OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Now()
	for i, entry := range []struct {
		title, status string
		day           int
	}{
		{"past task", "pending", -1},
		{"today pending", "pending", 0},
		{"today completed", "completed", 0},
		{"future task", "pending", 1},
	} {
		if _, err := db.Exec("INSERT INTO tasks (id, title, status, created_at) VALUES (?, ?, ?, ?)",
			i+1, entry.title, entry.status, now.AddDate(0, 0, entry.day).Format(time.RFC3339)); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func assertTodayTasks(t *testing.T, output string) {
	t.Helper()
	for _, text := range []string{"Today's Tasks", "⬜ [2] today pending", "✅ [3] today completed"} {
		if !strings.Contains(output, text) {
			t.Errorf("missing %q in %q", text, output)
		}
	}
	for _, text := range []string{"past task", "future task"} {
		if strings.Contains(output, text) {
			t.Errorf("unexpected %q in %q", text, output)
		}
	}
}

func TestViewTasksShowsTodayOnly(t *testing.T) {
	taskScreenDB(t)
	assertTodayTasks(t, executeTasks())
}

func TestTaskActionScreens(t *testing.T) {
	for _, index := range []int{2, 3} {
		t.Run(getSubMenu(3)[index], func(t *testing.T) {
			db := taskScreenDB(t)
			m := initialModel()
			m.state, m.parent, m.subMenu, m.subCursor = stateSubMenu, 3, getSubMenu(3), index
			next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			m = next.(model)
			if m.state != stateInput {
				t.Fatalf("state = %q", m.state)
			}
			view := ansi.Strip(m.View().Content)
			for _, text := range []string{"Today's Tasks", "[ ] [2] today pending", "[x] [3] today completed"} {
				if !strings.Contains(view, text) {
					t.Errorf("missing %q in task input screen", text)
				}
			}
			if strings.Contains(view, "past task") || strings.Contains(view, "future task") {
				t.Fatal("input screen contains tasks from another day")
			}
			m.input = "2"
			next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			m = next.(model)
			if m.state != stateResult || !strings.Contains(m.output, "Today's Tasks") {
				t.Fatalf("missing refreshed tasks: %+v", m)
			}
			got, err := task.NewRepository(db).GetByID(2)
			if index == 2 {
				if err != nil || got.Status != "completed" || !strings.Contains(m.output, "✅ [2] today pending") {
					t.Fatalf("completion did not refresh status: %+v, %v, %s", got, err, m.output)
				}
			} else if !task.IsNotFound(err) || strings.Contains(m.output, "today pending") {
				t.Fatalf("deletion did not refresh list: %v, %s", err, m.output)
			}
			if !strings.Contains(m.output, "✅ [3] today completed") {
				t.Fatal("completed tasks should remain visible")
			}
		})
	}
}

func TestTaskActionsRejectOtherDates(t *testing.T) {
	db := taskScreenDB(t)
	for _, item := range []string{"Complete Task", "Delete Task"} {
		for _, id := range []string{"1", "4", "999"} {
			if output := processInput(3, getInputHint(item), id); !strings.Contains(output, "today's tasks") {
				t.Fatalf("%s accepted task %s: %s", item, id, output)
			}
		}
	}
	for _, id := range []int{1, 4} {
		got, err := task.NewRepository(db).GetByID(id)
		if err != nil || got.Status != "pending" {
			t.Fatalf("task %d changed: %+v, %v", id, got, err)
		}
	}
}

func TestViewTasksEmptyToday(t *testing.T) {
	db := taskScreenDB(t)
	if _, err := db.Exec("DELETE FROM tasks WHERE id IN (2, 3)"); err != nil {
		t.Fatal(err)
	}
	output := executeTasks()
	if !strings.Contains(output, "No tasks for today.") || strings.Contains(output, "past task") {
		t.Fatalf("incorrect empty state: %s", output)
	}
}

func TestQInputAndQuit(t *testing.T) {
	m := initialModel()
	m.state = stateInput
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd != nil || next.(model).input != "q" {
		t.Fatal("q should be entered as text")
	}
	m.state = stateDashboard
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("q should quit from the dashboard")
	}
}

func TestCalendarRowsEndOnSaturday(t *testing.T) {
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	rows := strings.Split(strings.TrimSpace(executeCalendar()), "\n")[2:]
	for i, row := range rows {
		want := 7
		if i == 0 {
			want -= int(first.Weekday())
		}
		if i < len(rows)-1 && len(strings.Fields(row)) != want {
			t.Fatalf("calendar row %d = %q, want %d days", i, row, want)
		}
	}
}

func TestDatabaseQueryErrorsAreVisible(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("requires isolated config directory")
	}
	for _, table := range []string{"journals", "tasks", "reviews"} {
		t.Run(table, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("AppData", dir)
			t.Setenv("XDG_CONFIG_HOME", dir)
			db, err := storage.OpenDB()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err := db.Exec("DROP TABLE " + table); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec("CREATE TABLE " + table + " (date TEXT)"); err != nil {
				t.Fatal(err)
			}
			for name, output := range map[string]string{
				"today": executeToday(), "statistics": executeStats(), "widgets": renderWidgets(time.Now()),
			} {
				if !strings.Contains(output, "Error") {
					t.Errorf("%s hid query failure: %q", name, output)
				}
			}
		})
	}
}
