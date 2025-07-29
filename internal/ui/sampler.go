package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jesseward/impulse/pkg/module"
)

type samplerModel struct {
	table  table.Model
	module module.Module
	width  int
	height int
}

func newSamplerModel(m module.Module) samplerModel {
	columns := []table.Column{
		{Title: "", Width: 3},
		{Title: "Name", Width: 25},
		{Title: "Length", Width: 8},
		{Title: "Vol", Width: 5},
		{Title: "Loop Start", Width: 12},
		{Title: "Loop Len", Width: 9},
	}

	var rows []table.Row
	for i, sample := range m.Samples() {
		sampleName := strings.TrimRight(string(sample.Name()[:]), "\x00")
		loopLength := 0
		if sample.LoopLength() >= 2 {
			loopLength = int(sample.LoopLength())
		}
		row := table.Row{
			fmt.Sprintf("%02d", i+1),
			sampleName,
			fmt.Sprintf("%d", sample.Length()*2),
			fmt.Sprintf("%d", sample.Volume()),
			fmt.Sprintf("%d", sample.LoopStart()*2),
			fmt.Sprintf("%d", loopLength),
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("27")).
		Bold(false)
	t.SetStyles(s)

	return samplerModel{table: t, module: m}
}

func (m samplerModel) Init() tea.Cmd {
	return nil
}

func (m samplerModel) Update(msg tea.Msg) (samplerModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m samplerModel) View() string {

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Inherit(borderColorStyle).
		Width(m.width - 2).
		Height(m.height)
	return style.Render(m.table.View())
}
