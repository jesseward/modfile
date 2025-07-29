package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jesseward/impulse/internal/player"
	"github.com/jesseward/impulse/pkg/module"
)

var (
	noteStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	instrumentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	effectStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("21"))
)

const (
	trackerHeight      = 24
	playheadDisplayRow = trackerHeight/2 - 1
	channelWidth       = 14 // Approximate width for one channel column
	rowNumWidth        = 3  // Width for the row number column
)

type trackerModel struct {
	module  module.Module
	width   int
	height  int
	row     int
	pattern int
}

func newTrackerModel(m module.Module) trackerModel {
	return trackerModel{
		module: m,
	}
}

func (m *trackerModel) update(state player.PlayerStateUpdate) {
	m.row = state.Row
	m.pattern = state.Pattern
}

func (m trackerModel) View() string {
	if m.module == nil || m.pattern >= m.module.NumPatterns() {
		return ""
	}

	var b strings.Builder

	// Calculate how many channels can be displayed
	availableWidth := m.width - 2 - 2 // subtract border and padding
	maxVisibleChannels := max((availableWidth-rowNumWidth)/channelWidth, 0)
	numChannelsToDisplay := min(m.module.NumChannels(), maxVisibleChannels)

	title := titleStyle.Render(m.module.Name())
	b.WriteString(title + "\n")

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	header := " "
	for ch := range numChannelsToDisplay {
		header += fmt.Sprintf("    Chan %-4d", ch+1)
	}
	if m.module.NumChannels() > numChannelsToDisplay {
		header += "..."
	}
	b.WriteString(headerStyle.Render(header) + "\n")

	// availableHeight is the total height of the component, minus border, padding and header row
	availableHeight := m.height - 4 - 1
	for displayRow := 1; displayRow <= availableHeight; displayRow++ {
		patternRow := m.row + (displayRow - playheadDisplayRow)

		rowStyle := lipgloss.NewStyle() // .Background(lipgloss.Color("6"))
		if displayRow == playheadDisplayRow {
			rowStyle = rowStyle.Background(lipgloss.Color("254"))
		}

		if patternRow >= 0 && patternRow < m.module.NumRows(patternRow) {
			rowNumStr := fmt.Sprintf("%02d", patternRow+1)
			b.WriteString(rowStyle.Foreground(lipgloss.Color("15")).Render(rowNumStr))

			for ch := range numChannelsToDisplay {
				cellData := m.module.PatternCell(m.pattern, patternRow, ch)
				noteStr := noteStyle.Copy().Inherit(rowStyle).Render(cellData.HumanNote)
				instrumentStr := instrumentStyle.Copy().Inherit(rowStyle).Render(fmt.Sprintf("%02X", cellData.Instrument))
				effectStr := effectStyle.Copy().Inherit(rowStyle).Render(fmt.Sprintf("%X%02X", cellData.Effect, cellData.EffectParam))
				cellStr := fmt.Sprintf(" %s %s %s |", noteStr, instrumentStr, effectStr)
				b.WriteString(rowStyle.Render(cellStr))
			}
			b.WriteString("\n")
		} else {
			b.WriteString("\n")
		}
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		Inherit(borderColorStyle).
		Width(m.width - 2).
		Height(m.height - 2)
	return style.Render(b.String())
}
