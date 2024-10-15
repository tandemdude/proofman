package main

import (
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/commands"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "proofman",
		Usage: "Dependency manager and utility tool for Isabelle",
		Commands: []*cli.Command{
			commands.IndexAfpCommand,
			commands.InitCommand,
			commands.VersionCommand,
			// install <package> <version> [use lockfile?]
			// uninstall <package>
			// lock?
			// tree
			// go mod tidy equivalent?
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "quiet",
				Aliases:  []string{"q"},
				Usage:    "Show less logging output",
				Category: "Logging",
			},
			&cli.BoolFlag{
				Name:     "verbose",
				Aliases:  []string{"v"},
				Usage:    "Show more logging output",
				Category: "Logging",
			},
		},
		Before: func(cCtx *cli.Context) error {
			if cCtx.Bool("quiet") {
				internal.LogLevel = internal.LogLvlQuiet
			}
			if cCtx.Bool("verbose") {
				internal.LogLevel = internal.LogLvlVerbose
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
