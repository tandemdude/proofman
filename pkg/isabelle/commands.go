package isabelle

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// unsupportedVersions is a pattern that matches the Isabelle versions that proofman cannot
// support due to there being a lack of ROOTS file in the repository during that release. Proofman
// theoretically supports Isabelle version 2015 and later
var unsupportedVersions = regexp.MustCompile(`^Isabelle20(0[56789]|1[012])(-\d+)?$`)

func runIsabelleCommand(args ...string) (string, error) {
	cmd := exec.Command("isabelle", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), err
}

func Version() (string, error) {
	version, err := runIsabelleCommand("version")
	if err != nil {
		return "", err
	}

	if unsupportedVersions.MatchString(version) {
		return "", fmt.Errorf("unsupported Isabelle version '%s' - please update your installation", version)
	}

	return version, nil
}

func MkRoot() (string, error) {
	return runIsabelleCommand("mkroot")
}

func Install(dir string) (string, error) {
	return runIsabelleCommand("install", dir)
}
