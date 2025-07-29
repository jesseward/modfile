package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jesseward/impulse/internal/player"
	"github.com/jesseward/impulse/pkg/protracker"
	"github.com/jesseward/impulse/pkg/s3m"
	"github.com/jesseward/impulse/pkg/xm"
	"github.com/urfave/cli/v2"
)

func convertAction(c *cli.Context) error {
	filePath := c.String("file")
	format := c.String("format")
	output := c.String("output")

	logFile, err := os.OpenFile("impulse.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to open log file: %v", err), 1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	module, err := loadModule(filePath)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	var audioPlayer player.AudioPlayer
	var writer io.WriteCloser
	if output != "" {
		var err error
		writer, err = os.Create(output)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to create output file %s : %v", output, err), 1)
		}
	} else {
		writer = os.Stdout
	}
	defer writer.Close()

	opts := player.DefaultPlayerOptions()
	if format == "wav" {
		audioPlayer = player.NewWavPlayer(writer, opts)
	} else {
		audioPlayer = player.NewStreamPlayer(writer, opts)
	}
	defer audioPlayer.Close()

	switch m := module.(type) {
	case *protracker.ModFile, *s3m.S3M, *xm.Module:
		p := player.NewPlayer(m, log.Printf, nil, opts)
		if err := p.WriteRaw(audioPlayer, nil); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to render audio file: %v", err), 1)
		}
	}
	return nil
}
