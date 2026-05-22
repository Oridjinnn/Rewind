package ide

import (
	"fmt"
	"strings"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

// EnableRecording enables recording for an IDE (optionally per project).
func EnableRecording(r Recorder, ideName, projectPath string) error {
	ideName = normalizeIDEName(ideName)
	if !ValidIDEs[ideName] {
		return fmt.Errorf("unsupported IDE: %s (valid: %s)", ideName, listValidIDEs())
	}
	if projectPath == "" {
		projectPath = "*"
	}
	return r.SetPermission(types.IDEPermission{
		IDEName:           ideName,
		ProjectPath:       projectPath,
		RecordingEnabled:  true,
		FileRecording:     true,
		TerminalRecording: true,
		AiRecording:       true,
	})
}

// DisableRecording disables recording for an IDE.
func DisableRecording(r Recorder, ideName, projectPath string) error {
	ideName = normalizeIDEName(ideName)
	if projectPath == "" {
		projectPath = "*"
	}
	return r.SetPermission(types.IDEPermission{
		IDEName:           ideName,
		ProjectPath:       projectPath,
		RecordingEnabled:  false,
		FileRecording:     false,
		TerminalRecording: false,
		AiRecording:       false,
	})
}

// SetGranularPermissions sets specific permission flags.
func SetGranularPermissions(r Recorder, perm types.IDEPermission) error {
	perm.IDEName = normalizeIDEName(perm.IDEName)
	if !ValidIDEs[perm.IDEName] {
		return fmt.Errorf("unsupported IDE: %s", perm.IDEName)
	}
	if perm.ProjectPath == "" {
		perm.ProjectPath = "*"
	}
	return r.SetPermission(perm)
}

// PrintPermissions displays current permissions in a readable format.
func PrintPermissions(perms []types.IDEPermission) {
	if len(perms) == 0 {
		fmt.Println("No permissions configured yet.")
		fmt.Println("Use 'rewind ide permissions <ide> on' to enable recording.")
		return
	}
	fmt.Println("")
	fmt.Println("IDE PERMISSIONS")
	fmt.Println("================")
	for _, p := range perms {
		status := "OFF"
		if p.RecordingEnabled {
			status = "ON"
		}
		fmt.Printf("  %-20s | project: %-30s | [%s]\n", IDEToHumanName(p.IDEName), p.ProjectPath, status)
		fmt.Printf("    file: %v  terminal: %v  ai: %v\n", p.FileRecording, p.TerminalRecording, p.AiRecording)
	}
	fmt.Println("")
}

// PrintActivity prints IDE activities in a readable format.
func PrintActivity(activities []types.IDEActivity) {
	if len(activities) == 0 {
		fmt.Println("No IDE activity recorded yet.")
		return
	}
	fmt.Println("")
	fmt.Println("IDE ACTIVITY")
	fmt.Println("=============")
	for _, a := range activities {
		ts := a.ExecutedAt
		if len(ts) > 19 {
			ts = ts[:19]
		}
		file := a.FilePath
		if file == "" {
			file = "-"
		}
		lang := a.Language
		if lang != "" {
			lang = " [" + lang + "]"
		}
		fmt.Printf("  %s | %-20s | %-20s | %s%s\n",
			ts, IDEToHumanName(a.IDEName), ActivityToHumanName(a.ActivityType), file, lang)
	}
	fmt.Println("")
}

// PrintProjects lists tracked IDE projects.
func PrintProjects(projects []types.IDEProject) {
	if len(projects) == 0 {
		fmt.Println("No IDE projects tracked yet.")
		return
	}
	fmt.Println("")
	fmt.Println("IDE PROJECTS")
	fmt.Println("=============")
	for _, p := range projects {
		recording := ""
		if p.IsRecording {
			recording = " [RECORDING]"
		}
		fmt.Printf("  %-25s | %s | %d events | %s%s\n",
			p.Name, IDEToHumanName(p.IDEName), p.EventCount, p.LastActivity[:19], recording)
	}
	fmt.Println("")
}

func normalizeIDEName(name string) string {
	aliases := map[string]string{
		"vs code":     "vscode",
		"vscode":      "vscode",
		"cursor":       "cursor",
		"intellij":     "intellij-idea",
		"idea":         "intellij-idea",
		"intellij idea":"intellij-idea",
		"goland":       "goland",
		"pycharm":      "pycharm",
		"webstorm":     "webstorm",
		"android":      "android-studio",
		"android studio":"android-studio",
		"eclipse":      "eclipse",
		"sublime":      "sublime",
		"subl":         "sublime",
		"nvim":         "nvim",
		"neovim":       "nvim",
		"vim":          "vim",
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	if alias, ok := aliases[lower]; ok {
		return alias
	}
	return lower
}

func listValidIDEs() string {
	var names []string
	for name := range ValidIDEs {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}