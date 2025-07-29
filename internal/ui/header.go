package ui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/jesseward/impulse/internal/player"
	"github.com/jesseward/impulse/pkg/module"
)

const (
	keyWidth = 12
)

var (
	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true).
			Inherit(borderColorStyle).
			Padding(0, 1)
	labelStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("27"))
	valueStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

type headerModel struct {
	module   module.Module
	width    int
	state    player.PlayerStateUpdate
	duration int
	speed    int
	bpm      int
}

type row = struct {
	left, lvalue   string
	center, cvalue string
	right, rvalue  string
}

func newHeaderModel(m module.Module) headerModel {
	return headerModel{
		module: m,
		state: player.PlayerStateUpdate{
			Order:   0,
			Pattern: int(m.PatternOrder()[0]),
			Row:     0,
		},
		speed: m.DefaultSpeed(),
		bpm:   m.DefaultBPM(),
	}
}

func (m *headerModel) update(state player.PlayerStateUpdate, duration int) {
	m.state = state
	m.duration = duration
	m.speed = state.Speed
	m.bpm = state.BPM
}

func (m headerModel) height() int {
	return 8
}

func (m headerModel) View() string {

	data := []row{
		{"Filetype", m.module.Type(), "Pattern", strconv.Itoa(m.state.Pattern), "Tempo", strconv.Itoa(m.bpm)},
		{"Position", strconv.Itoa(m.state.Order), "Time", fmt.Sprintf("%02d:%02d", m.duration/60, m.duration%60), "Speed", strconv.Itoa(m.speed)},
	}

	contentWidth := m.width - headerStyle.GetHorizontalFrameSize() - 3

	columnWidth := contentWidth / 3
	lastColumnWidth := contentWidth - (columnWidth * 2)

	var rows []string
	for _, r := range data {
		// --- ✨ FIX: Build each part of the column separately ---
		key1 := labelStyle.Copy().Width(keyWidth).Render(r.left)
		sep := separatorStyle.Render(" | ")
		val1 := valueStyle.Render(r.lvalue)
		// Join the parts to create the aligned content for the column.
		col1Content := lipgloss.JoinHorizontal(lipgloss.Left, key1, sep, val1)
		// Render the full column with its calculated width.
		col1 := lipgloss.NewStyle().Width(columnWidth).Render(col1Content)

		key2 := labelStyle.Copy().Width(keyWidth).Render(r.center)
		val2 := valueStyle.Render(r.cvalue)
		col2Content := lipgloss.JoinHorizontal(lipgloss.Left, key2, sep, val2)
		col2 := lipgloss.NewStyle().Width(columnWidth).Render(col2Content)

		key3 := labelStyle.Copy().Width(keyWidth).Render(r.right)
		val3 := valueStyle.Render(r.rvalue)
		col3Content := lipgloss.JoinHorizontal(lipgloss.Left, key3, sep, val3)
		col3 := lipgloss.NewStyle().Width(lastColumnWidth).Render(col3Content)

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, rows...)
	return headerStyle.Width(m.width - 2).Render(content)
}
