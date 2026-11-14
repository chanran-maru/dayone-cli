package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"DayOneCLI/internal/journal"
	"DayOneCLI/internal/review"
	"DayOneCLI/internal/statistics"
	"DayOneCLI/internal/storage"
	"DayOneCLI/internal/task"
)

// 상태 상수
const (
	stateDashboard = "dashboard"
	stateResult    = "result"
	stateSubMenu   = "submenu"
	stateInput     = "input"
)

type model struct {
	cursor      int
	menus       []string
	state       string // "dashboard", "result", "submenu", "input"
	output      string
	subMenu     []string // 하위 메뉴 항목
	subCursor   int      // 하위 메뉴 커서
	parent      int      // 상위 메뉴 인덱스
	input       string   // 입력 모드에서 입력 중인 텍스트
	inputHint   string   // 입력 프롬프트
	width       int
	height      int
	scroll      int
	summary     string
	taskPreview string
}

type dashboardLoadedMsg struct {
	summary string
	tasks   string
}

func loadDashboard() tea.Msg {
	return dashboardLoadedMsg{summary: renderWidgets(time.Now()), tasks: executeTasks()}
}

func initialModel() model {
	return model{
		menus: []string{
			"Today's Overview",
			"Calendar",
			"Journal",
			"Tasks",
			"Daily Review",
			"Statistics",
			"Quit",
		},
		state: stateDashboard,
	}
}

func (m model) Init() tea.Cmd {
	return loadDashboard
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.scroll = 0
		return m, nil
	case dashboardLoadedMsg:
		m.summary, m.taskPreview = msg.summary, msg.tasks
		return m, nil
	case tea.KeyPressMsg:
		key := msg.String()

		// 입력 중에는 q도 일반 문자로 취급한다.
		if key == "ctrl+c" || (key == "q" && m.state != stateInput) {
			return m, tea.Quit
		}
		if key == "pgdown" || key == "pgup" {
			_, limit := m.layout()
			step := max(1, m.height/2)
			if key == "pgup" {
				step = -step
			}
			m.scroll = max(0, min(limit, m.scroll+step))
			return m, nil
		}

		switch m.state {
		case stateDashboard:
			return m.updateDashboard(msg)
		case stateResult:
			return m.updateResult(msg)
		case stateSubMenu:
			return m.updateSubMenu(msg)
		case stateInput:
			return m.updateInput(msg)
		}
	}

	return m, nil
}

// 대시보드 메뉴 업데이트
func (m model) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			m.scroll = 0
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.scroll = 0
			if m.cursor < len(m.menus)-1 {
				m.cursor++
			}
		case "enter":
			m.scroll = 0
			if m.cursor == len(m.menus)-1 {
				return m, tea.Quit
			}
			// 하위 메뉴가 있는 경우
			subItems := getSubMenu(m.cursor)
			if len(subItems) > 0 {
				m.state = stateSubMenu
				m.subMenu = subItems
				m.subCursor = 0
				m.parent = m.cursor
				if m.parent == 3 {
					m.output = executeTasks()
				}
				return m, nil
			}
			// 하위 메뉴가 없는 경우(Today, Statistics) → 결과 표시
			m.output = executeMenu(m.cursor)
			m.state = stateResult
			return m, nil
		}
	}
	return m, nil
}

// 결과 화면 업데이트
func (m model) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.state = stateDashboard
			m.output = ""
			m.scroll = 0
			return m, loadDashboard
		case "esc":
			m.state = stateDashboard
			m.output = ""
			m.scroll = 0
			return m, loadDashboard
		}
	}
	return m, nil
}

