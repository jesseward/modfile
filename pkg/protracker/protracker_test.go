package protracker

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jesseward/impulse/pkg/module"
)

func TestRead(t *testing.T) {
	var testModData []byte

	// Module Name (20 bytes) - not null-terminated
	testModData = append(testModData, []byte("THIS IS A TEST SONG!")...)

	// Sample 1 (30 bytes)
	testModData = append(testModData, []byte("TEST SAMPLE 1")...)
	testModData = append(testModData, make([]byte, 22-len("TEST SAMPLE 1"))...) // Name
	testModData = append(testModData, []byte{0x00, 0x02}...)                    // Length = 2 words = 4 bytes
	testModData = append(testModData, 0x00)                                     // Finetune = 0
	testModData = append(testModData, 64)                                       // Volume
	testModData = append(testModData, []byte{0x00, 0x00}...)                    // LoopStart = 0
	testModData = append(testModData, []byte{0x00, 0x01}...)                    // LoopLength = 1 word = 2 bytes

	// Sample 2 (30 bytes)
	testModData = append(testModData, []byte("TEST SAMPLE 2")...)
	testModData = append(testModData, make([]byte, 22-len("TEST SAMPLE 2"))...) // Name
	testModData = append(testModData, []byte{0x00, 0x04}...)                    // Length = 4 words = 8 bytes
	testModData = append(testModData, 0x0F)                                     // Finetune = 15 -> -1
	testModData = append(testModData, 32)                                       // Volume
	testModData = append(testModData, []byte{0x00, 0x01}...)                    // LoopStart = 1 word = 2 bytes
	testModData = append(testModData, []byte{0x00, 0x02}...)                    // LoopLength = 2 words = 4 bytes

	// Samples 3-31 (29 * 30 bytes) - all zero
	testModData = append(testModData, make([]byte, 29*30)...)

	// Song Length (1 byte)
	testModData = append(testModData, 1)

	// Unused (1 byte)
	testModData = append(testModData, 127)

	// Pattern Order Table (128 bytes)
	patternOrder := make([]byte, 128)
	patternOrder[0] = 0 // Play pattern 0
	testModData = append(testModData, patternOrder...)

	// Magic ID (4 bytes)
	testModData = append(testModData, []byte("M.K.")...)

	// Pattern 0 (64 rows * 4 channels * 4 bytes/cell = 1024 bytes)
	patternData := make([]byte, 1024)
	// Row 0, Channel 0: Sample 1, Note C-2 (Period 428), Effect 0xA, params 0x12
	patternData[0] = 0x11 // Sample 1, upper nibble
	patternData[1] = 0xAC // Period low
	patternData[2] = 0x1A // Sample 1, lower nibble, Effect A
	patternData[3] = 0x12 // Effect params
	testModData = append(testModData, patternData...)

	// Sample 1 Data (4 bytes)
	testModData = append(testModData, []byte{10, 20, 30, 40}...)
	// Sample 2 Data (8 bytes)
	testModData = append(testModData, []byte{50, 60, 70, 80, 90, 100, 110, 120}...)

	reader := bytes.NewReader(testModData)
	mod, err := Read(reader)

	if err != nil {
		t.Fatalf("Read() returned an unexpected error: %v", err)
	}

	if name := mod.Name(); name != "THIS IS A TEST SONG!" {
		t.Errorf("Expected module name 'THIS IS A TEST SONG!', got '%s'", name)
	}

	// Check Sample 1
	s1 := mod.Samples()[0].(*Sample)
	if name := s1.Name(); name != "TEST SAMPLE 1" {
		t.Errorf("Expected sample 1 name 'TEST SAMPLE 1', got '%s'", name)
	}
	if s1.Length() != 4 {
		t.Errorf("Expected sample 1 length 4, got %d", s1.Length())
	}
	if s1.Finetune() != 0 {
		t.Errorf("Expected sample 1 finetune 0, got %d", s1.Finetune())
	}
	if s1.Volume() != 64 {
		t.Errorf("Expected sample 1 volume 64, got %d", s1.Volume())
	}
	if s1.LoopStart() != 0 {
		t.Errorf("Expected sample 1 loop start 0, got %d", s1.LoopStart())
	}
	if s1.LoopLength() != 2 {
		t.Errorf("Expected sample 1 loop length 2, got %d", s1.LoopLength())
	}
	if !reflect.DeepEqual(s1.Data(), []int16{10 << 8, 20 << 8, 30 << 8, 40 << 8}) {
		t.Errorf("Expected sample 1 data [2560, 5120, 7680, 10240], got %v", s1.Data())
	}

	// Check Sample 2
	s2 := mod.Samples()[1].(*Sample)
	if name := s2.Name(); name != "TEST SAMPLE 2" {
		t.Errorf("Expected sample 2 name 'TEST SAMPLE 2', got '%s'", name)
	}
	if s2.Length() != 8 {
		t.Errorf("Expected sample 2 length 8, got %d", s2.Length())
	}
	if int8(s2.Finetune()) != -1 {
		t.Errorf("Expected sample 2 finetune -1, got %d", s2.Finetune())
	}
	if s2.Volume() != 32 {
		t.Errorf("Expected sample 2 volume 32, got %d", s2.Volume())
	}
	if s2.LoopStart() != 2 {
		t.Errorf("Expected sample 2 loop start 2, got %d", s2.LoopStart())
	}
	if s2.LoopLength() != 4 {
		t.Errorf("Expected sample 2 loop length 4, got %d", s2.LoopLength())
	}
	if !reflect.DeepEqual(s2.Data(), []int16{50 << 8, 60 << 8, 70 << 8, 80 << 8, 90 << 8, 100 << 8, 110 << 8, 120 << 8}) {
		t.Errorf("Expected sample 2 data, got %v", s2.Data())
	}

	if mod.SongLength() != 1 {
		t.Errorf("Expected song length 1, got %d", mod.SongLength())
	}

	if string(mod.MagicID[:]) != "M.K." {
		t.Errorf("Expected magic ID 'M.K.', got '%s'", string(mod.MagicID[:]))
	}

	if mod.NumChannels() != 4 {
		t.Errorf("Expected 4 channels, got %d", mod.NumChannels())
	}

	if len(mod.Patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(mod.Patterns))
	}

	cell := mod.Patterns[0][0]
	if cell.SampleNumber != 17 {
		t.Errorf("Expected sample number 17, got %d", cell.SampleNumber)
	}

	if cell.Period != 428 {
		t.Errorf("Expected period 428, got %d", cell.Period)
	}

	if cell.Effect.Command != 0xA {
		t.Errorf("Expected effect command 0xA, got %X", cell.Effect.Command)
	}

	if cell.Effect.X != 1 {
		t.Errorf("Expected effect X value 1, got %d", cell.Effect.X)
	}

	if cell.Effect.Y != 2 {
		t.Errorf("Expected effect Y value 2, got %d", cell.Effect.Y)
	}
}

func TestEffectString(t *testing.T) {
	tests := []struct {
		name     string
		effect   module.Effect
		expected string
	}{
		{
			name:     "No effect",
			effect:   module.Effect{Command: 0, X: 0, Y: 0},
			expected: "...",
		},
		{
			name:     "Effect A12",
			effect:   module.Effect{Command: 0xA, X: 1, Y: 2},
			expected: "A12",
		},
		{
			name:     "Effect 00F",
			effect:   module.Effect{Command: 0, X: 0, Y: 0xF},
			expected: "00F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.effect.EffectString(); result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
