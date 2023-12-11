package modfile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func loadBuffer(filename string) []byte {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	buf, err := os.ReadFile(filepath.Join(basepath, "testdata", filename))
	if err != nil {
		panic(err)
	}
	return buf
}
func TestRead(t *testing.T) {
	tests := []struct {
		buffer    []byte
		title     string
		modFormat *ModuleFormat
	}{
		{buffer: loadBuffer("taketrip.mod"), title: "take a trip from me", modFormat: &ModFormatMK},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			module, err := Read(tt.buffer)
			if err != nil {
				t.Fatalf("unexpected error when reading modfile: %v", err)
			}

			if module.Name != tt.title {
				t.Errorf("wanted title '%s, got '%s'", tt.title, module.Name)
			}

			if *module.Format != *tt.modFormat {
				t.Errorf("wanted format '%s', got '%s'", tt.modFormat.Name, module.Format.Name)
			}

			if len(module.Patterns) != 38 {
				t.Errorf("wanted patterns %d, got %d", 38, len(module.Patterns))
			}

			if module.Songlength != 55 {
				t.Errorf("wanted songlength %d, got %d", 55, module.Songlength)
			}
		})
	}
}
