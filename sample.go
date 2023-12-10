package modfile

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Sample struct {
	Name         string
	SampleLength uint16
	Finetune     uint8
	Volume       uint8
	// Once this sample has been played completely from beginning to end, if the RepeatLength
	// (next field) is greater than two bytes it will loop back to this position in the sample
	// and continue playing.
	// Once it has played for the repeat length,it continues to loop back to the repeat start
	// offset.  This means the sample continue playing until it is told to stop.
	RepeatPoint  uint16
	RepeatLength uint16
	AudioData    []byte
}

// IsLooped returns true if the sample is to be looped. A sample is only looped if the
// RepeatLength value is greater than 2 bytes
func (s *Sample) IsLooped() bool {
	return s.RepeatLength > 2
}

func (s Sample) String() string {
	return fmt.Sprintf("%-24s | %5d  | %5v\n", s.Name, s.SampleLength, s.IsLooped())
}

func NewSample(buffer []byte) *Sample {
	sample := new(Sample)
	var offset uint16

	sample.Name = strings.TrimRight(string(buffer[offset:offset+lengthSampleName]), "\x00")
	offset += lengthSampleName

	sample.SampleLength = binary.BigEndian.Uint16(buffer[offset:offset+2]) * 2
	offset += 2

	sample.Finetune = uint8(buffer[offset])
	offset++

	sample.Volume = uint8(buffer[offset])
	offset++

	sample.RepeatPoint = binary.BigEndian.Uint16(buffer[offset:offset+2]) * 2
	offset += 2

	sample.RepeatLength = binary.BigEndian.Uint16(buffer[offset:offset+2]) * 2
	return sample
}
