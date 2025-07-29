package main

import (
	"errors"
	"fmt"

	"github.com/jesseward/impulse/pkg/module"
	"github.com/urfave/cli/v2"
)

func infoAction(c *cli.Context) error {
	if c.NArg() == 0 {
		return cli.Exit(errors.New("no file specified"), 1)
	}
	filePath := c.Args().Get(0)
	module, err := loadModule(filePath)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	printModuleInfo(module)
	return nil
}

func printModuleInfo(module module.Module) {
	fmt.Printf("Successfully parsed file: %s\n", module.Name())
	fmt.Printf("Module Type: %s\n", module.Type())
	fmt.Printf("Song Length: %d\n", module.SongLength())
	fmt.Printf("Song BPM: %d\n", module.DefaultBPM())
	fmt.Printf("Song Speed: %d\n", module.DefaultSpeed())

	fmt.Printf("Number of channels: %d\n", module.NumChannels())
	fmt.Printf("Number of patterns: %d\n", module.NumPatterns())
	fmt.Println("Samples:")
	for i, sample := range module.Samples() {
		fmt.Printf("Sample %d: %s\n", i+1, sample.Name())
	}
}
