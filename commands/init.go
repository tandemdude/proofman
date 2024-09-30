package commands

import (
	"fmt"
	"github.com/tandemdude/proofman/internal"
	"github.com/urfave/cli/v2"
	"os/exec"
)

const generatedHgIgnore = `syntax: glob

venv/**
`
const generatedGitIgnore = `venv/
`

func init_(cCtx *cli.Context) error {
	fmt.Println("running 'isabelle mkroot' to initialise the project")
	cmd := exec.Command("isabelle", "mkroot")
	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("failed to create a ROOT file in the current directory - %s", err)
	}

	// Create version control ignore files
	fmt.Println("creating version control ignore files")
	err = internal.WriteFile(".hgignore", generatedHgIgnore)
	if err != nil {
		return fmt.Errorf("failed to create a '.hgignore' file in the current directory - %s", err)
	}
	err = internal.WriteFile(".gitignore", generatedGitIgnore)
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
