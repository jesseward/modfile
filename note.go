package modfile

import (
	"fmt"
	"sort"
)

type Note struct {
	Note         string
	sampleNumber uint8
	period       uint16
	effect       uint8
	effectData   uint8
}

func (n Note) String() string {
	return fmt.Sprintf("..%3s..%.2d..%.2d..%.3d..", n.Note, n.sampleNumber, n.effect, n.effectData)
}

// frequencyTable maps  frequencies for notes/octaves
var frequencyTable = []int{
	// B, A#, A, G#, G, F#, F, E, D#, D, C#, C
	57, 60, 64, 67, 71, 76, 80, 85, 90, 95, 101, 107,
	113, 120, 127, 135, 143, 151, 160, 170, 180, 190, 202, 214,
	226, 240, 254, 269, 285, 302, 320, 339, 360, 381, 404, 428,
	453, 480, 508, 538, 570, 604, 640, 678, 720, 762, 808, 856,
	907, 961, 1017, 1077, 1141, 1209, 1281, 1357, 1440, 1525, 1616, 1712,
}

// noteTable is the index of notes per frequency table below
var noteTable = [12]string{"B-", "A#", "A-", "G#", "G-", "F#", "F-", "E-", "D#", "D-", "C#", "C-"}

func NewNote(buffer []byte) *Note {

	/*
		Channel Data:the four bytes of channel data in a pattern division)

		7654-3210 7654-3210 7654-3210 7654-3210
		wwww xxxxxxxxxxxxxx yyyy zzzzzzzzzzzzzz

		    wwwwyyyy (8 bits) is the sample for this channel/division
		xxxxxxxxxxxx (12 bits) is the sample's period (or effect parameter)
		zzzzzzzzzzzz (12 bits) is the effect for this channel/division
	*/

	c := new(Note)
	c.sampleNumber = buffer[0]&240 | buffer[2]>>4
	c.period = uint16(buffer[0])&0x0f*uint16(256) + uint16(buffer[1])
	c.effect = buffer[2] & 15
	c.effectData = buffer[3]

	if c.period > 0 {
		noteIdx := sort.SearchInts(frequencyTable, int(c.period))
		if noteIdx < len(frequencyTable) {
			baseNote := noteTable[noteIdx%len(noteTable)]
			octave := 4 - noteIdx/len(noteTable)
			c.Note = fmt.Sprintf("%s%d", baseNote, octave)
		}
	}
	return c
}
