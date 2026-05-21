package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/redact"
	"github.com/habeldavidson007-glitch/rewind/internal/storage"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

// GetSetupScript returns the shell setup code that user should add to their rc file
func GetSetupScript(shell string) string {
	rewindPath, _ := os.Executable()

	switch shell {
	case "bash":
		return fmt.Sprintf(`
# Rewind auto-record setup for bash
export REWIND_ENABLED=true
export REWIND_PATH="%s"

# Hook into every command via DEBUG trap
trap 'rewind_record_command "$BASH_COMMAND"' DEBUG

rewind_record_command() {
	local cmd="$1"
	# Skip rewind commands themselves, and internal shell operations
	if [[ "$cmd" == rewind\ * ]] || [[ "$cmd" == "echo"* ]] || [[ "$cmd" =~ ^[[:space:]]*$ ]]; then
		return
	fi
	# Store command in env for postexec hook
	REWIND_LAST_CMD="$cmd"
	REWIND_CMD_START=$(date +%%s%%N)
}

# Use PROMPT_COMMAND for post-execution tracking
if [[ -z "$PROMPT_COMMAND" ]]; then
	PROMPT_COMMAND='rewind_track_exit'
else
	PROMPT_COMMAND="rewind_track_exit; $PROMPT_COMMAND"
fi

rewind_track_exit() {
	local exit_code=$?
	if [[ ! -z "$REWIND_ENABLED" ]] && [[ ! -z "$REWIND_LAST_CMD" ]]; then
		"$REWIND_PATH" track-command "$REWIND_LAST_CMD" $exit_code &
		unset REWIND_LAST_CMD
	fi
}
`, rewindPath)

	case "zsh":
		return fmt.Sprintf(`
# Rewind auto-record setup for zsh
export REWIND_ENABLED=true
export REWIND_PATH="%s"

# Pre-command hook
precmd_rewind() {
	local exit_code=$?
	if [[ ! -z "$REWIND_LAST_CMD" ]] && [[ ! -z "$REWIND_ENABLED" ]]; then
		"$REWIND_PATH" track-command "$REWIND_LAST_CMD" $exit_code &
		unset REWIND_LAST_CMD
	fi
}

# Post-execution hook
preexec_rewind() {
	REWIND_LAST_CMD="$1"
}

# Add hooks if not already present
if [[ ! " $precmd_functions " =~ "precmd_rewind" ]]; then
	precmd_functions+=(precmd_rewind)
fi

if [[ ! " $preexec_functions " =~ "preexec_rewind" ]]; then
	preexec_functions+=(preexec_rewind)
fi
`, rewindPath)

	case "fish":
		return fmt.Sprintf(`
# Rewind auto-record setup for fish
set -gx REWIND_ENABLED true
set -gx REWIND_PATH "%s"

# Fish event listener for command execution
function rewind_postexec --on-event fish_postexec
	if test -n "$REWIND_ENABLED"
		"$REWIND_PATH" track-command "$argv[1]" $status &
	end
end
`, rewindPath)

	default:
		return "# Unknown shell"
	}
}

// TrackCommand records the command that was just executed
func TrackCommand(cmdStr string, exitCode int) error {
	// Apply secret redaction if enabled
	cmdStr = redact.RedactCommand(cmdStr)
	if cmdStr == "" {
		// Command contained secrets and REWIND_REDACT=skip mode is active
		return nil
	}

	// Parse command and args
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := parts[0]
	args := parts[1:]

	// Create a minimal session record
	// Note: stdout/stderr are already captured by shell, so we just record metadata
	session, err := createMinimalSession(cmd, args, exitCode)
	if err != nil {
		return err
	}

	// Convert SessionRecord to types.Session for storage
	ts := types.Session{
		ID:        session.ID,
		Command:   session.Command,
		Title:     session.Title,
		Summary:   session.Summary,
		StartedAt: session.StartedAt,
		EndedAt:   session.EndedAt,
		ExitCode:  session.ExitCode,
	}
	for _, e := range session.Events {
		ts.Events = append(ts.Events, types.Event{
			Timestamp: e.Timestamp,
			Type:      e.Type,
			Content:   e.Content,
		})
	}

	// Save session
	sessionPath := filepath.Join("sessions", ts.ID+".json")
	err = storage.SaveSession(ts, sessionPath)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// createMinimalSession creates a session record for auto-tracked command
func createMinimalSession(cmd string, args []string, exitCode int) (*SessionRecord, error) {
	now := time.Now()
	sessionID := fmt.Sprintf("%d_%s", now.UnixNano(), generateShortID())

	// Reconstruct full command
	fullCmd := cmd
	if len(args) > 0 {
		fullCmd = cmd + " " + strings.Join(args, " ")
	}

	// Try to capture output if still available (this is best-effort)
	// Since we're hooking after execution, output is already printed to terminal
	// We'll just record that the command ran
	return &SessionRecord{
		ID:        sessionID,
		Command:   fullCmd,
		Title:     cmd,
		Summary:   fmt.Sprintf("Auto-tracked: %s", fullCmd),
		StartedAt: now.Format(time.RFC3339),
		EndedAt:   now.Format(time.RFC3339),
		ExitCode:  exitCode,
		Events: []EventRecord{
			{
				Timestamp: now.Format(time.RFC3339Nano),
				Type:      "command",
				Content:   fullCmd,
			},
			{
				Timestamp: now.Format(time.RFC3339Nano),
				Type:      "exit",
				Content:   fmt.Sprintf("%d", exitCode),
			},
		},
	}, nil
}

// generateShortID creates a short random ID
func generateShortID() string {
	const charset = "0123456789abcdef"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Simple structs for shell integration (minimal)
type EventRecord struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

type SessionRecord struct {
	ID        string        `json:"id"`
	Command   string        `json:"command"`
	Title     string        `json:"title,omitempty"`
	Summary   string        `json:"summary,omitempty"`
	StartedAt string        `json:"started_at"`
	EndedAt   string        `json:"ended_at"`
	ExitCode  int           `json:"exit_code"`
	Events    []EventRecord `json:"events"`
}

// PrintSetupInstructions shows user how to install the shell hook
func PrintSetupInstructions(shell string) {
	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  Rewind Shell Integration Setup                                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Printf("Shell detected: %s\n\n", shell)

	rcFile := getRCFile(shell)
	fmt.Printf("1. Add this to your %s:\n\n", rcFile)
	fmt.Println(GetSetupScript(shell))
	fmt.Printf("\n2. Reload your shell:\n")
	fmt.Printf("   source ~/%s\n\n", rcFile)
	fmt.Println("3. Done! Every command will now be auto-recorded in ./sessions/")
	fmt.Println("")
	fmt.Println("Disable anytime with: export REWIND_ENABLED=false")
	fmt.Println("")
}

// getRCFile returns the rc file for the given shell
func getRCFile(shell string) string {
	switch shell {
	case "bash":
		return ".bashrc"
	case "zsh":
		return ".zshrc"
	case "fish":
		return ".config/fish/config.fish"
	default:
		return ".bashrc"
	}
}

// DetectShell detects which shell is running
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}

	// Extract shell name from path
	parts := strings.Split(shell, "/")
	if len(parts) > 0 {
		shellName := parts[len(parts)-1]
		if shellName == "sh" {
			return "bash"
		}
		return shellName
	}

	return "bash"
}
