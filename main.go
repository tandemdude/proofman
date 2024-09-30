package main

import (
	"github.com/tandemdude/proofman/commands"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "proofman",
		Usage: "Dependency manager and utility tool for Isabelle",
		Commands: []*cli.Command{
			commands.InitCommand,
			commands.VenvCommand,
			commands.VersionCommand,
			// install <package> <version> [use lockfile?]
			// uninstall <package>
			// lock?
			// package
			// tree
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
