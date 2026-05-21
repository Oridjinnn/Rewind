package shellhistory

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/habeldavidson007-glitch/rewind/internal/storage"
)

// Manager handles shell history operations using any ShellHistoryStorage backend.
type Manager struct {
	store storage.ShellHistoryStorage
}

// NewManager creates a new shell history manager with the given storage backend.
func NewManager(store storage.ShellHistoryStorage) *Manager {
	return &Manager{store: store}
}

// TrackCommand records a shell command with metadata.
func (m *Manager) TrackCommand(cmd string, exitCode int, sessionID string) error {
	workDir, _ := os.Getwd()
	return m.store.SaveEntry(cmd, exitCode, workDir, sessionID)
}

// ViewHistory returns recent shell history entries.
func (m *Manager) ViewHistory(limit int) ([]storage.ShellHistoryEntry, error) {
	return m.store.GetHistory(limit, 0)
}

// SearchHistory searches shell history by command text.
func (m *Manager) SearchHistory(query string, limit int) ([]storage.ShellHistoryEntry, error) {
	return m.store.SearchHistory(query, limit)
}

// ShowStats displays shell history statistics.
func (m *Manager) ShowStats() (storage.HistoryStats, error) {
	return m.store.GetHistoryStats()
}

// PrintHistory prints shell history entries in a readable format.
func PrintHistory(entries []storage.ShellHistoryEntry) {
	fmt.Println("")
	fmt.Println("SHELL HISTORY")
	fmt.Println("=============")
	for _, e := range entries {
		exitStr := ""
		if e.ExitCode != 0 {
			exitStr = fmt.Sprintf(" [exit: %d]", e.ExitCode)
		}
		dirStr := ""
		if e.WorkingDir != "" {
			dirStr = fmt.Sprintf(" (%s)", e.WorkingDir)
		}
		fmt.Printf("  %s  %s%s%s\n", e.ExecutedAt, e.Command, exitStr, dirStr)
	}
	fmt.Println("")
}

// PrintStats prints history statistics in a readable format.
func PrintStats(stats storage.HistoryStats) {
	fmt.Println("")
	fmt.Println("HISTORY STATISTICS")
	fmt.Println("==================")
	fmt.Printf("Total commands:    %d\n", stats.TotalCommands)
	fmt.Printf("Unique commands:   %d\n", stats.UniqueCommands)
	fmt.Println("")
	fmt.Println("Top Commands:")
	fmt.Println("-------------")
	for _, cc := range stats.TopCommands {
		fmt.Printf("  %-40s %d\n", cc.Command, cc.Count)
	}
	fmt.Println("")
}

// DetectedShellHistoryPath returns common shell history file paths.
func DetectedShellHistoryPath() map[string]string {
	home, _ := os.UserHomeDir()
	return map[string]string{
		"bash":   filepath.Join(home, ".bash_history"),
		"zsh":    filepath.Join(home, ".zsh_history"),
		"fish":   filepath.Join(home, ".local", "share", "fish", "fish_history"),
		"common": filepath.Join(home, ".bash_history"),
	}
}