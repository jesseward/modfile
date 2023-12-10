package modfile

import (
	"fmt"
	"strings"
)

func (m *Mod) PrintPatternData() string {

	var pattern strings.Builder

	for c := uint8(0); c < m.numberOfPatterns; c++ {

		for row, channel := range m.Patterns[m.SequenceTable[c]].rowChannel {
			channelNotes := ""
			for _, note := range channel {
				channelNotes = fmt.Sprintf("%s|| %s ||", channelNotes, note.String())

			}
			pattern.WriteString(fmt.Sprintf("..%.2d..%s\n", row, channelNotes))
		}
	}
	return pattern.String()
}

func (m *Mod) PrintModInfo() string {
	var mod strings.Builder

	mod.WriteString(fmt.Sprintf("%10s : %s\n", "Title", m.Name))
	mod.WriteString(fmt.Sprintf("%10s : %s\n", "Format", m.Format.Name))
	mod.WriteString(fmt.Sprintf("%10s : %d\n", "Channels", m.Format.Channels))
	mod.WriteString(fmt.Sprintf("%10s : %d\n", "Patterns", m.numberOfPatterns))
	mod.WriteString(fmt.Sprintf("%10s : %d\n", "Length", m.Songlength))
	mod.WriteString(fmt.Sprintf("%-2s|%-25s|%8s|%5s\n", "#", "Sample Name", "Length", "Loops"))

	for i := uint8(0); i < m.Format.Samples; i++ {
		mod.WriteString(fmt.Sprintf("%.2d %s", i, m.Samples[i].String()))
	}
	mod.WriteString("Pattern Data\n")

	return mod.String()
}
