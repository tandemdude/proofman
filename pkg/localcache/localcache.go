package localcache

import (
	"github.com/tandemdude/proofman/internal"
	"os"
	"path/filepath"
)

func basePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".proofman"), nil
}

func WriteFile(name string, data string) error {
	base, err := basePath()
	if err != nil {
		return err
	}

	filePath := filepath.Join(base, name)
	if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	err = internal.WriteFile(filePath, data, false)
	if err != nil {
		return err
	}

	return nil
}

func ReadFile(name string) string {
	base, err := basePath()
	if err != nil {
		return ""
	}

	filePath := filepath.Join(base, name)
	if exists, err := internal.PathExists(filePath); err != nil || !exists {
		return ""
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	return string(content)
}
