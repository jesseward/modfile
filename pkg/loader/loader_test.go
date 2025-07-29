package loader

import (
	"os"
	"testing"
)

func TestLoad_S3M(t *testing.T) {
	file, err := os.Open("../../examples/acid_atmosphere_q-sou.s3m")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	module, err := Load(file)
	if err != nil {
		t.Fatalf("Load() error = %v, wantErr nil", err)
	}
	if module == nil {
		t.Fatal("Load() module = nil, want not nil")
	}

	if module.Type() != "S3M" {
		t.Errorf("Load() module.Type() = %v, want S3M", module.Type())
	}
}
