package main

import (
	"github.com/tandemdude/proofman/commands"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "proofman",
		Usage: "Dependency manager and utility tool for Isabelle",
		Commands: []*cli.Command{
			commands.InitCommand,
			commands.VersionCommand,
			// install <package> <version> [use lockfile?]
			// uninstall <package>
			// lock?
			// package
			// tree
			// go mod tidy equivalent?
			// upload
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
