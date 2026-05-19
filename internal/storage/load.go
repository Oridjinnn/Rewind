package storage

import (
	"encoding/json"
	"os"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func LoadSession(path string) (types.Session, error) {

	var session types.Session

	file, err := os.Open(path)
	if err != nil {
		return session, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&session)

	return session, err
}
