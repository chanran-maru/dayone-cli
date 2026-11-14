package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func previewModel(state string) model {
	m := initialModel()
	m.state, m.parent, m.subMenu, m.subCursor = state, 3, getSubMenu(3), 2
	m.inputHint = getInputHint("Complete Task")
	m.summary = "JOURNAL  Written\nREVIEW   Not written\nTASKS    2 pending\nSTREAK   7 days"
	m.output = "=== Today's Tasks (2026-11-14) ===\n2 pending  ·  1 completed\n\n⬜ [12] 프로젝트 UI 정리하기\n⬜ [13] Read a chapter\n✅ [14] Morning walk"
	m.taskPreview = m.output
	return m
}

func TestLayoutFitsTerminal(t *testing.T) {
	for _, size := range [][2]int{{20, 10}, {28, 18}, {36, 24}, {64, 24}, {80, 24}, {100, 30}, {140, 40}} {
		for _, state := range []string{stateDashboard, stateSubMenu, stateInput, stateResult} {
			t.Run(fmt.Sprintf("%dx%d/%s", size[0], size[1], state), func(t *testing.T) {
				m := previewModel(state)
				m.width, m.height = size[0], size[1]
				m.input = strings.Repeat("한글과 English 입력 ", 20)
				m.output += "\n" + strings.Repeat("A very long task 한글을 포함한 긴 작업 설명 ", 40)
				view := m.View().Content
				if width, height := lipgloss.Size(view); width > m.width || height > m.height {
					t.Fatalf("layout size = %dx%d, terminal = %dx%d\n%s", width, height, m.width, m.height, ansi.Strip(view))
				}
			})
		}
	}
}

func TestScrollingKeepsInputVisible(t *testing.T) {
	m := previewModel(stateInput)
	m.width, m.height = 80, 24
	for i := 0; i < 60; i++ {
		m.output += fmt.Sprintf("\n⬜ [%d] More tasks", i+100)
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	m = next.(model)
	if m.scroll == 0 {
		t.Fatal("page down did not scroll")
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, m.inputHint) || !strings.Contains(view, "ctrl+c quit") {
		t.Fatalf("scrolling hid the input or key hints:\n%s", view)
	}
	for i := 0; i < 20; i++ {
		next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
		m = next.(model)
	}
	_, limit := m.layout()
	if m.scroll != limit {
		t.Fatal("scroll did not stop at end of content")
	}
	for i := 0; i < 20; i++ {
		next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
		m = next.(model)
	}
	if m.scroll != 0 {
		t.Fatal("page up did not return to start")
	}
	next, _ = m.Update(tea.WindowSizeMsg{Width: 48, Height: 24})
	if next.(model).width != 48 {
		t.Fatal("resize was not applied")
	}
}
