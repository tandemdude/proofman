package git

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), err
}

func runGitCommandInDirectory(inDirectory string, args ...string) (string, error) {
	currdir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	err = os.Chdir(inDirectory)
	if err != nil {
		return "", err
	}

	out, err := runGitCommand(args...)
	return out, errors.Join(err, os.Chdir(currdir))
}

func SetupRemote(remoteUrl, directory string) error {
	_, err := runGitCommandInDirectory(directory, "init")
	if err != nil {
		return err
	}

	_, err = runGitCommandInDirectory(directory, "remote", "add", "origin", remoteUrl)
	if err != nil {
		return err
	}

	_, err = runGitCommandInDirectory(directory, "fetch", "origin", "main")
	if err != nil {
		return err
	}

	_, err = runGitCommandInDirectory(directory, "reset", "--hard", "origin/main")
	if err != nil {
		return err
	}

	return nil
}

func MakeBranch(inDirectory string, name string) error {
	_, err := runGitCommandInDirectory(inDirectory, "checkout", "-b", name)
	return err
}

func AddAll(inDirectory string) error {
	_, err := runGitCommandInDirectory(inDirectory, "add", ".")
	return err
}

func Commit(inDirectory string, message string) error {
	_, err := runGitCommandInDirectory(inDirectory, "commit", "-m", message)
	return err
}

func Push(inDirectory string) error {
	_, err := runGitCommandInDirectory(inDirectory, "push", "-u", "origin", "HEAD")
	return err
}
