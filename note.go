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

/*
Channel Data:the four bytes of channel data in a pattern division)

7654-3210 7654-3210 7654-3210 7654-3210
wwww xxxxxxxxxxxxxx yyyy zzzzzzzzzzzzzz

    wwwwyyyy (8 bits) is the sample for this channel/division
xxxxxxxxxxxx (12 bits) is the sample's period (or effect parameter)
zzzzzzzzzzzz (12 bits) is the effect for this channel/division
*/

func NewNote(buffer []byte) *Note {
	c := new(Note)
	c.sampleNumber = buffer[0]&240 | buffer[2]>>4
	c.period = uint16(buffer[0])&0x0f*uint16(256) + uint16(buffer[1])
	c.effect = buffer[2] & 15
	c.effectData = buffer[3]

	if c.period > 0 {
		noteIdx := sort.SearchInts(FrequencyTable, int(c.period))
		if noteIdx < len(FrequencyTable) {
			baseNote := NoteTable[noteIdx%len(NoteTable)]
			octave := 4 - noteIdx/len(NoteTable)
			c.Note = fmt.Sprintf("%s%d", baseNote, octave)
		}
	}
	return c
}
