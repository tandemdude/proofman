package isabelle

import "os/exec"

func runIsabelleCommand(args ...string) (string, error) {
	cmd := exec.Command("isabelle", args...)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func Version() (string, error) {
	return runIsabelleCommand("version")
}

func MkRoot() (string, error) {
	return runIsabelleCommand("mkroot")
}

func Install(dir string) (string, error) {
	return runIsabelleCommand("install", dir)
}
