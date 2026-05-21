package shellhistory

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Oridjinnn/Rewind/internal/redact"
	"github.com/Oridjinnn/Rewind/internal/storage"
)

// ImportConfig holds import configuration.
type ImportConfig struct {
	ShellType  string
	CustomPath string
	Store      storage.ShellHistoryStorage
}

// ImportFromShell imports shell history from a shell history file.
func ImportFromShell(cfg ImportConfig) (int, error) {
	historyPath := cfg.CustomPath
	if historyPath == "" {
		paths := DetectedShellHistoryPath()
		historyPath = paths[cfg.ShellType]
		if historyPath == "" {
			historyPath = paths["common"]
		}
	}

	// Check if file exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("history file not found: %s", historyPath)
	}

	commands, err := ParseHistoryFile(historyPath, cfg.ShellType)
	if err != nil {
		return 0, fmt.Errorf("failed to parse history file: %w", err)
	}

	return cfg.Store.ImportHistory(commands)
}

// ParseHistoryFile parses a shell history file and extracts commands.
func ParseHistoryFile(path string, shellType string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string

	switch shellType {
	case "zsh":
		commands, err = parseZshHistory(file)
	case "fish":
		commands, err = parseFishHistory(path)
	default: // bash and others
		commands, err = parseBashHistory(file)
	}

	return commands, err
}

// parseBashHistory parses .bash_history (simple one-command-per-line format).
func parseBashHistory(file *os.File) ([]string, error) {
	var commands []string
	scanner := bufio.NewScanner(file)
	// Increase buffer for long commands
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Apply secret redaction before storing
		cmd := redact.RedactCommand(line)
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands, scanner.Err()
}

// parseZshHistory parses .zsh_history (may include timestamps and extended format).
func parseZshHistory(file *os.File) ([]string, error) {
	var commands []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		cmd := extractZshCommand(line)
		if cmd != "" {
			// Apply secret redaction before storing
			redacted := redact.RedactCommand(cmd)
			if redacted != "" {
				commands = append(commands, redacted)
			}
		}
	}

	return commands, scanner.Err()
}

// extractZshCommand extracts the actual command from a zsh history line.
// Zsh history format: ": timestamp:0;command"
func extractZshCommand(line string) string {
	// Extended format: ": 1234567890:0;command"
	if strings.HasPrefix(line, ":") {
		// Find the semicolon that separates metadata from command
		if idx := strings.Index(line, ";"); idx != -1 {
			cmd := strings.TrimSpace(line[idx+1:])
			if cmd != "" {
				return cmd
			}
		}
	}

	// Plain format (fallback)
	cmd := strings.TrimSpace(line)
	if cmd != "" && !strings.HasPrefix(cmd, "#") {
		return cmd
	}

	return ""
}

// parseFishHistory parses fish shell history (YAML-like format).
func parseFishHistory(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- cmd:") {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "- cmd:"))
			if cmd != "" {
				// Apply secret redaction before storing
				redacted := redact.RedactCommand(cmd)
				if redacted != "" {
					commands = append(commands, redacted)
				}
			}
		}
	}

	return commands, scanner.Err()
}

// PrintSupportedFormats prints information about supported history formats.
func PrintSupportedFormats() {
	fmt.Println("")
	fmt.Println("Supported Shell History Formats")
	fmt.Println("===============================")
	fmt.Println("  bash  - ~/.bash_history (plain text, one command per line)")
	fmt.Println("  zsh   - ~/.zsh_history (extended format with timestamps)")
	fmt.Println("  fish  - ~/.local/share/fish/fish_history (YAML format)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  rewind import-history [shell-type]")
	fmt.Println("  rewind import-history bash")
	fmt.Println("  rewind import-history /path/to/custom/history/file")
	fmt.Println("")
}

// AutoDetectAndImport tries to detect the shell and import its history.
func AutoDetectAndImport(store storage.ShellHistoryStorage) (map[string]int, error) {
	results := make(map[string]int)

	for shell, path := range DetectedShellHistoryPath() {
		if shell == "common" {
			continue
		}

		if _, err := os.Stat(path); err != nil {
			continue
		}

		cfg := ImportConfig{
			ShellType:  shell,
			CustomPath: path,
			Store:      store,
		}

		count, err := ImportFromShell(cfg)
		if err != nil {
			fmt.Printf("warning: failed to import %s history: %v\n", shell, err)
			continue
		}

		results[shell] = count
		fmt.Printf("Imported %d commands from %s history (%s)\n", count, shell, filepath.Base(path))
	}

	return results, nil
}