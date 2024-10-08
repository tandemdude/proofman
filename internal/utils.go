package internal

import (
	"errors"
	"os"
)

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func WriteFile(filename string, content string, makeExecutable bool) error {
	// Create or overwrite the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write content to the file
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	if makeExecutable {
		err = os.Chmod(filename, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func NilOr[T any](item *T, default_ T) T {
	if item == nil {
		return default_
	}
	return *item
}