// 하위 메뉴 업데이트
func (m model) updateSubMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			m.scroll = 0
			if m.subCursor > 0 {
				m.subCursor--
			}
		case "down", "j":
			m.scroll = 0
			if m.subCursor < len(m.subMenu)-1 {
				m.subCursor++
			}
		case "esc":
			m.state = stateDashboard
			m.subMenu = nil
			m.subCursor = 0
			m.scroll = 0
			return m, loadDashboard
		case "enter":
			m.scroll = 0
			// Back 선택 시 대시보드로 돌아가기
			if m.subCursor == len(m.subMenu)-1 {
				m.state = stateDashboard
				m.subMenu = nil
				m.subCursor = 0
				return m, loadDashboard
			}
			// 입력이 필요한 항목인지 확인
			item := m.subMenu[m.subCursor]
			if needsInput(item) {
				m.state = stateInput
				m.input = ""
				m.inputHint = getInputHint(item)
				m.output = ""
				if item == "Complete Task" || item == "Delete Task" {
					m.output = executeTasks()
				}
				return m, nil
			}
			// 하위 메뉴 선택 실행
			result := executeSubMenu(m.parent, m.subCursor, m.subMenu)
			m.output = result
			m.state = stateResult
			m.subMenu = nil
			return m, nil
		}
	}
	return m, nil
}

// 입력 모드 업데이트
func (m model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.state = stateSubMenu
			m.input = ""
			m.scroll = 0
			return m, nil
		case "enter":
			m.scroll = 0
			// 입력 처리
			if m.input != "" {
				result := processInput(m.parent, m.inputHint, m.input)
				m.output = result
				if m.inputHint == getInputHint("Complete Task") || m.inputHint == getInputHint("Delete Task") {
					m.output += "\n\n" + executeTasks()
				}
				m.state = stateResult
				m.input = ""
				return m, nil
			}
		case "backspace":
			if len(m.input) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.input)
				m.input = m.input[:len(m.input)-size]
			}
		default:
			// 일반 문자 입력 (bubbletea v2: msg.Key().Text 사용)
			if text := msg.Key().Text; text != "" {
				m.input += text
			}
		}
	}
	return m, nil
}

// 입력이 필요한 항목인지 확인
func needsInput(item string) bool {
	switch item {
	case "Add Event", "Write Journal", "Read by Date", "Add Task",
		"Complete Task", "Delete Task", "Write Review":
		return true
	}
	return false
}

// 입력 프롬프트 반환
func getInputHint(item string) string {
	switch item {
	case "Add Event":
		return "Enter event title:"
	case "Write Journal":
		return "Enter journal content:"
	case "Read by Date":
		return "Enter date (YYYY-MM-DD):"
	case "Add Task":
		return "Enter task title:"
	case "Complete Task":
		return "Enter task ID to complete:"
	case "Delete Task":
		return "Enter task ID to delete:"
	case "Write Review":
		return "Enter mood (happy/sad/neutral):"
	}
	return "Enter input:"
}

// 메뉴 인덱스에 따른 하위 메뉴 반환
func getSubMenu(index int) []string {
	switch index {
	case 1:
		// Calendar
		return []string{
			"View Calendar",
			"Add Event",
			"Back",
		}
	case 2:
		// Journal
		return []string{
			"Write Journal",
			"View Journals",
			"Read by Date",
			"Back",
		}
	case 3:
		// Tasks
		return []string{
			"Add Task",
			"View Tasks",
			"Complete Task",
			"Delete Task",
			"Back",
		}
	case 4:
		// Daily Review
		return []string{
			"Write Review",
			"View Reviews",
			"Back",
		}
	default:
		return nil
	}
}

