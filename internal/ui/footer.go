package ui

import "github.com/charmbracelet/lipgloss"

type footerModel struct {
	width int
}

func newFooterModel() footerModel {
	return footerModel{}
}

func (m footerModel) height() int {
	return 1
}

func (m *footerModel) SetWidth(width int) {
	m.width = width - 2
}

func (m footerModel) View() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFC0")).
		Border(lipgloss.NormalBorder(), true).
		Inherit(borderColorStyle).
		Faint(true).
		Width(m.width).
		Align(lipgloss.Center)

	text := "'tab' toggle pattern or sample view | 'spacebar' Start/Stop Song | 'q' Quit"
	return style.Render(text)
}
