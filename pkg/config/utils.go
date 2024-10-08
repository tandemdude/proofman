package config

import (
	"github.com/pelletier/go-toml/v2"
	"github.com/tandemdude/proofman/internal"
	"os"
	"path"
	"time"
)

func FileExistsIn(directory string) (bool, error) {
	return internal.PathExists(path.Join(directory, internal.ConfigFileName))
}

func FromFile(directory string) (*ProofmanConfig, error) {
	cfgFilePath := path.Join(directory, internal.ConfigFileName)

	contents, err := os.ReadFile(cfgFilePath)
	if err != nil {
		return nil, err
	}

	cfg := &ProofmanConfig{}
	err = toml.Unmarshal(contents, cfg)
	if err != nil {
		// file exists but is invalid
		return nil, err
	}

	if err = Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Default() *ProofmanConfig {
	return &ProofmanConfig{
		Project: Project{
			Name:        "NewProject",
			Description: "New Isabelle project using Proofman",
			Version:     time.Now().Format(time.DateOnly),
			Requires:    []string{},
		},
		Package: Package{
			Exclude: []string{
				".venv/**", "**/*.ipkg",
			},
		},
	}
}
