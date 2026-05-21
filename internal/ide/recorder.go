package ide

import (
	"fmt"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/storage"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

// SQLiteRecorder implements Recorder using the SQLiteStore.
type SQLiteRecorder struct {
	store    *storage.SQLiteStore
	serverRunning bool
	serverPort    int
}

// NewSQLiteRecorder creates a recorder backed by SQLiteStore.
func NewSQLiteRecorder(store *storage.SQLiteStore) *SQLiteRecorder {
	return &SQLiteRecorder{store: store}
}

// SetServerState updates whether the HTTP server is running.
func (r *SQLiteRecorder) SetServerState(running bool, port int) {
	r.serverRunning = running
	r.serverPort = port
}

// RecordEvent converts an IDEEvent into IDEActivity and stores it.
func (r *SQLiteRecorder) RecordEvent(event types.IDEEvent) error {
	if event.Event.Type == "" {
		return fmt.Errorf("event type is required")
	}

	// Check permission
	ok, err := r.CheckPermission(event.IDE, event.ProjectPath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("recording disabled for %s on %s", event.IDE, event.ProjectPath)
	}

	// Get permission details for content filtering
	perm, _ := r.GetPermission(event.IDE, event.ProjectPath)

	ts := event.Event.Timestamp
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339Nano)
	}

	activity := types.IDEActivity{
		IDEName:      event.IDE,
		ProjectName:  event.Project,
		ProjectPath:  event.ProjectPath,
		ActivityType: event.Event.Type,
		FilePath:     event.Event.File,
		Language:     event.Event.Language,
		Metadata:     event.Event.Metadata,
		ExecutedAt:   ts,
		SessionID:    event.Event.SessionID,
	}

	// Only include content snapshots if file_recording is enabled
	if perm.FileRecording && event.Event.ContentSnapshot != "" {
		activity.ContentSnapshot = event.Event.ContentSnapshot
	}

	// Also enrich metadata with numeric fields
	if activity.Metadata == nil {
		activity.Metadata = make(map[string]any)
	}
	if event.Event.LinesAdded > 0 {
		activity.Metadata["lines_added"] = event.Event.LinesAdded
	}
	if event.Event.LinesRemoved > 0 {
		activity.Metadata["lines_removed"] = event.Event.LinesRemoved
	}
	if event.Event.DurationMs > 0 {
		activity.Metadata["duration_ms"] = event.Event.DurationMs
	}
	if event.Event.ExitCode != 0 {
		activity.Metadata["exit_code"] = event.Event.ExitCode
	}
	if event.Event.CursorLine > 0 {
		activity.Metadata["cursor_line"] = event.Event.CursorLine
		activity.Metadata["cursor_column"] = event.Event.CursorColumn
	}
	return r.store.RecordIDEActivity(activity)
}

// RecordBatch records multiple IDE events.
func (r *SQLiteRecorder) RecordBatch(events []types.IDEEvent) error {
	for _, event := range events {
		if err := r.RecordEvent(event); err != nil {
			// Log but continue
			continue
		}
	}
	return nil
}

// CheckPermission verifies recording is enabled.
func (r *SQLiteRecorder) CheckPermission(ideName, projectPath string) (bool, error) {
	perm, err := r.GetPermission(ideName, projectPath)
	if err != nil {
		return false, err
	}
	return perm.RecordingEnabled, nil
}

// GetPermission returns full permission for IDE/project.
func (r *SQLiteRecorder) GetPermission(ideName, projectPath string) (types.IDEPermission, error) {
	return r.store.GetIDEPermission(ideName, projectPath)
}

// SetPermission updates permission settings.
func (r *SQLiteRecorder) SetPermission(perm types.IDEPermission) error {
	return r.store.SetIDEPermission(perm)
}

// GetStatus returns the current IDE recording status.
func (r *SQLiteRecorder) GetStatus() (types.IDEStatus, error) {
	count, _ := r.store.GetActivityCount()

	ides, err := r.store.GetDistinctIDENames()
	if err != nil {
		ides = []string{}
	}

	perms, err := r.store.GetAllIDEPermissions()
	if err != nil {
		perms = []types.IDEPermission{}
	}

	projects, err := r.store.GetIDEProjects()
	if err != nil {
		projects = []types.IDEProject{}
	}

	activeProject := ""
	if len(projects) > 0 {
		activeProject = projects[0].Name
	}

	return types.IDEStatus{
		ServerRunning: r.serverRunning,
		ServerPort:    r.serverPort,
		ConnectedIDEs: ides,
		Permissions:   perms,
		ActivityCount: count,
		ActiveProject: activeProject,
	}, nil
}

// GetProjects returns all tracked IDE projects.
func (r *SQLiteRecorder) GetProjects() ([]types.IDEProject, error) {
	return r.store.GetIDEProjects()
}

// QueryActivity returns filtered IDE activities.
func (r *SQLiteRecorder) QueryActivity(filter types.IDEActivityFilter) ([]types.IDEActivity, error) {
	return r.store.QueryIDEActivity(filter)
}

// Close cleans up the recorder.
func (r *SQLiteRecorder) Close() error {
	return r.store.Close()
}