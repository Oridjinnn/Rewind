package storage

import (
	"os"
)

func ListSessions(path string) ([]os.DirEntry, error) {

	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	return files, nil
}
