package main

import (
	"fmt"
	"os"

	"github.com/jesseward/impulse/pkg/loader"
	"github.com/jesseward/impulse/pkg/module"
)

func loadModule(filePath string) (module.Module, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	m, err := loader.Load(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %v", err)
	}
	return m, nil
}
