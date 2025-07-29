package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jesseward/impulse/internal/player"
	"github.com/jesseward/impulse/internal/ui"
	"github.com/jesseward/impulse/pkg/module"
	"github.com/jesseward/impulse/pkg/protracker"
	"github.com/jesseward/impulse/pkg/s3m"
	"github.com/jesseward/impulse/pkg/xm"
	"github.com/urfave/cli/v2"
)

func playAction(c *cli.Context) error {
	filePath := c.String("file")
	startUI := c.Bool("ui")

	logFile, err := os.OpenFile("impulse.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to open log file: %v", err), 1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	var m module.Module
	m, err = loadModule(filePath)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	if startUI {
		ui.New(m)
		return nil
	}

	printModuleInfo(m)

	opts := player.DefaultPlayerOptions()
	audioPlayer, err := player.NewOtoPlayer(opts)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create OtoPlayer: %v", err), 1)
	}
	defer audioPlayer.Close()

	switch mod := m.(type) {
	case *protracker.ModFile, *s3m.S3M, *xm.Module:
		p := player.NewPlayer(mod, log.Printf, nil, opts)
		if err := p.WriteRaw(audioPlayer, nil); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to render audio file: %v", err), 1)
		}
	default:
		return cli.Exit("ERROR: Module file type not supported for playback.", 1)
	}
	return nil
}
