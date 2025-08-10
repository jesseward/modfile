package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jesseward/impulse/pkg/module"
)

type waveformModel struct {
	sample module.Sample
	width  int
	height int
}

func newWaveformModel(s module.Sample, w, h int) waveformModel {
	return waveformModel{
		sample: s,
		width:  w,
		height: h,
	}
}

func (m waveformModel) View() string {
	title := titleStyle.Render("Sample '" + m.sample.Name() + "'")
	waveform := noteStyle.Render(module.AsciiWaveform(m.sample, m.width/2, m.height/2))

	// Calculate the size of the dialog box (50% of screen) and add some padding
	dialogWidth := (m.width / 2) + 6
	dialogHeight := (m.height / 2) + 6

	// Create the styled dialog box
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).Inherit(borderColorStyle).
		Padding(1, 1).
		Width(dialogWidth).   // Set the width
		Height(dialogHeight). // Set the height
		Render(title + "\n\n" + waveform)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center, // Horizontal alignment
		lipgloss.Center, // Vertical alignment
		dialogBox,
	)
}
