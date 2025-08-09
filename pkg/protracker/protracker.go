package protracker

import (
	"math"
	"strings"

	"github.com/jesseward/impulse/pkg/module"
)

var PeriodToNote = map[uint16]string{
	856: "C-1", 808: "C#1", 762: "D-1", 720: "D#1", 678: "E-1", 640: "F-1", 604: "F#1", 570: "G-1", 538: "G#1", 508: "A-1", 480: "A#1", 453: "B-1",
	428: "C-2", 404: "C#2", 381: "D-2", 360: "D#2", 339: "E-2", 320: "F-2", 302: "F#2", 285: "G-2", 269: "G#2", 254: "A-2", 240: "A#2", 226: "B-2",
	214: "C-3", 202: "C#3", 190: "D-3", 180: "D#3", 170: "E-3", 160: "F-3", 151: "F#3", 143: "G-3", 135: "G#3", 127: "A-3", 120: "A#3", 113: "B-3",
	0: module.EmptyNote, // 0 indicates the note is not set
}

// Sample represents the metadata for a single sample in the MOD file.
type Sample struct {
	name       [22]byte
	length     uint32
	finetune   int8
	volume     uint8
	loopStart  uint32
	loopLength uint32
	data       []int16
}

// ChannelAudio represents a single cell in a pattern.
type ChannelSequence struct {
	SampleNumber uint8
	Period       uint16
	Effect       module.Effect
}

// Protracker represents a parsed Amiga MOD file.
type ModFile struct {
	songName     [20]byte
	samples      [31]Sample
	songLength   uint8
	Unused       uint8
	patternOrder [128]uint8
	MagicID      [4]byte
	Patterns     [][]ChannelSequence
	numChannels  int
}

func (m *ModFile) PatternCell(pattern, row, channel int) module.Cell {
	if pattern >= len(m.Patterns) || row >= 64 || channel >= m.numChannels {
		return module.Cell{}
	}
	p := m.Patterns[pattern]
	if row*m.numChannels+channel >= len(p) {
		return module.Cell{}
	}
	cell := p[row*m.numChannels+channel]

	return module.Cell{
		HumanNote:    PeriodToNote[cell.Period],
		SampleNumber: cell.SampleNumber,
		Period:       cell.Period,
		Effect:       cell.Effect.Command,
		EffectParam:  cell.Effect.X<<4 | cell.Effect.Y,
	}
}

func (m *ModFile) PatternOrder() []int {
	var patternOrder []int
	for _, patternIndex := range m.patternOrder {
		patternOrder = append(patternOrder, int(patternIndex))
	}
	return patternOrder
}

func (m *ModFile) Name() string {
	return strings.TrimRight(string(m.songName[:]), "\x00")
}

func (m *ModFile) Type() string {
	return "Protracker"
}
func (m *ModFile) SongLength() int {
	return int(m.songLength)
}

func (m *ModFile) NumChannels() int {
	return m.numChannels
}

func (m *ModFile) NumPatterns() int {
	return len(m.Patterns)
}

func (m *ModFile) NumRows(pattern int) int {
	return 64
}

func (m *ModFile) NoteToString(note byte) string {
	if note == 0 {
		return "..."
	}
	return PeriodToNote[uint16(note)]
}

func (m *ModFile) Samples() []module.Sample {
	samples := make([]module.Sample, 0, len(m.samples))
	for i := range m.samples {
		samples = append(samples, &m.samples[i])
	}
	return samples
}

func (m *ModFile) DefaultSpeed() int {
	return 6
}

func (m *ModFile) DefaultBPM() int {
	return 125
}

func (s *Sample) Data() []int16 {
	return s.data
}

// GetName returns the name of the sample.
func (s *Sample) Name() string {
	return strings.TrimRight(string(s.name[:]), "\x00")
}

// GetLength returns the length of the sample.
func (s *Sample) Length() uint32 {
	return uint32(s.length)
}

// GetVolume returns the volume of the sample.
func (s *Sample) Volume() uint8 {
	return s.volume
}

// GetLoopStart returns the loop start of the sample.
func (s *Sample) LoopStart() uint32 {
	return uint32(s.loopStart)
}

// GetLoopLength returns the loop length of the sample.
func (s *Sample) LoopLength() uint32 {
	return uint32(s.loopLength)
}

func (s *Sample) Finetune() uint32 {
	return uint32(s.finetune)
}

func (s *Sample) Flags() byte {
	return 0
}

func (s *Sample) IsPingPong() bool {
	return false
}

func (s *Sample) RelativeNote() int8 {
	return 0
}

func (s *Sample) Panning() byte {
	return 128 // Protracker is mono
}

func (s *Sample) LoopEnd() uint32 {
	return uint32(s.loopStart + s.loopLength)
}

// renderWaveform generates a multi-line ASCII representation of the audio data.
// It works by downsampling the audio data to fit the specified width and height.
//
// The process is as follows:
//  1. The audio data is divided into a number of "buckets" equal to the width of the view.
//  2. For each bucket, the minimum (trough) and maximum (peak) sample values are found.
//  3. These peak and trough values are then scaled to the height of the view.
//  4. A vertical bar is drawn from the trough to the peak for each bucket, creating a solid,
//     filled waveform.
func (s *Sample) AsciiWaveform(width, height int) string {

	if len(s.data) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	// Create a 2D grid for the waveform display
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	bucketSize := float64(len(s.data)) / float64(width)
	halfHeight := float64(height) / 2.0

	for i := 0; i < width; i++ {
		start := int(float64(i) * bucketSize)
		end := int(float64(i+1) * bucketSize)
		if end > len(s.data) {
			end = len(s.data)
		}
		if start >= end {
			continue
		}

		bucket := s.data[start:end]
		var minVal, maxVal int16 = 0, 0
		for _, s := range bucket {
			if s < minVal {
				minVal = s
			}
			if s > maxVal {
				maxVal = s
			}
		}

		// Normalize and scale to the view height
		yMax := int(math.Round(float64(maxVal)/32767.0*halfHeight + halfHeight))
		yMin := int(math.Round(float64(minVal)/32767.0*halfHeight + halfHeight))

		// Clamp values to be within the grid
		if yMax >= height {
			yMax = height - 1
		}
		if yMin < 0 {
			yMin = 0
		}

		// Draw the vertical bar for the current bucket
		for y := yMin; y <= yMax; y++ {
			if y >= 0 && y < height {
				grid[y][i] = '█'
			}
		}
	}

	// Convert the grid to a single string
	var builder strings.Builder
	for y := 0; y < height; y++ {
		builder.WriteString(string(grid[y]))
		builder.WriteRune('\n')
	}
	return builder.String()
}

func (c *ChannelSequence) GetChannel() (int, int, module.Effect) {
	return int(c.SampleNumber), int(c.Period), c.Effect
}
