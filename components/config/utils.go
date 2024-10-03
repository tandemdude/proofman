package config

import (
	"errors"
	"github.com/pelletier/go-toml/v2"
	"github.com/tandemdude/proofman/internal"
	"os"
	"path"
)

func FileExists(directory string) (bool, error) {
	return internal.PathExists(path.Join(directory, internal.ConfigFileName))
}

func FromFile(directory string) (*ProofmanConfig, error) {
	cfgFilePath := path.Join(directory, internal.ConfigFileName)

	if exists, err := FileExists(cfgFilePath); err != nil || !exists {
		return nil, errors.New("config file not found")
	}

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

	return cfg, nil
}
