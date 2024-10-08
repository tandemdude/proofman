package commands

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pelletier/go-toml/v2"
	"github.com/tandemdude/proofman/internal"
	"github.com/tandemdude/proofman/internal/files"
	"github.com/tandemdude/proofman/internal/logging"
	"github.com/tandemdude/proofman/pkg/config"
	"github.com/tandemdude/proofman/pkg/isabelle"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func createVenv(pwd, path string) error {
	// Check if a virtual environment is already active
	if os.Getenv("PROOFMAN_VENV_ACTIVE") == "1" {
		return errors.New("a virtual environment is already active - run 'deactivate' to exit it")
	}

	// Retrieve the directory path from the command arguments
	if len(path) == 0 {
		return errors.New("a path is required")
	}

	path, err := filepath.Abs(filepath.Join(pwd, path))
	if err != nil {
		return err
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

	proxiedBinDir := filepath.Join(binDir, "_proxied")
	_, err = isabelle.Install(proxiedBinDir)
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

	// Bootstrap with the required 'activate' and proxied isabelle scripts
	err = internal.WriteFile(filepath.Join(binDir, "activate"), files.NewActivateSh(path), false)
	if err != nil {
		return fmt.Errorf("failed to create activate script - %s", err)
	}
	err = internal.WriteFile(filepath.Join(binDir, "isabelle"), files.IsabelleProxyScript, true)
	if err != nil {
		return fmt.Errorf("failed to create isabelle script - %s", err)
	}
	err = internal.WriteFile(filepath.Join(binDir, "isabelle_java"), files.IsabelleJavaProxyScript, true)
	if err != nil {
		return fmt.Errorf("failed to create isabelle_java script - %s", err)
	}

	// Bootstrap the local isabelle settings file
	err = internal.WriteFile(filepath.Join(localHomeUserDir, "settings"), files.IsabelleHomeUserSettings, true)
	if err != nil {
		return fmt.Errorf("failed to create Isabelle user settings script - %s", err)
	}

	logging.Unquiet("created a new virtual environment at '%s'", path)

	return nil
}

func init_(_ *cli.Context) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// TODO - look into cleaning the directory string (in case of spaces etc)
	dirName := filepath.Base(pwd)

	// Take input for config values
	namePrompt := promptui.Prompt{
		Label:   "Project Name [a-zA-Z0-9_-]",
		Default: dirName,
		Validate: func(s string) error {
			matches := config.NamePattern.MatchString(s)
			if !matches {
				return errors.New("invalid project name")
			}
			return nil
		},
	}
	projectName, err := namePrompt.Run()
	if err != nil {
		return err
	}

	gitOrMercurialPrompt := promptui.Select{
		Label: "Version Control System",
		Items: []string{"Git", "Mercurial", "N/A"},
	}
	_, gitOrMercurial, err := gitOrMercurialPrompt.Run()
	if err != nil {
		return err
	}

	logging.Verbose("initialising the Isabelle project")
	_, err = isabelle.MkRoot()
	if err != nil {
		return fmt.Errorf("failed to create ROOT configuration in the current directory - %s", err)
	}

	// Create version control ignore files
	if gitOrMercurial != "N/A" {
		if gitOrMercurial == "Mercurial" {
			err = internal.WriteFile(".hgignore", files.HgIgnore, false)
			if err != nil {
				return fmt.Errorf("failed to create a '.hgignore' file in the current directory - %s", err)
			}
		} else if gitOrMercurial == "Git" {
			err = internal.WriteFile(".gitignore", files.GitIgnore, false)
			if err != nil {
				return fmt.Errorf("failed to create a '.gitignore' file in the current directory - %s", err)
			}
		}
	}

	// Create proofman config file
	logging.Verbose("creating Proofman config file")
	defaultConfig := config.Default()
	defaultConfig.Project.Name = projectName

	dumped, err := toml.Marshal(config.Default())
	if err != nil {
		return fmt.Errorf("failed to create default configuration - %s", err)
	}

	err = internal.WriteFile(internal.ConfigFileName, string(dumped), false)
	if err != nil {
		return fmt.Errorf("failed to create a 'proofman.toml' file in the current directory - %s", err)
	}

	// Create the venv
	logging.Verbose("initialising the virtual environment")
	err = createVenv(pwd, ".venv")
	if err != nil {
		return err
	}

	logging.Unquiet("initialisation completed successfully - happy proving!")

	return nil
}

var InitCommand = &cli.Command{
	Name:   "init",
	Usage:  "Initialises a new Isabelle project in the current directory",
	Action: init_,
}
