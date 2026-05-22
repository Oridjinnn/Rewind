package storage

import (
	"encoding/json"
	"os"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

func SaveSession(session types.Session, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(session)
}
