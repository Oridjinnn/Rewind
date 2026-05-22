package ide

import (
	"fmt"
	"time"

	"github.com/Oridjinnn/Rewind/internal/shellhistory"
	"github.com/Oridjinnn/Rewind/internal/storage"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

// Bridge connects IDE activity to shell history and sessions.
// This enables cross-referencing IDE actions with terminal commands.
type Bridge struct {
	recorder      *SQLiteRecorder
	historyMgr    *shellhistory.Manager
	store         *storage.SQLiteStore
}

// NewBridge creates a bridge between IDE, shell history, and sessions.
func NewBridge(recorder *SQLiteRecorder, store *storage.SQLiteStore) *Bridge {
	return &Bridge{
		recorder:   recorder,
		historyMgr: shellhistory.NewManager(store),
		store:      store,
	}
}

// RecordTerminalCmd records a terminal command that originated from an IDE.
// This links the shell history entry to both the IDE session and the terminal session.
func (b *Bridge) RecordTerminalCmd(ideName, projectPath, cmd string, exitCode int, sessionID string) error {
	// Record in shell_history (which also auto-creates a session if needed)
	if err := b.historyMgr.TrackCommand(cmd, exitCode, sessionID); err != nil {
		return err
	}

	// Record as IDE activity
	event := types.IDEEvent{
		Protocol:  ProtocolVersion,
		IDE:       ideName,
		Project:   projectPath,
		ProjectPath: projectPath,
		Event: types.IDEEventData{
			Type:      "terminal_cmd",
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			ExitCode:  exitCode,
			Message:   cmd,
			SessionID: sessionID,
		},
	}

	return b.recorder.RecordEvent(event)
}

// GetProjectContext returns combined IDE + shell activity for a project.
// This gives AI models richer context about what happened across IDE and terminal.
func (b *Bridge) GetProjectContext(projectPath string, limit int) (map[string]interface{}, error) {
	// Get IDE activities
	ideFilter := types.IDEActivityFilter{
		ProjectPath: projectPath,
		Limit:       limit,
	}
	ideActivities, _ := b.recorder.QueryActivity(ideFilter)

	// Get shell history entries linked to this project
	// (via session_id matching)
	shellEntries, _ := b.store.GetHistory(limit, 0)

	// Get sessions for this project
	sessions, _ := b.store.SearchSessions(projectPath)

	result := map[string]interface{}{
		"project":        projectPath,
		"ide_activities": ideActivities,
		"shell_commands": shellEntries,
		"sessions":       sessions,
		"total_ide":      len(ideActivities),
		"total_shell":    len(shellEntries),
	}

	return result, nil
}

// RecallAcrossIDE searches across IDE activity and shell history.
func (b *Bridge) RecallAcrossIDE(query string, limit int) ([]string, error) {
	var results []string

	// Search IDE activities via keyword (proper multi-column recall)
	ideActivities, _ := b.recorder.QueryActivity(types.IDEActivityFilter{
		Keyword: query,
		Limit:   limit,
	})

	for _, a := range ideActivities {
		results = append(results, fmt.Sprintf("[IDE:%s] %s - %s (%s)",
			IDEToHumanName(a.IDEName), ActivityToHumanName(a.ActivityType), a.FilePath, a.ExecutedAt[:19]))
	}

	// Search shell history
	shellEntries, _ := b.store.SearchHistory(query, limit)
	for _, e := range shellEntries {
		results = append(results, fmt.Sprintf("[SHELL] %s (exit: %d) (%s)",
			e.Command, e.ExitCode, e.ExecutedAt[:19]))
	}

	return results, nil
}

// PrintBridgeSummary shows a combined IDE + shell overview.
func (b *Bridge) PrintBridgeSummary(projectPath string, limit int) {
	fmt.Println("")
	fmt.Println("BRIDGE SUMMARY: IDE + SHELL")
	fmt.Println("============================")
	fmt.Printf("Project: %s\n", projectPath)

	ctx, err := b.GetProjectContext(projectPath, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("IDE events:  %d\n", ctx["total_ide"])
	fmt.Printf("Shell cmds:  %d\n", ctx["total_shell"])
	fmt.Println("")

	// Print latest IDE activities
	if activities, ok := ctx["ide_activities"].([]types.IDEActivity); ok && len(activities) > 0 {
		fmt.Println("Recent IDE Activity:")
		for i, a := range activities {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(activities)-5)
				break
			}
			fmt.Printf("  %s | %s | %s\n",
				a.ExecutedAt[:19], ActivityToHumanName(a.ActivityType), a.FilePath)
		}
		fmt.Println("")
	}

	// Print latest shell commands
	if entries, ok := ctx["shell_commands"].([]storage.ShellHistoryEntry); ok && len(entries) > 0 {
		fmt.Println("Recent Shell Commands:")
		for i, e := range entries {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(entries)-5)
				break
			}
			fmt.Printf("  %s | %s (exit: %d)\n", e.ExecutedAt[:19], e.Command, e.ExitCode)
		}
		fmt.Println("")
	}
}

// Close cleans up the bridge resources.
func (b *Bridge) Close() error {
	return b.recorder.Close()
}