package commands

import (
	"bytes"
	"github.com/tandemdude/proofman/components/config"
	"github.com/tandemdude/proofman/components/parser"
	"github.com/urfave/cli/v2"
	"os"
)

type WithVersion struct {
	Value   string `json:"value"`
	Version string `json:"version"`
}

type PackageMetadata struct {
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Provides []string      `json:"sessions"`
	Requires []WithVersion `json:"requires"`
}

func package_(cCtx *cli.Context) error {
	// look for an existing configuration file
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	existing, err := config.FileExists(pwd)
	if err != nil || !existing {
		// the file doesn't exist - package from scratch, infer dependencies
	}

	// configuration file exists, we can use it to provide the dependencies
	cfg, err := config.FromFile(pwd)
	if err != nil {
		return err
	}

	// parse the ROOT file to figure out what sessions are provided
	rootContent, err := os.ReadFile("ROOT")
	if err != nil {
		return err
	}
	rootStructure, err := parser.ParseRootFile(bytes.NewReader(rootContent))
	if err != nil {
		return err
	}

	meta := PackageMetadata{
		Name:     cfg.Project.Name,
		Version:  cfg.Project.Version,
		Provides: make([]string, 0),
	}

	for _, chapter := range rootStructure.Chapters {
		for _, session := range chapter.Sessions {
			meta.Provides = append(meta.Provides, session.Name)
		}
	}

	return nil
}

var PackageCommand = &cli.Command{
	Name:   "package",
	Usage:  "Packages the current project into a format (.ipkg) suitable for upload to the Proofman Package Index",
	Action: package_,
}
