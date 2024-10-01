package commands

import (
	"errors"
	"fmt"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/files"
	"github.com/tandemdude/proofman/internal/isabelle"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func createVenv(path string) error {
	// Check if a virtual environment is already active
	if os.Getenv("PROOFMAN_VENV_ACTIVE") == "1" {
		return errors.New("a virtual environment is already active - run 'deactivate' to exit it")
	}

	// Retrieve the directory path from the command arguments
	if len(path) == 0 {
		path = "venv"
	}

	exists, err := internal.PathExists(path)
	if err != nil {
		return fmt.Errorf("error checking if '%s' exists - %s", path, err)
	}

	if exists {
		info, _ := os.Stat(path)
		if !info.IsDir() {
			return fmt.Errorf("'%s' is not a directory", path)
		}

		dir, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening '%s' - %s", path, err)
		}
		defer dir.Close()

		names, err := dir.Readdirnames(1)
		if err != nil || len(names) > 0 {
			return fmt.Errorf("error listing '%s' contents, or the directory is non-empty", path)
		}
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory - %s", err)
	}

	isabelleVersion, err := isabelle.Version()
	if err != nil {
		return fmt.Errorf("failed to call isabelle executable - %s", err)
	}

	binDir := filepath.Join(path, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory - %s", err)
	}
	_, err = isabelle.Install(binDir)
	if err != nil {
		return fmt.Errorf("failed to install isabelle executables into virtual environment - %s", err)
	}

	dependenciesDir := filepath.Join(path, "deps")
	if err := os.MkdirAll(dependenciesDir, 0755); err != nil {
		return fmt.Errorf("failed to create dependencies directory - %s", err)
	}

	localHomeUserDir := filepath.Join(path, ".isabelle", isabelleVersion, "etc")
	if err := os.MkdirAll(localHomeUserDir, 0755); err != nil {
		return fmt.Errorf("failed to create local home user directory - %s", err)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory - %s", err)
	}
	absoluteVenvPath, err := filepath.Abs(filepath.Join(workDir, path))
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path - %s", err)
	}

	// Bootstrap with the required 'activate' and 'deactivate' scripts
	err = internal.WriteFile(filepath.Join(binDir, "activate"), files.NewActivateSh(absoluteVenvPath))
	if err != nil {
		return fmt.Errorf("failed to create activate script - %s", err)
	}
	err = internal.WriteFile(filepath.Join(binDir, "deactivate"), files.DeactivateSh)
	if err != nil {
		return fmt.Errorf("failed to create deactivate script - %s", err)
	}

	// Bootstrap the local isabelle settings file
	err = internal.WriteFile(filepath.Join(localHomeUserDir, "settings"), files.IsabelleHomeUserSettings)
	if err != nil {
		return fmt.Errorf("failed to create Isabelle user settings script - %s", err)
	}

	fmt.Printf("created a new virtual environment at '%s'\n", path)

	return nil
}

func venv(cCtx *cli.Context) error {
	return createVenv(cCtx.Args().Get(0))
}

var VenvCommand = &cli.Command{
	Name:   "venv",
	Usage:  "Creates a new virtual environment at the given path",
	Action: venv,
}
