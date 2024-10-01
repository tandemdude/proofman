package commands

import (
	"fmt"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/files"
	"github.com/tandemdude/proofman/internal/isabelle"
	"github.com/urfave/cli/v2"
)

func init_(_ *cli.Context) error {
	fmt.Println("running 'isabelle mkroot' to initialise the project")
	_, err := isabelle.MkRoot()
	if err != nil {
		return fmt.Errorf("failed to create ROOT configuration in the current directory - %s", err)
	}

	// Create version control ignore files
	fmt.Println("creating version control ignore files")
	err = internal.WriteFile(".hgignore", files.HgIgnore)
	if err != nil {
		return fmt.Errorf("failed to create a '.hgignore' file in the current directory - %s", err)
	}
	err = internal.WriteFile(".gitignore", files.GitIgnore)
	if err != nil {
		return fmt.Errorf("failed to create a '.gitignore' file in the current directory - %s", err)
	}

	// Create the venv
	fmt.Println("initialising the virtual environment")
	err = createVenv("venv")
	if err != nil {
		return err
	}

	return nil
}

var InitCommand = &cli.Command{
	Name:   "init",
	Usage:  "Initialises a new Isabelle project in the current directory",
	Action: init_,
}
