package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

// MigrateJSONToSQLite migrates all JSON sessions from a directory into an SQLiteStore.
func MigrateJSONToSQLite(sessionsDir string, store *SQLiteStore) (int, error) {
	files, err := filepath.Glob(filepath.Join(sessionsDir, "*.json"))
	if err != nil {
		return 0, fmt.Errorf("failed to glob sessions: %w", err)
	}

	migrated := 0
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("warning: skipping %s: %v\n", file, err)
			continue
		}

		var session types.Session
		if err := json.Unmarshal(data, &session); err != nil {
			fmt.Printf("warning: skipping %s: %v\n", file, err)
			continue
		}

		// Use filename as ID if session has no ID
		if session.ID == "" {
			base := filepath.Base(file)
			session.ID = strings.TrimSuffix(base, ".json")
		}

		if err := store.SaveSession(session); err != nil {
			fmt.Printf("warning: failed to migrate %s: %v\n", session.ID, err)
			continue
		}

		migrated++
	}

	return migrated, nil
}