package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"DayOneCLI/internal/journal"
	"DayOneCLI/internal/review"
	"DayOneCLI/internal/storage"
	"DayOneCLI/internal/statistics"
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
	cursor    int
	menus     []string
	state     string // "dashboard", "result", "submenu", "input"
	output    string
	subMenu   []string // 하위 메뉴 항목
	subCursor int       // 하위 메뉴 커서
	parent    int       // 상위 메뉴 인덱스
	input     string    // 입력 모드에서 입력 중인 텍스트
	inputHint string    // 입력 프롬프트
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
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		// Ctrl+C / q는 항상 종료
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
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
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menus)-1 {
				m.cursor++
			}
		case "enter":
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
			return m, nil
		case "esc":
			m.state = stateDashboard
			m.output = ""
			return m, nil
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
			if m.subCursor > 0 {
				m.subCursor--
			}
		case "down", "j":
			if m.subCursor < len(m.subMenu)-1 {
				m.subCursor++
			}
		case "esc":
			m.state = stateDashboard
			m.subMenu = nil
			m.subCursor = 0
			return m, nil
		case "enter":
			// Back 선택 시 대시보드로 돌아가기
			if m.subCursor == len(m.subMenu)-1 {
				m.state = stateDashboard
				m.subMenu = nil
				m.subCursor = 0
				return m, nil
			}
			// 입력이 필요한 항목인지 확인
			item := m.subMenu[m.subCursor]
			if needsInput(item) {
				m.state = stateInput
				m.input = ""
				m.inputHint = getInputHint(item)
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
			return m, nil
		case "enter":
			// 입력 처리
			if m.input != "" {
				result := processInput(m.parent, m.inputHint, m.input)
				m.output = result
				m.state = stateResult
				m.input = ""
				return m, nil
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
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
		if err != nil {
			return fmt.Sprintf("No journal found for %s", input)
		}
		return fmt.Sprintf("%s\n\n%s", j.Date, j.Content)

	case "Enter task title:":
		repo := task.NewRepository(db)
		err := repo.Create(input)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("Task added: %s", input)

	case "Enter task ID to complete:":
		id, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Sprintf("Invalid ID: %s", input)
		}
		repo := task.NewRepository(db)
		err = repo.Complete(id)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("Task %d completed!", id)

	case "Enter task ID to delete:":
		id, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Sprintf("Invalid ID: %s", input)
		}
		repo := task.NewRepository(db)
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
	} else {
		sb.WriteString("Journal: Not written\n")
	}

	// 작업
	taskRepo := task.NewRepository(db)
	pendingTasks, err := taskRepo.List("pending")
	if err == nil && len(pendingTasks) > 0 {
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
	} else {
		sb.WriteString("\nReview: Not written\n")
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
		if (int(first.Weekday())+day-1)%7 == 0 {
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
	tasks, err := repo.ListAll()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if len(tasks) == 0 {
		sb.WriteString("No tasks yet.\n")
		return sb.String()
	}

	sb.WriteString("=== Tasks ===\n")
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
		stars := strings.Repeat("★", rev.Rating) + strings.Repeat("☆", 5-rev.Rating)
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
	if err == nil {
		sb.WriteString("Journal\n")
		sb.WriteString(fmt.Sprintf("  Total: %d\n", journalStats.TotalCount))
		sb.WriteString(fmt.Sprintf("  This month: %d\n", journalStats.ThisMonth))
		sb.WriteString(fmt.Sprintf("  This week: %d\n", journalStats.ThisWeek))
		sb.WriteString(fmt.Sprintf("  Current streak: %d days\n", journalStats.CurrentStreak))
		sb.WriteString(fmt.Sprintf("  Longest streak: %d days\n\n", journalStats.LongestStreak))
	}

	// 작업 통계
	taskStats, err := svc.GetTaskStats()
	if err == nil {
		sb.WriteString("Tasks\n")
		sb.WriteString(fmt.Sprintf("  Total: %d\n", taskStats.Total))
		sb.WriteString(fmt.Sprintf("  Pending: %d\n", taskStats.Pending))
		sb.WriteString(fmt.Sprintf("  Completed: %d\n\n", taskStats.Completed))
	}

	// 리뷰 통계
	reviewStats, err := svc.GetReviewStats()
	if err == nil {
		sb.WriteString("Reviews\n")
		sb.WriteString(fmt.Sprintf("  Total: %d\n", reviewStats.TotalCount))
		if reviewStats.TotalCount > 0 {
			sb.WriteString(fmt.Sprintf("  Average rating: %.1f/5 %s\n",
				reviewStats.AvgRating,
				statistics.FormatStars(reviewStats.AvgRating)))
		}
	}

	return sb.String()
}

func (m model) View() tea.View {
	// 색상 테마
	primaryColor := lipgloss.Color("#60A5FA")   // 파란색
	successColor := lipgloss.Color("#34D399")   // 초록색
	mutedColor := lipgloss.Color("#6B7280")     // 회색
	brightColor := lipgloss.Color("#F3F4F6")    // 밝은 흰색
	borderColor := lipgloss.Color("#374151")    // 어두운 회색(테두리)

	// 스타일 정의
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Padding(0, 1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1F2937")).
		Background(successColor).
		Padding(0, 1).
		Margin(0, 0)

	menuItemStyle := lipgloss.NewStyle().
		Foreground(brightColor).
		Padding(0, 1)

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Margin(1, 0)

	widgetBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Margin(0, 0)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4B5563"))

	date := time.Now().Format("Monday, January 2, 2006")
	now := time.Now()

	// 메뉴 아이콘
	menuIcons := []string{
		"📋", // Today's Overview
		"📅", // Calendar
		"📝", // Journal
		"✅", // Tasks
		"📊", // Daily Review
		"📈", // Statistics
		"🚪", // Quit
	}

	var s string

	switch m.state {
	case stateResult:
		// 결과 화면 표시
		header := titleStyle.Render("📖 DayOne CLI") + "\n"
		header += subtitleStyle.Render("Your personal daily dashboard") + "\n"
		header += boxStyle.Render(" " + date + " ") + "\n\n"

		content := boxStyle.Render("\n" + m.output + "\n")

		footer := "\n" + hintStyle.Render("⏎ Enter: Back to menu    ⎋ q: Quit")

		s = header + content + footer

	case stateSubMenu:
		// 하위 메뉴 표시
		parentName := m.menus[m.parent]
		parentIcon := menuIcons[m.parent]

		header := titleStyle.Render("📖 DayOne CLI") + "\n"
		header += subtitleStyle.Render("Your personal daily dashboard") + "\n"
		header += boxStyle.Render(" "+parentIcon+" "+parentName+" ") + "\n\n"

		// 메뉴 항목 렌더링
		var menuLines string
		for i, item := range m.subMenu {
			if i == m.subCursor {
				menuLines += selectedStyle.Render("  " + item + "  ") + "\n"
			} else {
				menuLines += menuItemStyle.Render("  " + item + "  ") + "\n"
			}
		}

		menuBox := boxStyle.Render(menuLines)

		footer := "\n" + hintStyle.Render("↑/↓ Navigate    ⏎ Select    ⎋ Back    ⎋ q: Quit")

		s = header + menuBox + footer

	case stateInput:
		// 입력 화면 표시
		parentName := m.menus[m.parent]
		parentIcon := menuIcons[m.parent]

		header := titleStyle.Render("📖 DayOne CLI") + "\n"
		header += subtitleStyle.Render("Your personal daily dashboard") + "\n"
		header += boxStyle.Render(" "+parentIcon+" "+parentName+" ") + "\n\n"

		inputBox := widgetBoxStyle.Render(
			"\n"+m.inputHint+"\n" +
				"▸ "+m.input+"█",
		)

		footer := "\n\n" + hintStyle.Render("⏎ Enter: Submit    ⎋ Esc: Cancel    ⎋ q: Quit")

		s = header + inputBox + footer

	default:
		// 대시보드 메뉴 표시
		// 헤더
		header := titleStyle.Render("📖 DayOne CLI") + "\n"
		header += subtitleStyle.Render("Your personal daily dashboard") + "\n"
		header += boxStyle.Render(" "+date+" ") + "\n\n"

		// 위젯 영역 - 오늘 요약
		var widgets string
		widgets += renderWidgets(now)

		// 메뉴 영역
		var menuLines string
		for i, item := range m.menus {
			icon := menuIcons[i]
			if i == m.cursor {
				menuLines += selectedStyle.Render(" "+icon+" "+item+" ") + "\n"
			} else {
				menuLines += menuItemStyle.Render(" "+icon+" "+item+" ") + "\n"
			}
		}

		menuBox := boxStyle.Render(menuLines)

		footer := "\n" + hintStyle.Render("↑/↓ Navigate    ⏎ Select    ⎋ q: Quit")

		s = header + widgets + menuBox + footer
	}

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

// 위젯 렌더링 - 대시보드에 오늘 요약 표시
func renderWidgets(now time.Time) string {
	date := now.Format("2006-01-02")

	db, err := storage.OpenDB()
	if err != nil {
		return ""
	}
	defer db.Close()

	// 일기 확인
	journalRepo := journal.NewRepository(db)
	_, errJournal := journalRepo.GetByDate(date)
	hasJournal := (errJournal == nil)

	// 작업 확인
	taskRepo := task.NewRepository(db)
	pendingTasks, _ := taskRepo.List("pending")
	pendingCount := len(pendingTasks)

	// 리뷰 확인
	reviewRepo := review.NewRepository(db)
	_, errReview := reviewRepo.GetByDate(date)
	hasReview := (errReview == nil)

	// 통계 확인
	svc := statistics.NewService(db)
	journalStats, _ := svc.GetJournalStats()
	streak := journalStats.CurrentStreak

	// 위젯 1: 오늘 상태
	var todayStatus string
	if hasJournal {
		todayStatus += "📝 Journal ✓\n"
	} else {
		todayStatus += "📝 Journal ✗\n"
	}
	if hasReview {
		todayStatus += "📊 Review ✓\n"
	} else {
		todayStatus += "📊 Review ✗\n"
	}
	if pendingCount > 0 {
		todayStatus += fmt.Sprintf("✅ Tasks: %d pending", pendingCount)
	} else {
		todayStatus += "✅ Tasks: All done!"
	}

	// 위젯 2: 연속 일수
	streakText := fmt.Sprintf("🔥 %d day streak", streak)

	// 위젯 박스 스타일
	widgetStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1, 2).
		Width(20)

	widget1 := widgetStyle.Render(todayStatus)
	widget2 := widgetStyle.Render(streakText)

	// 두 위젯을 나란히 배치
	return lipgloss.JoinHorizontal(lipgloss.Top, widget1, "  ", widget2) + "\n\n"
}

func RunDashboard() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}
