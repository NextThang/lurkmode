package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type footer struct {
	content string
	style   lipgloss.Style
}

func newFooter() footer {
	return footer{
		content: "  ↑/↓: Navigate • t: Toogle timestamp • q: Quit",
		style:   lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

func (f footer) Update(msg tea.Msg) (footer, tea.Cmd) {
	return f, nil
}

func (f footer) View() string {
	return f.style.Render(f.content)
}
