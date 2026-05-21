package ide

import (
	"encoding/json"
	"fmt"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

const (
	ProtocolVersion = "rewind-ide-v1"
	ServerPort      = 9876
)

// ValidIDEs lists all IDE names supported by the protocol.
var ValidIDEs = map[string]bool{
	"vscode":         true,
	"cursor":          true,
	"intellij-idea":   true,
	"goland":          true,
	"pycharm":         true,
	"webstorm":        true,
	"android-studio":  true,
	"eclipse":         true,
	"sublime":         true,
	"nvim":            true,
	"vim":             true,
}

// ValidActivityTypes lists all recognized activity types.
var ValidActivityTypes = map[string]bool{
	"file_open":         true,
	"file_save":          true,
	"file_edit":          true,
	"file_close":         true,
	"file_delete":        true,
	"file_create":        true,
	"terminal_cmd":       true,
	"build_start":        true,
	"build_end":          true,
	"build_error":        true,
	"test_run":           true,
	"test_pass":          true,
	"test_fail":          true,
	"git_commit":         true,
	"git_push":           true,
	"git_pull":           true,
	"git_branch":         true,
	"git_stash":          true,
	"debug_start":        true,
	"debug_breakpoint":   true,
	"debug_step":         true,
	"debug_end":          true,
	"ai_chat":            true,
	"ai_completion":      true,
	"ai_accept":          true,
	"ai_reject":          true,
	"refactor":           true,
	"search":             true,
	"run_config":         true,
	"dependency_change":  true,
}

// ParseEvent deserializes and validates an IDE event from JSON.
func ParseEvent(data []byte) (types.IDEEvent, error) {
	var event types.IDEEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return event, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := ValidateEvent(event); err != nil {
		return event, err
	}

	return event, nil
}

// ValidateEvent checks all required fields and valid values.
func ValidateEvent(event types.IDEEvent) error {
	if event.Protocol != ProtocolVersion {
		return fmt.Errorf("unsupported protocol: %s (expected %s)", event.Protocol, ProtocolVersion)
	}
	if event.IDE == "" {
		return fmt.Errorf("ide name is required")
	}
	if !ValidIDEs[event.IDE] {
		return fmt.Errorf("unsupported IDE: %s", event.IDE)
	}
	if event.Event.Type == "" {
		return fmt.Errorf("event type is required")
	}
	if !ValidActivityTypes[event.Event.Type] {
		return fmt.Errorf("unsupported activity type: %s", event.Event.Type)
	}
	return nil
}

// IDEToHumanName maps IDE machine names to display names.
func IDEToHumanName(ide string) string {
	names := map[string]string{
		"vscode":         "VS Code",
		"cursor":          "Cursor",
		"intellij-idea":   "IntelliJ IDEA",
		"goland":          "GoLand",
		"pycharm":         "PyCharm",
		"webstorm":        "WebStorm",
		"android-studio":  "Android Studio",
		"eclipse":         "Eclipse",
		"sublime":         "Sublime Text",
		"nvim":            "Neovim",
		"vim":             "Vim",
	}
	if name, ok := names[ide]; ok {
		return name
	}
	return ide
}

// ActivityToHumanName maps activity types to display names.
func ActivityToHumanName(activity string) string {
	names := map[string]string{
		"file_open":         "Opened file",
		"file_save":          "Saved file",
		"file_edit":          "Edited file",
		"file_close":         "Closed file",
		"file_delete":        "Deleted file",
		"file_create":        "Created file",
		"terminal_cmd":       "Terminal command",
		"build_start":        "Build started",
		"build_end":          "Build completed",
		"build_error":        "Build error",
		"test_run":           "Tests ran",
		"test_pass":          "Tests passed",
		"test_fail":          "Tests failed",
		"git_commit":         "Git commit",
		"git_push":           "Git push",
		"git_pull":           "Git pull",
		"git_branch":         "Git branch",
		"git_stash":          "Git stash",
		"debug_start":        "Debug started",
		"debug_breakpoint":   "Breakpoint hit",
		"debug_step":         "Step debug",
		"debug_end":          "Debug ended",
		"ai_chat":            "AI chat",
		"ai_completion":      "AI completion",
		"ai_accept":          "AI accept",
		"ai_reject":          "AI reject",
		"refactor":           "Refactored",
		"search":             "Searched",
		"run_config":         "Run configuration",
		"dependency_change":  "Dependency changed",
	}
	if name, ok := names[activity]; ok {
		return name
	}
	return activity
}