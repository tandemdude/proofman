package commands

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os/exec"
)

const releaseVersion = "0.0.1"
const releaseDate = "development"

func version(_ *cli.Context) error {
	fmt.Printf("Proofman version \"%s\" %s\n", releaseVersion, releaseDate)

	cmd := exec.Command("isabelle", "version")
	err := cmd.Run()

	isabelleFound := true
	if err != nil {
		isabelleFound = false
	}

	isabelleVersion := ""
	if isabelleFound {
		out, err := cmd.CombinedOutput()
		if err != nil {
			isabelleFound = false
		}
		isabelleVersion = string(out)
	}

	if isabelleFound {
		fmt.Printf("Isabelle \"%s\" executable found\n", isabelleVersion)
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
