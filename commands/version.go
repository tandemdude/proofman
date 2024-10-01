package commands

import (
	"fmt"
	"github.com/tandemdude/proofman/internal/isabelle"
	"github.com/urfave/cli/v2"
)

const releaseVersion = "0.0.1"
const releaseDate = "development"

func version(_ *cli.Context) error {
	fmt.Printf("Proofman version '%s' %s\n", releaseVersion, releaseDate)

	isabelleVersion, err := isabelle.Version()

	isabelleFound := false
	if err == nil {
		isabelleFound = true
	}

	if isabelleFound {
		fmt.Printf("Isabelle '%s' executable found\n", isabelleVersion)
	} else {
		fmt.Println("Isabelle executable not found - some commands may not work correctly!")
	}

	return nil
}

var VersionCommand = &cli.Command{
	Name:   "version",
	Usage:  "Prints version information",
	Action: version,
}
