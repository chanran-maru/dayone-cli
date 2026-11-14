package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	accent  = lipgloss.Color("#5EEAD4")
	muted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	bright  = lipgloss.NewStyle().Foreground(lipgloss.Color("#E2E8F0"))
	heading = bright.Bold(true)
	border  = lipgloss.Color("#475569")
)

func (m model) View() tea.View {
	content, _ := m.layout()
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// 레이아웃은 저장소에 접근하지 않는다. 키 입력과 창 크기 변경 시에는 캐시된 데이터를 그린다.
func (m model) layout() (string, int) {
	w, h := m.width, m.height
	if w == 0 {
		w = 100
	}
	if h == 0 {
		h = 30
	}
	w = min(w, 112)
	if w < 28 || h < 18 {
		return ansi.Truncate("DayOne · Enlarge terminal", max(1, w), ""), 0
	}
	inner := w - 2
	title, description := m.pageHeading()
	brand := lipgloss.NewStyle().Bold(true).Foreground(accent).Render("DayOne") + muted.Render("  /  DAILY WORKSPACE")
	date := muted.Render(time.Now().Format("Mon, 02 Jan 2006"))
	if inner < 60 {
		date = muted.Render(time.Now().Format("01.02 Mon"))
	}
	gap := max(1, inner-lipgloss.Width(brand)-lipgloss.Width(date))
	header := ansi.Truncate(brand+strings.Repeat(" ", gap)+date, inner, "") + "\n" + ansi.Truncate(muted.Render("Workspace  /  ")+bright.Render(title), inner, "…")
	wide := w >= 80
	panelWidth := inner
	if wide {
		panelWidth -= 28
	}
	contentWidth := panelWidth - 6
	bodyHeight := h - 6
	contentHeight := bodyHeight - 2
	text := m.pageContent(contentWidth)
	if !wide && (m.state == stateDashboard || m.state == stateSubMenu) {
		items, selected := m.menus, m.cursor
		if m.state == stateSubMenu {
			items, selected = m.subMenu, m.subCursor
		}
		text = lipgloss.NewStyle().Foreground(accent).Bold(true).Render(fmt.Sprintf("› %s  (%d/%d)", items[selected], selected+1, len(items))) + "\n\n" + text
	}
	form := ""
	formHeight := 0
	submit := "save"
	if m.inputHint == getInputHint("Complete Task") || m.inputHint == getInputHint("Delete Task") {
		submit = "confirm"
	}
	if m.state == stateInput {
		field := ansi.TruncateLeft(m.input, max(0, lipgloss.Width(m.input)-(contentWidth-5)), "…")
		form = "\n" + heading.Render(ansi.Truncate(m.inputHint, contentWidth, "…")) + "\n" +
			lipgloss.NewStyle().Foreground(accent).Render("› "+field+"▏") + "\n" + muted.Render(ansi.Truncate("Enter to "+submit+" · Esc to cancel", contentWidth, "…"))
		formHeight = 3
	}
	visibleHeight := max(1, contentHeight-3-formHeight)
	wrapped := ansi.Wrap(text, contentWidth, "")
	lines := strings.Split(strings.TrimRight(wrapped, "\n"), "\n")
	limit := max(0, len(lines)-visibleHeight)
	offset := min(m.scroll, limit)
	visible := strings.Join(lines[offset:min(len(lines), offset+visibleHeight)], "\n")
	visible = lipgloss.NewStyle().Width(contentWidth).Height(visibleHeight).Render(visible)
	pageTitle := heading.Render(title) + "\n" + muted.Render(ansi.Truncate(description, contentWidth, "…")) + "\n\n"
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).
		Padding(0, 2).Width(panelWidth).Height(bodyHeight)
	main := panel.Render(pageTitle + visible + form)
	body := main
	if wide {
		nav := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).
			Padding(0, 1).Width(26).Height(bodyHeight).
			Render(m.navigation(22))
		body = lipgloss.JoinHorizontal(lipgloss.Top, nav, "  ", main)
	}
	keys := "↑↓ move  enter select  q quit"
	if m.state == stateResult {
		keys = "enter / esc back  q quit"
	}
	if m.state == stateSubMenu {
		keys = "↑↓ move  enter select  esc back  q quit"
	}
	if m.state == stateInput {
		keys = "enter " + submit + "  esc cancel  ctrl+c quit"
	}
	if limit > 0 {
		keys = fmt.Sprintf("pgup/pgdn %d/%d  ·  ", offset+1, limit+1) + keys
	}
	footer := muted.Render(ansi.Truncate(keys, inner, "…"))
	return lipgloss.NewStyle().Padding(1, 1).Render(header + "\n\n" + body + "\n" + footer), limit
}

