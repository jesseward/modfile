package s3m

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jesseward/impulse/pkg/module"
)

func TestParse(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "acid_atmosphere_q-sou.s3m")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	s3m, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if s3m == nil {
		t.Fatal("Parse() returned nil S3M struct")
	}

	// Basic checks
	expectedSongName := "acid atmospheere  (q-sound)"
	// s3m names are null-padded strings.
	actualSongName := strings.TrimRight(string(s3m.Header.SongName[:]), "\x00")
	if actualSongName != expectedSongName {
		t.Errorf("unexpected song name: got %q, want %q", actualSongName, expectedSongName)
	}

	if s3m.Header.InstrumentCount != 16 {
		t.Errorf("unexpected instrument count: got %d, want %d", s3m.Header.InstrumentCount, 16)
	}

	if s3m.ActualPatternCount != 26 {
		t.Errorf("unexpected actual pattern count: got %d, want %d", s3m.ActualPatternCount, 26)
	}

	if s3m.NumChannels() != 16 {
		t.Errorf("unexpected channel count: got %d, want %d", s3m.NumChannels(), 16)
	}
}

func TestS3M_ImplementsModule(t *testing.T) {
	var _ module.Module = (*S3M)(nil)
}

func TestInstrument_ImplementsSample(t *testing.T) {
	var _ module.Sample = (*Instrument)(nil)
}

func TestNoteToString(t *testing.T) {
	tests := []struct {
		name     string
		note     byte
		expected string
	}{
		{
			name:     "C-4",
			note:     0x40, // Octave 4, Note C
			expected: "C-4",
		},
		{
			name:     "C#5",
			note:     0x51, // Octave 5, Note C#
			expected: "C#5",
		},
		{
			name:     "Note Off",
			note:     254,
			expected: "---",
		},
		{
			name:     "Empty Note",
			note:     255,
			expected: "...",
		},
		{
			name:     "B-8",
			note:     0x8B, // Octave 8, Note B
			expected: "B-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := NoteToString(tt.note); got != tt.expected {
				t.Errorf("NoteToString() = %v, want %v", got, tt.expected)
			}
		})
	}
}