// 하위 메뉴 선택 실행
func executeSubMenu(parent, subIndex int, subMenu []string) string {
	item := subMenu[subIndex]

	switch parent {
	case 1:
		// Calendar
		switch item {
		case "View Calendar":
			return executeCalendar()
		case "Add Event":
			return "Enter event title: (type and press Enter)"
		}
	case 2:
		// Journal
		switch item {
		case "Write Journal":
			return "Enter journal content: (type and press Enter)"
		case "View Journals":
			return executeJournal()
		case "Read by Date":
			return "Enter date (YYYY-MM-DD): (type and press Enter)"
		}
	case 3:
		// Tasks
		switch item {
		case "Add Task":
			return "Enter task title: (type and press Enter)"
		case "View Tasks":
			return executeTasks()
		case "Complete Task":
			return "Enter task ID to complete: (type and press Enter)"
		case "Delete Task":
			return "Enter task ID to delete: (type and press Enter)"
		}
	case 4:
		// Daily Review
		switch item {
		case "Write Review":
			return "Enter mood (happy/sad/neutral): (type and press Enter)"
		case "View Reviews":
			return executeReview()
		}
	}
	return ""
}

// 입력 처리 - inputHint로 어떤 입력인지 판단
func processInput(parent int, inputHint string, input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "Input cannot be empty."
	}

	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer db.Close()

	switch inputHint {
	case "Enter event title:":
		return fmt.Sprintf("Event added: %s\n(Note: Full event scheduling coming soon)", input)

	case "Enter journal content:":
		repo := journal.NewRepository(db)
		now := time.Now()
		date := now.Format("2006-01-02")
		err := repo.Create(date, input)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("Journal saved for %s!", date)

	case "Enter date (YYYY-MM-DD):":
		repo := journal.NewRepository(db)
		j, err := repo.GetByDate(input)
		if journal.IsNotFound(err) {
			return fmt.Sprintf("No journal found for %s", input)
		}
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("%s\n\n%s", j.Date, j.Content)

	case "Enter task title:":
		repo := task.NewRepository(db)
		err := repo.Create(input)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("Task added: %s", input)

	case "Enter task ID to complete:", "Enter task ID to delete:":
		id, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Sprintf("Invalid ID: %s", input)
		}
		repo := task.NewRepository(db)
		tasks, err := repo.ListByDate(time.Now().Format("2006-01-02"))
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		found := false
		for _, t := range tasks {
			if t.ID == id {
				found = true
				break
			}
		}
		if !found {
			return "Please enter an ID from today's tasks."
		}
		if inputHint == "Enter task ID to complete:" {
			if err := repo.Complete(id); err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return fmt.Sprintf("Task %d completed!", id)
		}
		err = repo.Delete(id)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("Task %d deleted!", id)

	case "Enter mood (happy/sad/neutral):":
		repo := review.NewRepository(db)
		now := time.Now()
		date := now.Format("2006-01-02")
		rating := 3
		err := repo.Create(date, input, rating, "")
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("Review saved for %s! Mood: %s", date, input)
	}

	return "Feature coming soon!"
}

// 메뉴 인덱스에 따라 해당 기능 실행하고 결과 문자열 반환
func executeMenu(index int) string {
	var sb strings.Builder

	switch index {
	case 0:
		sb.WriteString(executeToday())
	case 1:
		sb.WriteString(executeCalendar())
	case 2:
		sb.WriteString(executeJournal())
	case 3:
		sb.WriteString(executeTasks())
	case 4:
		sb.WriteString(executeReview())
	case 5:
		sb.WriteString(executeStats())
	}

	return sb.String()
}