func (m model) pageHeading() (string, string) {
	switch m.state {
	case stateInput:
		if m.subCursor < len(m.subMenu) {
			item := m.subMenu[m.subCursor]
			if item == "Complete Task" {
				return item, "Enter the ID of a pending task below."
			}
			if item == "Delete Task" {
				return item, "Enter the ID of the task to remove."
			}
			return item, "Type below, then press Enter to save."
		}
		return "New entry", "Type below, then press Enter to save."
	case stateSubMenu:
		return m.menus[m.parent], "Choose an action from the menu."
	case stateResult:
		return "Your workspace", "Results and the latest status of your records."
	default:
		return "Today", "Your tasks and daily check-in, in one place."
	}
}

func (m model) navigation(width int) string {
	items, selected := m.menus, m.cursor
	label := "WORKSPACE"
	if m.state == stateSubMenu || m.state == stateInput {
		items, selected = m.subMenu, m.subCursor
		label = strings.ToUpper(m.menus[m.parent])
	}
	var rows []string
	rows = append(rows, "", muted.Bold(true).Render(label), "")
	for i, item := range items {
		line := "  " + item
		style := bright
		if i == selected {
			line = "› " + item
			style = lipgloss.NewStyle().Bold(true).Foreground(accent).Background(lipgloss.Color("#183438"))
		}
		rows = append(rows, style.Render(ansi.Truncate(line, width, "…")))
	}
	return strings.Join(rows, "\n")
}

func (m model) pageContent(width int) string {
	switch m.state {
	case stateInput:
		if m.output != "" {
			return formatOutput(m.output)
		}
		return muted.Render("Your entry will be saved locally.")
	case stateResult:
		return formatOutput(m.output)
	case stateSubMenu:
		if m.parent == 3 {
			return formatOutput(m.output)
		}
		switch m.parent {
		case 1:
			return formatOutput(executeCalendar())
		case 2:
			return heading.Render("Keep a record of your day.") + "\n\n" + muted.Render("Write a new entry, browse your journal,\nor revisit a particular date.")
		case 4:
			return heading.Render("How was your day?") + "\n\n" + muted.Render("Capture your mood and look back\non the moments that mattered.")
		}
	default:
		summary := m.summary
		if summary == "" {
			summary = "Loading today's overview…"
		}
		tasks := m.taskPreview
		if tasks == "" {
			tasks = "Your tasks will appear here."
		}
		return summaryCards(summary, width) + "\n\n" + formatOutput(tasks)
	}
	return ""
}

func summaryCards(summary string, width int) string {
	lines := strings.Split(summary, "\n")
	if len(lines) != 4 || width < 58 {
		return muted.Render(summary)
	}
	cardWidth := (width - 3) / 4
	cards := make([]string, 0, 4)
	for _, line := range lines {
		label, value, _ := strings.Cut(line, "  ")
		cards = append(cards, lipgloss.NewStyle().Width(cardWidth).Render(
			muted.Render(label)+"\n"+heading.Render(strings.TrimSpace(value))))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cards[0], " ", cards[1], " ", cards[2], " ", cards[3])
}

func formatOutput(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "=== "):
			lines[i] = heading.Render(strings.Trim(line, "= "))
		case strings.HasPrefix(line, "⬜ "):
			lines[i] = bright.Render("[ ] " + strings.TrimPrefix(line, "⬜ "))
		case strings.HasPrefix(line, "✅ "):
			lines[i] = lipgloss.NewStyle().Foreground(accent).Render("[x] " + strings.TrimPrefix(line, "✅ "))
		case strings.HasPrefix(line, "Error"), strings.HasPrefix(line, "Invalid"), strings.HasPrefix(line, "Please enter"):
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#FDA4AF")).Render(line)
		case strings.HasPrefix(line, "No "):
			lines[i] = muted.Render(line)
		default:
			lines[i] = bright.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}
