package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

func LoadSession(
	path string,
) (
	types.Session,
	error,
) {

	var session types.Session

	data, err := os.ReadFile(path)

	if err != nil {
		return session, err
	}

	err = json.Unmarshal(
		data,
		&session,
	)

	if err != nil {
		return session, err
	}

	return session, nil
}

func LoadAllSessions() (
	[]types.Session,
	error,
) {

	var sessions []types.Session

	files, err := filepath.Glob(
		"sessions/*.json",
	)

	if err != nil {
		return sessions, err
	}

	for _, file := range files {

		data, err := os.ReadFile(file)

		if err != nil {
			continue
		}

		var session types.Session

		err = json.Unmarshal(
			data,
			&session,
		)

		if err != nil {
			continue
		}

		sessions = append(
			sessions,
			session,
		)
	}

	return sessions, nil
}
