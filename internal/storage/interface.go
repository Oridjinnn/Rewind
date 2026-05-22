package storage

import (
	"github.com/Oridjinnn/Rewind/pkg/types"
)

// Storage defines the interface for session storage backends.
type Storage interface {
	// SaveSession persists a session.
	SaveSession(session types.Session) error

	// LoadSession retrieves a session by ID.
	LoadSession(id string) (types.Session, error)

	// LoadAllSessions retrieves all sessions.
	LoadAllSessions() ([]types.Session, error)

	// ListSessions returns a list of session IDs.
	ListSessions() ([]string, error)

	// DeleteSession removes a session by ID.
	DeleteSession(id string) error

	// SearchSessions searches sessions by query (title, command, summary).
	SearchSessions(query string) ([]types.Session, error)

	// Close closes the storage backend.
	Close() error
}

// ShellHistoryStorage defines the interface for shell history storage.
type ShellHistoryStorage interface {
	// SaveEntry saves a shell history entry.
	SaveEntry(cmd string, exitCode int, workDir string, sessionID string) error

	// GetHistory retrieves shell history entries.
	GetHistory(limit int, offset int) ([]ShellHistoryEntry, error)

	// SearchHistory searches shell history by command text.
	SearchHistory(query string, limit int) ([]ShellHistoryEntry, error)

	// GetHistoryStats returns usage statistics.
	GetHistoryStats() (HistoryStats, error)

	// ImportHistory imports commands from a shell history file.
	ImportHistory(commands []string) (int, error)

	// Close closes the shell history storage.
	Close() error
}

// ShellHistoryEntry represents a single shell history record.
type ShellHistoryEntry struct {
	ID         int64
	Command    string
	ExitCode   int
	WorkingDir string
	ExecutedAt string
	SessionID  string
}

// HistoryStats holds aggregate statistics about shell history.
type HistoryStats struct {
	TotalCommands  int
	UniqueCommands int
	TopCommands    []CommandCount
}

// CommandCount pairs a command with its usage count.
type CommandCount struct {
	Command string
	Count   int
}