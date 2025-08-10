package module

import (
	"fmt"
	"math"
	"strings"
)

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

// renderWaveform generates a multi-line ASCII representation of the audio data.
// It works by downsampling the audio data to fit the specified width and height.
//
// The process is as follows:
//  1. The audio data is divided into a number of "buckets" equal to the width of the view.
//  2. For each bucket, the minimum (trough) and maximum (peak) sample values are found.
//  3. These peak and trough values are then scaled to the height of the view.
//  4. A vertical bar is drawn from the trough to the peak for each bucket, creating a solid,
//     filled waveform.
func AsciiWaveform(s Sample, width, height int) string {

	if len(s.Data()) == 0 || width <= 0 || height <= 0 {
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

	bucketSize := float64(len(s.Data())) / float64(width)
	halfHeight := float64(height) / 2.0

	for i := range width {
		start := int(float64(i) * bucketSize)
		end := min(int(float64(i+1)*bucketSize), len(s.Data()))
		if start >= end {
			continue
		}

		bucket := s.Data()[start:end]
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
	for y := range height {
		builder.WriteString(string(grid[y]))
		builder.WriteRune('\n')
	}
	return builder.String()
}
