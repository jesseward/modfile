package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "impulse",
		Usage: "A command-line MOD and S3M player",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "prof",
				Usage: "enable pprof agent on localhost:6060",
			},
		},
		Before: func(c *cli.Context) error {
			if c.Bool("prof") {
				go func() {
					log.Println("Starting pprof agent on localhost:6060")
					if err := http.ListenAndServe("localhost:6060", nil); err != nil {
						log.Printf("pprof agent failed to start: %v", err)
					}
				}()
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:   "play",
				Usage:  "Play a MOD or S3M file",
				Action: playAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to the MOD or S3M file",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "ui",
						Usage: "start the terminal UI",
					},
					&cli.BoolFlag{
						Name:  "v2",
						Usage: "start the new v2 terminal UI",
					},
				},
			},
			{
				Name:   "convert",
				Usage:  "Convert a MOD or S3M file to WAV or RAW format",
				Action: convertAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to the MOD or S3M file",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "wav",
						Usage: "output format (wav or raw)",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "path to the output file",
					},
				},
			},
			{
				Name:   "info",
				Usage:  "Display information about a MOD or S3M file",
				Action: infoAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}