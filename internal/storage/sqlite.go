package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements both Storage and ShellHistoryStorage interfaces using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed storage.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		command TEXT NOT NULL,
		model TEXT DEFAULT '',
		title TEXT DEFAULT '',
		summary TEXT DEFAULT '',
		tags TEXT DEFAULT '[]',
		mood TEXT DEFAULT '',
		started_at DATETIME NOT NULL,
		ended_at DATETIME NOT NULL,
		exit_code INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		type TEXT NOT NULL,
		content TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS shell_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		exit_code INTEGER DEFAULT 0,
		working_dir TEXT DEFAULT '',
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		session_id TEXT DEFAULT '',
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET DEFAULT
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON sessions(started_at);
	CREATE INDEX IF NOT EXISTS idx_sessions_command ON sessions(command);
	CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);
	CREATE INDEX IF NOT EXISTS idx_shell_history_executed_at ON shell_history(executed_at);
	CREATE INDEX IF NOT EXISTS idx_shell_history_command ON shell_history(command);
	`
	_, err := s.db.Exec(schema)
	return err
}

// SaveSession implements Storage.SaveSession.
func (s *SQLiteStore) SaveSession(session types.Session) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	tagsJSON, err := json.Marshal(session.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	_, err = tx.Exec(`
		INSERT INTO sessions (id, command, model, title, summary, tags, mood, started_at, ended_at, exit_code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			command=excluded.command,
			model=excluded.model,
			title=excluded.title,
			summary=excluded.summary,
			tags=excluded.tags,
			mood=excluded.mood,
			started_at=excluded.started_at,
			ended_at=excluded.ended_at,
			exit_code=excluded.exit_code
	`, session.ID, session.Command, session.Model, session.Title, session.Summary,
		string(tagsJSON), session.Mood, session.StartedAt, session.EndedAt, session.ExitCode)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	_, err = tx.Exec("DELETE FROM events WHERE session_id = ?", session.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	for _, event := range session.Events {
		_, err = tx.Exec(
			"INSERT INTO events (session_id, timestamp, type, content) VALUES (?, ?, ?, ?)",
			session.ID, event.Timestamp, event.Type, event.Content,
		)
		if err != nil {
			return fmt.Errorf("failed to save event: %w", err)
		}
	}

	return tx.Commit()
}

// LoadSession implements Storage.LoadSession.
func (s *SQLiteStore) LoadSession(id string) (types.Session, error) {
	var session types.Session
	var tagsJSON string

	row := s.db.QueryRow(`
		SELECT id, command, model, title, summary, tags, mood, started_at, ended_at, exit_code
		FROM sessions WHERE id = ?
	`, id)

	err := row.Scan(&session.ID, &session.Command, &session.Model, &session.Title,
		&session.Summary, &tagsJSON, &session.Mood, &session.StartedAt,
		&session.EndedAt, &session.ExitCode)
	if err != nil {
		return session, fmt.Errorf("session not found: %w", err)
	}

	json.Unmarshal([]byte(tagsJSON), &session.Tags)

	// Load events
	rows, err := s.db.Query(
		"SELECT timestamp, type, content FROM events WHERE session_id = ? ORDER BY id",
		session.ID,
	)
	if err != nil {
		return session, fmt.Errorf("failed to load events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event types.Event
		if err := rows.Scan(&event.Timestamp, &event.Type, &event.Content); err != nil {
			continue
		}
		session.Events = append(session.Events, event)
	}

	return session, nil
}

// LoadAllSessions implements Storage.LoadAllSessions.
func (s *SQLiteStore) LoadAllSessions() ([]types.Session, error) {
	rows, err := s.db.Query("SELECT id FROM sessions ORDER BY started_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		session, err := s.LoadSession(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// ListSessions implements Storage.ListSessions.
func (s *SQLiteStore) ListSessions() ([]string, error) {
	rows, err := s.db.Query("SELECT id FROM sessions ORDER BY started_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// DeleteSession implements Storage.DeleteSession.
func (s *SQLiteStore) DeleteSession(id string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

// SearchSessions implements Storage.SearchSessions.
func (s *SQLiteStore) SearchSessions(query string) ([]types.Session, error) {
	likePattern := "%" + query + "%"
	rows, err := s.db.Query(`
		SELECT id FROM sessions
		WHERE command LIKE ? OR title LIKE ? OR summary LIKE ?
		ORDER BY started_at DESC
	`, likePattern, likePattern, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search sessions: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		session, err := s.LoadSession(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Close implements Storage.Close.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- ShellHistoryStorage Implementation ---

// SaveEntry implements ShellHistoryStorage.SaveEntry.
func (s *SQLiteStore) SaveEntry(cmd string, exitCode int, workDir string, sessionID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		INSERT INTO shell_history (command, exit_code, working_dir, executed_at, session_id)
		VALUES (?, ?, ?, ?, ?)
	`, cmd, exitCode, workDir, now, sessionID)
	return err
}

// GetHistory implements ShellHistoryStorage.GetHistory.
func (s *SQLiteStore) GetHistory(limit int, offset int) ([]ShellHistoryEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, command, exit_code, working_dir, executed_at, session_id
		FROM shell_history
		ORDER BY executed_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

// SearchHistory implements ShellHistoryStorage.SearchHistory.
func (s *SQLiteStore) SearchHistory(query string, limit int) ([]ShellHistoryEntry, error) {
	likePattern := "%" + query + "%"
	rows, err := s.db.Query(`
		SELECT id, command, exit_code, working_dir, executed_at, session_id
		FROM shell_history
		WHERE command LIKE ?
		ORDER BY executed_at DESC
		LIMIT ?
	`, likePattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search history: %w", err)
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

// GetHistoryStats implements ShellHistoryStorage.GetHistoryStats.
func (s *SQLiteStore) GetHistoryStats() (HistoryStats, error) {
	stats := HistoryStats{}

	// Total commands
	err := s.db.QueryRow("SELECT COUNT(*) FROM shell_history").Scan(&stats.TotalCommands)
	if err != nil {
		return stats, err
	}

	// Unique commands
	err = s.db.QueryRow("SELECT COUNT(DISTINCT command) FROM shell_history").Scan(&stats.UniqueCommands)
	if err != nil {
		return stats, err
	}

	// Top commands (top 10)
	rows, err := s.db.Query(`
		SELECT command, COUNT(*) as cnt
		FROM shell_history
		GROUP BY command
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var cc CommandCount
		if err := rows.Scan(&cc.Command, &cc.Count); err != nil {
			continue
		}
		stats.TopCommands = append(stats.TopCommands, cc)
	}

	return stats, nil
}

// ImportHistory implements ShellHistoryStorage.ImportHistory.
func (s *SQLiteStore) ImportHistory(commands []string) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO shell_history (command, exit_code, working_dir, executed_at)
		VALUES (?, 0, '', datetime('now'))
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" || strings.HasPrefix(cmd, "#") {
			continue
		}
		if _, err := stmt.Exec(cmd); err != nil {
			continue
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}

func scanHistoryRows(rows *sql.Rows) ([]ShellHistoryEntry, error) {
	var entries []ShellHistoryEntry
	for rows.Next() {
		var e ShellHistoryEntry
		if err := rows.Scan(&e.ID, &e.Command, &e.ExitCode, &e.WorkingDir, &e.ExecutedAt, &e.SessionID); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}