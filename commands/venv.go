package commands

import (
	"errors"
	"fmt"
	"github.com/tandemdude/proofman/internal"
	"github.com/urfave/cli/v2"
	"os"
)

func createVenv(path string) error {
	// Check if a virtual environment is already active
	if os.Getenv("PROOFMAN_VENV_ACTIVE") == "1" {
		return errors.New("a virtual environment is already active - run 'deactivate' to exit it")
	}

	// Retrieve the directory path from the command arguments
	if len(path) == 0 {
		return errors.New("a directory name/path is required")
	}

	exists, err := internal.PathExists(path)
	if err != nil {
		return fmt.Errorf("error checking if %s exists - %s", path, err)
	}

	if exists {
		info, _ := os.Stat(path)
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", path)
		}

		dir, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening %s - %s", path, err)
		}
		defer dir.Close()

		files, err := dir.Readdirnames(1)
		if err != nil || len(files) > 0 {
			return fmt.Errorf("error listing %s contents, or the directory is non-empty", path)
		}
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Add required files to the virtual environment
	// - activate
	// - deactivate
	// - dependencies directory
	fmt.Printf("created a new virtual environment at %s\n", path)

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
