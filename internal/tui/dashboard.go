package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	cursor int
	menus  []string
}

func initialModel() model {
	return model{
		menus: []string{
			"Today's Overview",
			"Calendar",
			"Journal",
			"Tasks",
			"Daily Review",
			"Quit",
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

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
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C9FFF"))

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#86EFAC"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	date := time.Now().Format("Monday, January 2, 2006")

	s := titleStyle.Render("DAYONE CLI") + "\n"
	s += mutedStyle.Render("Your personal daily dashboard") + "\n\n"
	s += date + "\n\n"
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	for i, item := range m.menus {
		if i == m.cursor {
			s += selectedStyle.Render("> " + item) + "\n"
		} else {
			s += "  " + item + "\n"
		}
	}

	s += "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	s += mutedStyle.Render("↑/↓: Navigate  Enter: Select  q: Quit")
	s += fmt.Sprintf("\n\nSelected: %s", m.menus[m.cursor])

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func RunDashboard() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}
