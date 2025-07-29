package protracker

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Read reads and parses a MOD file from the given reader.
func Read(r io.Reader) (*ModFile, error) {
	m := &ModFile{}

	// Read the module name
	if _, err := io.ReadFull(r, m.songName[:]); err != nil {
		return nil, err
	}

	// Read the 31 samples
	for i := range 31 {
		var sampleBytes [30]byte
		if _, err := io.ReadFull(r, sampleBytes[:]); err != nil {
			return nil, err
		}
		m.samples[i].name = [22]byte(sampleBytes[0:22])
		m.samples[i].length = uint32(binary.BigEndian.Uint16(sampleBytes[22:24])) * 2
		finetune := int8(sampleBytes[24])
		if finetune > 7 {
			finetune -= 16
		}
		m.samples[i].finetune = finetune
		m.samples[i].volume = min(sampleBytes[25], 64)
		m.samples[i].loopStart = uint32(binary.BigEndian.Uint16(sampleBytes[26:28])) * 2
		m.samples[i].loopLength = uint32(binary.BigEndian.Uint16(sampleBytes[28:30])) * 2
	}

	// Read song length and unused byte
	if err := binary.Read(r, binary.BigEndian, &m.songLength); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Unused); err != nil {
		return nil, err
	}

	// Read pattern order table
	if _, err := io.ReadFull(r, m.patternOrder[:]); err != nil {
		return nil, err
	}

	// Read magic ID and determine number of channels
	if _, err := io.ReadFull(r, m.MagicID[:]); err != nil {
		return nil, err
	}
	m.numChannels = 4 // Default
	switch string(m.MagicID[:]) {
	case "6CHN":
		m.numChannels = 6
	case "8CHN", "OCTA":
		m.numChannels = 8
	case "FLT8":
		m.numChannels = 8
	}

	// Determine the number of patterns
	numPatterns := 0
	for _, patternIndex := range m.patternOrder {
		if int(patternIndex) > numPatterns {
			numPatterns = int(patternIndex)
		}
	}
	numPatterns++

	// Read pattern data
	m.Patterns = make([][]ChannelSequence, numPatterns)
	for i := 0; i < numPatterns; i++ {
		m.Patterns[i] = make([]ChannelSequence, 64*m.numChannels)
		for j := 0; j < 64*m.numChannels; j++ {
			var cellBytes [4]byte
			if _, err := io.ReadFull(r, cellBytes[:]); err != nil {
				return nil, err
			}
			m.Patterns[i][j].SampleNumber = (cellBytes[0] & 0xF0) | (cellBytes[2] >> 4)
			m.Patterns[i][j].Period = (uint16(cellBytes[0]&0x0F) << 8) | uint16(cellBytes[1])
			m.Patterns[i][j].Effect.Command = cellBytes[2] & 0x0F
			m.Patterns[i][j].Effect.X = (cellBytes[3] & 0xF0) >> 4
			m.Patterns[i][j].Effect.Y = cellBytes[3] & 0x0F
		}
	}

	// Read sample data
	for i, s := range m.samples {
		if s.length > 0 {
			sampleData := make([]byte, s.length)
			if _, err := io.ReadFull(r, sampleData); err != nil {
				return nil, fmt.Errorf("error reading sample data for sample %d: %w", i+1, err)
			}
			m.samples[i].data = make([]int16, len(sampleData))
			for j, v := range sampleData {
				m.samples[i].data[j] = int16(v) << 8
			}
		}
	}

	return m, nil
}
