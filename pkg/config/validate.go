package config

import (
	"fmt"
	"regexp"
)

var (
	NamePattern    = regexp.MustCompile(`^[\w-]+$`)
	VersionPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func Validate(cfg *ProofmanConfig) error {
	// Name cannot have any spaces, must match '[\w-]+'
	if !NamePattern.MatchString(cfg.Project.Name) {
		return fmt.Errorf("project name is invalid - must match %s", NamePattern.String())
	}

	// Description can be from 0-100 chars
	if len(cfg.Project.Description) > 100 {
		return fmt.Errorf("project description is too long - should be 0-100 chars")
	}

	// Version MUST be in format YYYY-MM-DD
	if !VersionPattern.MatchString(cfg.Project.Version) {
		return fmt.Errorf("project version is invalid - must match %s", VersionPattern.String())
	}

	return nil
}
