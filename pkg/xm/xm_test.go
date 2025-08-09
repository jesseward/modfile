package xm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "volume-envelope.xm")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	_, err = Read(f)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
}