func executeToday() string {
	var sb strings.Builder
	now := time.Now()
	date := now.Format("2006-01-02")

	sb.WriteString("=== Today's Overview ===\n")
	sb.WriteString(fmt.Sprintf("Date: %s %s\n\n", date, now.Format("Monday")))

	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer db.Close()

	// 일기
	journalRepo := journal.NewRepository(db)
	j, err := journalRepo.GetByDate(date)
	if err == nil {
		sb.WriteString(fmt.Sprintf("Journal: %s\n", j.Content))
	} else if journal.IsNotFound(err) {
		sb.WriteString("Journal: Not written\n")
	} else {
		return fmt.Sprintf("Error loading journal: %v", err)
	}

	// 작업
	taskRepo := task.NewRepository(db)
	pendingTasks, err := taskRepo.List("pending")
	if err != nil {
		return fmt.Sprintf("Error loading tasks: %v", err)
	}
	if len(pendingTasks) > 0 {
		sb.WriteString(fmt.Sprintf("\nPending Tasks (%d):\n", len(pendingTasks)))
		for _, t := range pendingTasks {
			sb.WriteString(fmt.Sprintf("  [%d] %s\n", t.ID, t.Title))
		}
	} else {
		sb.WriteString("\nPending Tasks: None\n")
	}

	// 리뷰
	reviewRepo := review.NewRepository(db)
	rev, err := reviewRepo.GetByDate(date)
	if err == nil {
		sb.WriteString(fmt.Sprintf("\nReview: %s %d/5\n", rev.Mood, rev.Rating))
		if rev.Content != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", rev.Content))
		}
	} else if review.IsNotFound(err) {
		sb.WriteString("\nReview: Not written\n")
	} else {
		return fmt.Sprintf("Error loading review: %v", err)
	}

	return sb.String()
}

