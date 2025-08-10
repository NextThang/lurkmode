package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type header struct {
	content string
	style   lipgloss.Style
}

func newHeader(content string) header {
	return header{
		content: content,
		style: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("#6441a5")).
			Align(lipgloss.Center),
	}
}

func (h header) Update(msg tea.Msg) (header, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.style = h.style.Width(msg.Width)
	}
	return h, nil
}

func (h header) View() string {
	return h.style.Render(h.content)
}
