package commands

import (
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/isabelle"
	"github.com/urfave/cli/v2"
)

const releaseVersion = "0.0.1"
const releaseDate = "development"

func version(_ *cli.Context) error {
	if internal.LogLevel == internal.LogLvlQuiet {
		logging.Quiet(releaseVersion)
		return nil
	}

	logging.Unquiet("Proofman version '%s' %s", releaseVersion, releaseDate)

	isabelleVersion, err := isabelle.Version()

	isabelleFound := false
	if err == nil {
		isabelleFound = true
	}

	if isabelleFound {
		logging.Unquiet("Isabelle version '%s' executable found", isabelleVersion)
	} else {
		logging.Unquiet("Isabelle executable not found - some commands may not work correctly!")
	}

	return nil
}

var VersionCommand = &cli.Command{
	Name:   "version",
	Usage:  "Prints version information",
	Action: version,
}