func executeCalendar() string {
	var sb strings.Builder
	now := time.Now()
	year, month, _ := now.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())

	sb.WriteString(fmt.Sprintf("\n      %s %d\n", month, year))
	sb.WriteString("Su Mo Tu We Th Fr Sa\n")

	for i := 0; i < int(first.Weekday()); i++ {
		sb.WriteString("   ")
	}

	last := first.AddDate(0, 1, -1).Day()
	for day := 1; day <= last; day++ {
		sb.WriteString(fmt.Sprintf("%2d ", day))
		if (int(first.Weekday())+day)%7 == 0 {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

func executeJournal() string {
	var sb strings.Builder
	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer db.Close()

	repo := journal.NewRepository(db)
	journals, err := repo.List()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if len(journals) == 0 {
		sb.WriteString("No journals written yet.\n")
		return sb.String()
	}

	sb.WriteString("=== Recent Journals ===\n")
	count := 0
	for _, j := range journals {
		sb.WriteString(fmt.Sprintf("%s - %s\n", j.Date, j.Content))
		count++
		if count >= 10 {
			sb.WriteString("  ... (showing last 10)\n")
			break
		}
	}

	return sb.String()
}

func executeTasks() string {
	var sb strings.Builder
	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer db.Close()

	repo := task.NewRepository(db)
	date := time.Now().Format("2006-01-02")
	tasks, err := repo.ListByDate(date)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	sb.WriteString(fmt.Sprintf("=== Today's Tasks (%s) ===\n", date))
	if len(tasks) == 0 {
		sb.WriteString("No tasks for today.\n")
		return sb.String()
	}
	completed := 0
	for _, t := range tasks {
		if t.Status == "completed" {
			completed++
		}
	}
	sb.WriteString(fmt.Sprintf("%d pending  ·  %d completed\n\n", len(tasks)-completed, completed))

	for _, t := range tasks {
		status := "⬜"
		if t.Status == "completed" {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("%s [%d] %s\n", status, t.ID, t.Title))
	}

	return sb.String()
}

func executeReview() string {
	var sb strings.Builder
	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer db.Close()

	repo := review.NewRepository(db)
	reviews, err := repo.List()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if len(reviews) == 0 {
		sb.WriteString("No reviews written yet.\n")
		return sb.String()
	}

	sb.WriteString("=== Recent Reviews ===\n")
	count := 0
	for _, rev := range reviews {
		stars := statistics.FormatStars(float64(rev.Rating))
		sb.WriteString(fmt.Sprintf("%s %s [%s]\n", rev.Date, rev.Mood, stars))
		if rev.Content != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", rev.Content))
		}
		count++
		if count >= 10 {
			sb.WriteString("  ... (showing last 10)\n")
			break
		}
	}

	return sb.String()
}

func executeStats() string {
	var sb strings.Builder
	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer db.Close()

	svc := statistics.NewService(db)

	sb.WriteString("=== Statistics ===\n\n")

	// 일기 통계
	journalStats, err := svc.GetJournalStats()
	if err != nil {
		return fmt.Sprintf("Error loading journal statistics: %v", err)
	}
	sb.WriteString("Journal\n")
	sb.WriteString(fmt.Sprintf("  Total: %d\n", journalStats.TotalCount))
	sb.WriteString(fmt.Sprintf("  This month: %d\n", journalStats.ThisMonth))
	sb.WriteString(fmt.Sprintf("  This week: %d\n", journalStats.ThisWeek))
	sb.WriteString(fmt.Sprintf("  Current streak: %d days\n", journalStats.CurrentStreak))
	sb.WriteString(fmt.Sprintf("  Longest streak: %d days\n\n", journalStats.LongestStreak))

	// 작업 통계
	taskStats, err := svc.GetTaskStats()
	if err != nil {
		return fmt.Sprintf("Error loading task statistics: %v", err)
	}
	sb.WriteString("Tasks\n")
	sb.WriteString(fmt.Sprintf("  Total: %d\n", taskStats.Total))
	sb.WriteString(fmt.Sprintf("  Pending: %d\n", taskStats.Pending))
	sb.WriteString(fmt.Sprintf("  Completed: %d\n\n", taskStats.Completed))

	// 리뷰 통계
	reviewStats, err := svc.GetReviewStats()
	if err != nil {
		return fmt.Sprintf("Error loading review statistics: %v", err)
	}
	sb.WriteString("Reviews\n")
	sb.WriteString(fmt.Sprintf("  Total: %d\n", reviewStats.TotalCount))
	if reviewStats.TotalCount > 0 {
		sb.WriteString(fmt.Sprintf("  Average rating: %.1f/5 %s\n",
			reviewStats.AvgRating,
			statistics.FormatStars(reviewStats.AvgRating)))
	}

	return sb.String()
}

// 위젯 렌더링 - 대시보드에 오늘 요약 표시
func renderWidgets(now time.Time) string {
	date := now.Format("2006-01-02")

	db, err := storage.OpenDB()
	if err != nil {
		return fmt.Sprintf("Error: %v\n", err)
	}
	defer db.Close()

	// 일기 확인
	journalRepo := journal.NewRepository(db)
	_, errJournal := journalRepo.GetByDate(date)
	if errJournal != nil && !journal.IsNotFound(errJournal) {
		return fmt.Sprintf("Error loading journal: %v\n", errJournal)
	}
	hasJournal := (errJournal == nil)

	// 작업 확인
	taskRepo := task.NewRepository(db)
	tasks, err := taskRepo.ListByDate(date)
	if err != nil {
		return fmt.Sprintf("Error loading tasks: %v\n", err)
	}
	pendingCount := 0
	for _, t := range tasks {
		if t.Status == "pending" {
			pendingCount++
		}
	}

	// 리뷰 확인
	reviewRepo := review.NewRepository(db)
	_, errReview := reviewRepo.GetByDate(date)
	if errReview != nil && !review.IsNotFound(errReview) {
		return fmt.Sprintf("Error loading review: %v\n", errReview)
	}
	hasReview := (errReview == nil)

	// 통계 확인
	svc := statistics.NewService(db)
	journalStats, err := svc.GetJournalStats()
	if err != nil {
		return fmt.Sprintf("Error loading journal statistics: %v\n", err)
	}
	streak := journalStats.CurrentStreak

	// 요약 데이터만 반환하고 화면 배치는 View에서 처리한다.
	journalStatus, reviewStatus := "Not written", "Not written"
	if hasJournal {
		journalStatus = "Written"
	}
	if hasReview {
		reviewStatus = "Written"
	}
	return fmt.Sprintf("JOURNAL  %s\nREVIEW   %s\nTASKS    %d pending\nSTREAK   %d days", journalStatus, reviewStatus, pendingCount, streak)
}
func RunDashboard() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}
