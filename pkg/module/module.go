package module

import "fmt"

const (
	EmptyNote = "..."
)

// Module is an interface that represents a music module.
type Module interface {
	Name() string
	Type() string
	SongLength() int
	NumChannels() int
	NumPatterns() int
	Samples() []Sample
	PatternOrder() []int
	DefaultSpeed() int
	DefaultBPM() int
	NumRows(pattern int) int
	// GetPatternCell returns a generic representation of a pattern cell.
	PatternCell(pattern, row, channel int) Cell
}

// Cell represents a single entry in a pattern.
type Cell struct {
	HumanNote    string
	Note         byte
	Instrument   byte
	Volume       byte
	Effect       byte
	EffectParam  byte
	SampleNumber uint8
	Period       uint16
}

// Sample is an interface that represents a sample in a music module.
type Sample interface {
	Name() string
	Length() uint32
	LoopStart() uint32
	LoopEnd() uint32
	LoopLength() uint32
	Volume() byte
	Finetune() uint32
	Data() []int16
	Flags() byte
	IsPingPong() bool
	RelativeNote() int8
	Panning() byte
	AsciiWaveform(width, height int) string
}

type Effect struct {
	Command byte
	X, Y    byte
}

// EffectString returns a human-readable representation of an effect.
func (e *Effect) EffectString() string {
	if e.Command == 0 && e.X == 0 && e.Y == 0 {
		return "..."
	}
	return fmt.Sprintf("%X%X%X", e.Command, e.X, e.Y)
}
