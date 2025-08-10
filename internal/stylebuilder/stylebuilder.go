package stylebuilder

import (
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

type StyleBuilder struct {
	Style   lipgloss.Style
	builder strings.Builder
}

func NewStyleBuilder(style lipgloss.Style) *StyleBuilder {
	return &StyleBuilder{
		Style: style,
	}
}

func (sb *StyleBuilder) WriteString(s string) {
	sb.builder.WriteString(sb.Style.Render(s))
}

func (sb *StyleBuilder) WriteStringWithStyle(s string, style lipgloss.Style) {
	sb.builder.WriteString(style.Inherit(sb.Style).Render(s))
}

func (sb *StyleBuilder) WriteStyledString(s string) {
	sb.builder.WriteString(s)
}

func (sb *StyleBuilder) String() string {
	return sb.builder.String()
}

func (sb *StyleBuilder) Reset() {
	sb.builder.Reset()
}
