package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Oridjinnn/Rewind/pkg/types"
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

	CREATE TABLE IF NOT EXISTS ide_activity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ide_name TEXT NOT NULL,
		project_name TEXT DEFAULT '',
		project_path TEXT DEFAULT '',
		activity_type TEXT NOT NULL,
		file_path TEXT DEFAULT '',
		language TEXT DEFAULT '',
		metadata TEXT DEFAULT '{}',
		content_snapshot TEXT DEFAULT '',
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		session_id TEXT DEFAULT '',
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS ide_permissions (
		ide_name TEXT NOT NULL,
		project_path TEXT DEFAULT '*',
		recording_enabled INTEGER DEFAULT 0,
		file_recording INTEGER DEFAULT 0,
		terminal_recording INTEGER DEFAULT 0,
		ai_recording INTEGER DEFAULT 0,
		last_toggled DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (ide_name, project_path)
	);

	CREATE INDEX IF NOT EXISTS idx_ide_activity_ide ON ide_activity(ide_name);
	CREATE INDEX IF NOT EXISTS idx_ide_activity_type ON ide_activity(activity_type);
	CREATE INDEX IF NOT EXISTS idx_ide_activity_file ON ide_activity(file_path);
	CREATE INDEX IF NOT EXISTS idx_ide_activity_time ON ide_activity(executed_at);
	CREATE INDEX IF NOT EXISTS idx_ide_activity_project ON ide_activity(project_path);
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
// Uses a single JOIN query to avoid N+1 queries (fixes audit issue #6).
func (s *SQLiteStore) LoadAllSessions() ([]types.Session, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.command, s.model, s.title, s.summary, s.tags, s.mood,
		       s.started_at, s.ended_at, s.exit_code,
		       e.timestamp, e.type, e.content
		FROM sessions s
		LEFT JOIN events e ON e.session_id = s.id
		ORDER BY s.started_at DESC, e.id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	var current *types.Session

	for rows.Next() {
		var (
			id, command, model, title, summary, tagsJSON, mood string
			startedAt, endedAt                                 string
			exitCode                                            int
			eventTimestamp, eventType, eventContent             sql.NullString
		)
		if err := rows.Scan(&id, &command, &model, &title, &summary, &tagsJSON, &mood,
			&startedAt, &endedAt, &exitCode,
			&eventTimestamp, &eventType, &eventContent); err != nil {
			continue
		}

		// Start new session if id changes
		if current == nil || current.ID != id {
			if current != nil {
				sessions = append(sessions, *current)
			}
			s := types.Session{
				ID:        id,
				Command:   command,
				Model:     model,
				Title:     title,
				Summary:   summary,
				Mood:      mood,
				StartedAt: startedAt,
				EndedAt:   endedAt,
				ExitCode:  exitCode,
				Events:    []types.Event{},
			}
			json.Unmarshal([]byte(tagsJSON), &s.Tags)
			current = &s
		}

		// Append event if present
		if eventTimestamp.Valid {
			current.Events = append(current.Events, types.Event{
				Timestamp: eventTimestamp.String,
				Type:      eventType.String,
				Content:   eventContent.String,
			})
		}
	}
	if current != nil {
		sessions = append(sessions, *current)
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

// --- IDE Activity Methods ---

// RecordIDEActivity saves an IDE event to the ide_activity table.
func (s *SQLiteStore) RecordIDEActivity(activity types.IDEActivity) error {
	metaJSON, _ := json.Marshal(activity.Metadata)
	_, err := s.db.Exec(`
		INSERT INTO ide_activity (ide_name, project_name, project_path, activity_type, file_path, language, metadata, content_snapshot, executed_at, session_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, activity.IDEName, activity.ProjectName, activity.ProjectPath, activity.ActivityType,
		activity.FilePath, activity.Language, string(metaJSON), activity.ContentSnapshot,
		activity.ExecutedAt, activity.SessionID)
	return err
}

// QueryIDEActivity retrieves IDE activities with optional filters.
func (s *SQLiteStore) QueryIDEActivity(filter types.IDEActivityFilter) ([]types.IDEActivity, error) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	query := "SELECT id, ide_name, project_name, project_path, activity_type, file_path, language, metadata, content_snapshot, executed_at, session_id FROM ide_activity WHERE 1=1"
	args := []interface{}{}

	if filter.IDEName != "" {
		query += " AND ide_name = ?"
		args = append(args, filter.IDEName)
	}
	if filter.ProjectPath != "" {
		query += " AND project_path = ?"
		args = append(args, filter.ProjectPath)
	}
	if filter.ActivityType != "" {
		query += " AND activity_type = ?"
		args = append(args, filter.ActivityType)
	}
	if filter.FilePath != "" {
		query += " AND file_path LIKE ?"
		args = append(args, "%"+filter.FilePath+"%")
	}
	if filter.Language != "" {
		query += " AND language = ?"
		args = append(args, filter.Language)
	}

	query += " ORDER BY executed_at DESC LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []types.IDEActivity
	for rows.Next() {
		var a types.IDEActivity
		var metaJSON string
		if err := rows.Scan(&a.ID, &a.IDEName, &a.ProjectName, &a.ProjectPath,
			&a.ActivityType, &a.FilePath, &a.Language, &metaJSON, &a.ContentSnapshot,
			&a.ExecutedAt, &a.SessionID); err != nil {
			continue
		}
		json.Unmarshal([]byte(metaJSON), &a.Metadata)
		if a.Metadata == nil {
			a.Metadata = make(map[string]any)
		}
		activities = append(activities, a)
	}
	return activities, nil
}

// GetIDEProjects returns all tracked IDE projects.
func (s *SQLiteStore) GetIDEProjects() ([]types.IDEProject, error) {
	rows, err := s.db.Query(`
		SELECT project_name, project_path, ide_name, MAX(executed_at), COUNT(*)
		FROM ide_activity
		WHERE project_path != ''
		GROUP BY project_path
		ORDER BY MAX(executed_at) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []types.IDEProject
	for rows.Next() {
		var p types.IDEProject
		if err := rows.Scan(&p.Name, &p.Path, &p.IDEName, &p.LastActivity, &p.EventCount); err != nil {
			continue
		}
		p.IsRecording = s.isProjectRecording(p.IDEName, p.Path)
		projects = append(projects, p)
	}
	return projects, nil
}

// GetIDEPermission returns permission settings for an IDE/project combo.
func (s *SQLiteStore) GetIDEPermission(ideName, projectPath string) (types.IDEPermission, error) {
	p := types.IDEPermission{IDEName: ideName, ProjectPath: projectPath}
	row := s.db.QueryRow(`
		SELECT recording_enabled, file_recording, terminal_recording, ai_recording, last_toggled
		FROM ide_permissions WHERE ide_name = ? AND project_path = ?
	`, ideName, projectPath)
	err := row.Scan(&p.RecordingEnabled, &p.FileRecording, &p.TerminalRecording, &p.AiRecording, &p.LastToggled)
	if err != nil {
		// Also try wildcard
		row = s.db.QueryRow(`
			SELECT recording_enabled, file_recording, terminal_recording, ai_recording, last_toggled
			FROM ide_permissions WHERE ide_name = ? AND project_path = '*'
		`, ideName)
		err2 := row.Scan(&p.RecordingEnabled, &p.FileRecording, &p.TerminalRecording, &p.AiRecording, &p.LastToggled)
		if err2 != nil {
			return p, err // return original error
		}
		p.ProjectPath = "*"
	}
	return p, nil
}

// SetIDEPermission saves permission settings for an IDE/project combo.
func (s *SQLiteStore) SetIDEPermission(perm types.IDEPermission) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO ide_permissions (ide_name, project_path, recording_enabled, file_recording, terminal_recording, ai_recording, last_toggled)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, perm.IDEName, perm.ProjectPath,
		boolToInt(perm.RecordingEnabled), boolToInt(perm.FileRecording),
		boolToInt(perm.TerminalRecording), boolToInt(perm.AiRecording))
	return err
}

// GetAllIDEPermissions returns all stored permissions.
func (s *SQLiteStore) GetAllIDEPermissions() ([]types.IDEPermission, error) {
	rows, err := s.db.Query("SELECT ide_name, project_path, recording_enabled, file_recording, terminal_recording, ai_recording, last_toggled FROM ide_permissions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []types.IDEPermission
	for rows.Next() {
		var p types.IDEPermission
		if err := rows.Scan(&p.IDEName, &p.ProjectPath, &p.RecordingEnabled, &p.FileRecording, &p.TerminalRecording, &p.AiRecording, &p.LastToggled); err != nil {
			continue
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// GetDistinctIDENames returns all unique IDE names that have recorded activity.
func (s *SQLiteStore) GetDistinctIDENames() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT ide_name FROM ide_activity UNION SELECT DISTINCT ide_name FROM ide_permissions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// GetActivityCount returns total IDE activity count.
func (s *SQLiteStore) GetActivityCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM ide_activity").Scan(&count)
	return count, err
}

func (s *SQLiteStore) isProjectRecording(ideName, projectPath string) bool {
	p, err := s.GetIDEPermission(ideName, projectPath)
	if err != nil {
		return false
	}
	return p.RecordingEnabled
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
