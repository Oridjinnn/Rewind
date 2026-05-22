package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Oridjinnn/Rewind/internal/recall"
	"github.com/Oridjinnn/Rewind/internal/markdown"
	"github.com/Oridjinnn/Rewind/internal/chat"
	"github.com/Oridjinnn/Rewind/internal/score"
	"github.com/Oridjinnn/Rewind/internal/inspect"
	"github.com/Oridjinnn/Rewind/internal/export"
	"github.com/Oridjinnn/Rewind/internal/stats"
	"github.com/Oridjinnn/Rewind/internal/timeline"
	"github.com/Oridjinnn/Rewind/internal/detect"
	"github.com/Oridjinnn/Rewind/internal/diff"
	rec "github.com/Oridjinnn/Rewind/internal/recorder"
	"github.com/Oridjinnn/Rewind/internal/replay"
	"github.com/Oridjinnn/Rewind/internal/storage"
	"github.com/Oridjinnn/Rewind/internal/shell"
	"github.com/Oridjinnn/Rewind/internal/ide"
	"github.com/Oridjinnn/Rewind/internal/shellhistory"
	"github.com/Oridjinnn/Rewind/internal/web"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

// getStorage returns the appropriate storage backend.
// Uses SQLite if REWIND_DB_PATH is set or falls back to JSON.
func getStorage() (storage.Storage, error) {
	dbPath := os.Getenv("REWIND_DB_PATH")
	if dbPath == "" {
		// Default to SQLite in the current directory
		dbPath = "rewind.db"
	}

	// Check if we should force JSON fallback
	if os.Getenv("REWIND_USE_JSON") == "true" {
		return nil, fmt.Errorf("JSON storage selected")
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	return store, nil
}

// getShellHistoryStore returns the shell history storage backend.
func getShellHistoryStore() (storage.ShellHistoryStorage, error) {
	store, err := getStorage()
	if err != nil {
		return nil, err
	}
	if s, ok := store.(storage.ShellHistoryStorage); ok {
		return s, nil
	}
	return nil, fmt.Errorf("storage backend does not support shell history")
}

// loadSession loads a session by ID, prioritizing SQLite with JSON fallback.
func loadSession(id string) (types.Session, error) {
	st, err := getStorage()
	if err == nil {
		defer st.Close()
		session, err := st.LoadSession(id)
		if err == nil {
			return session, nil
		}
	}

	// Fallback to JSON file storage
	sessionPath := filepath.Join("sessions", id+".json")
	return storage.LoadSession(sessionPath)
}

// getSQLiteStore returns a raw SQLiteStore for IDE operations.
func getSQLiteStore() (*storage.SQLiteStore, error) {
	st, err := getStorage()
	if err != nil {
		return nil, err
	}
	if s, ok := st.(*storage.SQLiteStore); ok {
		return s, nil
	}
	return nil, fmt.Errorf("SQLite storage not available")
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("  rewind run <command>")
		fmt.Println("  rewind replay <session_id>")
		fmt.Println("  rewind web [port]          # Start the Rewind web UI")
		fmt.Println("  rewind setup               # Setup auto-recording for your shell")
		fmt.Println("  rewind chat <model>        # Chat with memory")
		fmt.Println("  rewind recall <query>      # Search sessions")
		fmt.Println("  rewind list                # List all sessions")
		fmt.Println("  rewind search <query>      # Search sessions by text")
		fmt.Println("  rewind migrate             # Migrate JSON sessions to SQLite")
		fmt.Println("  rewind history [limit]     # View shell history")
		fmt.Println("  rewind import-history [shell|path] # Import from shell history files")
		fmt.Println("  rewind history-stats       # Shell history statistics")
		fmt.Println("  rewind ide start [port]    # Start IDE recording server")
		fmt.Println("  rewind ide stop            # Stop IDE recording server")
		fmt.Println("  rewind ide status          # Show IDE recording status")
		fmt.Println("  rewind ide permissions [ide] [on|off]   # Manage opt-in")
		fmt.Println("  rewind ide activity [ide] [limit]       # View IDE activity")
		fmt.Println("  rewind ide projects        # List tracked IDE projects")
		fmt.Println("  rewind ide analyze [project]            # Productivity insights")
		return
	}

	command := os.Args[1]

	switch command {

	case "recall":

		if len(os.Args) < 3 {
			fmt.Println("usage: rewind recall <query>")
			return
		}

		recall.Recall(
			os.Args[2],
		)

	case "run":

		if len(os.Args) < 3 {
			fmt.Println("missing command")
			return
		}

		targetCommand := os.Args[2]
		targetArgs := os.Args[3:]

		session, err := rec.RecordCommand(
			targetCommand,
			targetArgs,
		)

		if err != nil {
			fmt.Println("record failed:", err)
			return
		}

		// Try to save to SQLite first for v0.3 consistency
		st, err := getStorage()
		if err == nil { // If SQLite storage is available
			defer st.Close()
			err = st.SaveSession(session)
			if err != nil {
				fmt.Println("save to SQLite failed, falling back to JSON:", err)
			} else {
				fmt.Println("session recorded (SQLite):", session.ID)
				return
			}
		}

		// JSON Fallback (either getStorage failed, or SQLite SaveSession failed)
		sessionPath := filepath.Join("sessions", session.ID+".json")
		storage.SaveSession(session, sessionPath) // Assuming this always works for JSON
		fmt.Println("session recorded (JSON):", session.ID)

	case "replay":

		if len(os.Args) < 3 {
			fmt.Println("missing session id")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		replay.ReplaySession(session)

	case "diff":

		if len(os.Args) < 4 {
			fmt.Println("usage:")
			fmt.Println("  rewind diff <sessionA> <sessionB>")
			return
		}

		leftID := os.Args[2]
		rightID := os.Args[3]

		leftSession, err := loadSession(leftID)
		if err != nil {
			fmt.Println("failed loading left session:", err)
			return
		}

		rightSession, err := loadSession(rightID)
		if err != nil {
			fmt.Println("failed loading right session:", err)
			return
		}

		diff.CompareSessions(
			leftSession,
			rightSession,
		)

	case "export":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind export <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		reportPath, err := export.ExportHTML(
			session,
		)

		if err != nil {
			fmt.Println("export failed:", err)
			return
		}

		fmt.Println("")
		fmt.Println("report generated:")
		fmt.Println(reportPath)
		fmt.Println("")

	case "stats":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind stats <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		stats.RenderStats(session)

	case "timeline":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind timeline <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		timeline.RenderTimeline(session)

	case "detect":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind detect <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		detect.DetectLoops(session)

	case "inspect":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind inspect <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		inspect.PrintSessionInspect(
			session,
		)

	case "score":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind score <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		score.PrintSessionScore(
			session,
		)
	case "chat":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind chat <model>")
			return
		}

		model := os.Args[2]

		chat.StartChat(
			model,
		)

	case "markdown":

		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind markdown <session_id>")
			return
		}

		sessionID := os.Args[2]
		session, err := loadSession(sessionID)
		if err != nil {
			fmt.Println("load failed:", err)
			return
		}

		err = markdown.ExportMarkdown(
			session,
		)

		if err != nil {
			fmt.Println("export failed:", err)
			return
		}

	case "web":
		port := 8080
		if len(os.Args) >= 3 {
			if parsed, err := strconv.Atoi(os.Args[2]); err == nil {
				port = parsed
			}
		}
		fmt.Printf("Starting web UI on http://localhost:%d\n", port)
		if err := web.Serve(port); err != nil {
			fmt.Println("web server error:", err)
		}

	case "list":
		// List sessions from SQLite (fallback to JSON)
		st, err := getStorage()
		if err == nil {
			defer st.Close()
			ids, listErr := st.ListSessions()
			if listErr == nil && len(ids) > 0 {
				fmt.Println("")
				fmt.Println("AVAILABLE SESSIONS (SQLite)")
				fmt.Println("---------------------------")
				for _, id := range ids {
					fmt.Println("  " + id)
				}
				fmt.Println("")
				return
			}
		}

		// JSON fallback
		files, err := storage.ListSessions("sessions")
		if err != nil {
			fmt.Println("list failed:", err)
			return
		}

		fmt.Println("")
		fmt.Println("AVAILABLE SESSIONS (JSON)")
		fmt.Println("-------------------------")

		for _, file := range files {
			name := file.Name()
			if strings.HasSuffix(name, ".json") {
				fmt.Println("  " + strings.TrimSuffix(name, ".json"))
			}
		}
		fmt.Println("")

	case "setup":
		// Shell integration setup
		shellType := shell.DetectShell()
		shell.PrintSetupInstructions(shellType)

	case "track-command":
		// Auto-track command (called by shell hooks)
		if len(os.Args) < 3 {
			fmt.Println("usage: rewind track-command <command> [exit_code]")
			return
		}

		cmd := os.Args[2]
		exitCode := 0

		if len(os.Args) >= 4 {
			if code, err := strconv.Atoi(os.Args[3]); err == nil {
				exitCode = code
			}
		}

		err := shell.TrackCommand(cmd, exitCode)
		if err != nil {
			fmt.Printf("Track failed: %v\n", err)
		}

		// Also save to shell history in SQLite
		historyStore, hErr := getShellHistoryStore()
		if hErr == nil {
			mgr := shellhistory.NewManager(historyStore)
			mgr.TrackCommand(cmd, exitCode, "")
			historyStore.(storage.Storage).Close()
		}

	case "search":
		// Search sessions by text (SQLite backend)
		if len(os.Args) < 3 {
			fmt.Println("usage: rewind search <query>")
			return
		}

		query := os.Args[2]
		st, err := getStorage()
		if err != nil {
			// Fall back to JSON
			sessions, loadErr := storage.LoadAllSessions()
			if loadErr != nil {
				fmt.Println("search failed:", loadErr)
				return
			}
			fmt.Println("")
			fmt.Printf("Results for: %s (JSON mode)\n", query)
			fmt.Println("---------------------------")
			for _, s := range sessions {
				if strings.Contains(strings.ToLower(s.Command), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(s.Title), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(s.Summary), strings.ToLower(query)) {
					fmt.Printf("  %s - %s\n", s.ID, s.Command)
				}
			}
			fmt.Println("")
			return
		}
		defer st.Close()

		sessions, err := st.SearchSessions(query)
		if err != nil {
			fmt.Println("search failed:", err)
			return
		}

		fmt.Println("")
		fmt.Printf("Results for: %s\n", query)
		fmt.Println("---------------------------")
		for _, s := range sessions {
			fmt.Printf("  %s - %s\n", s.ID, s.Command)
		}
		fmt.Println("")

	case "migrate":
		// Migrate JSON sessions to SQLite
		fmt.Println("Migrating sessions from JSON to SQLite...")

		st, err := getStorage()
		if err != nil {
			fmt.Println("failed to open SQLite:", err)
			return
		}
		defer st.Close()

		sqliteStore, ok := st.(*storage.SQLiteStore)
		if !ok {
			fmt.Println("storage is not SQLite")
			return
		}

		count, err := storage.MigrateJSONToSQLite("sessions", sqliteStore)
		if err != nil {
			fmt.Println("migration failed:", err)
			return
		}

		fmt.Printf("Successfully migrated %d sessions to SQLite.\n", count)
		fmt.Println("")
		fmt.Println("Note: JSON files in sessions/ are preserved. To use SQLite by default,")
		fmt.Println("  set REWIND_DB_PATH=rewind.db or REWIND_USE_JSON=false")
		fmt.Println("")

	case "history":
		// View shell history
		limit := 20
		if len(os.Args) >= 3 {
			if l, err := strconv.Atoi(os.Args[2]); err == nil && l > 0 {
				limit = l
			}
		}

		historyStore, err := getShellHistoryStore()
		if err != nil {
			fmt.Println("shell history requires SQLite:", err)
			return
		}
		defer historyStore.(storage.Storage).Close()

		mgr := shellhistory.NewManager(historyStore)
		entries, err := mgr.ViewHistory(limit)
		if err != nil {
			fmt.Println("failed to get history:", err)
			return
		}

		shellhistory.PrintHistory(entries)

	case "import-history":
		// Import shell history from files
		historyStore, err := getShellHistoryStore()
		if err != nil {
			fmt.Println("shell history requires SQLite:", err)
			return
		}
		defer historyStore.(storage.Storage).Close()

		if len(os.Args) < 3 {
			// Auto-detect and import all
			fmt.Println("No shell specified, auto-detecting...")
			results, err := shellhistory.AutoDetectAndImport(historyStore)
			if err != nil {
				fmt.Println("import failed:", err)
				return
			}
			total := 0
			for _, count := range results {
				total += count
			}
			fmt.Printf("\nTotal imported: %d commands\n", total)
			return
		}

		shellType := os.Args[2]
		cfg := shellhistory.ImportConfig{
			ShellType: shellType,
			Store:     historyStore,
		}

		// Check if it's a path
		if _, err := os.Stat(shellType); err == nil {
			cfg.CustomPath = shellType
			cfg.ShellType = ""
		}

		count, err := shellhistory.ImportFromShell(cfg)
		if err != nil {
			fmt.Println("import failed:", err)
			shellhistory.PrintSupportedFormats()
			return
		}

		fmt.Printf("Successfully imported %d commands.\n", count)

	case "history-stats":
		// Show shell history statistics
		historyStore, err := getShellHistoryStore()
		if err != nil {
			fmt.Println("shell history requires SQLite:", err)
			return
		}
		defer historyStore.(storage.Storage).Close()

		mgr := shellhistory.NewManager(historyStore)
		stats, err := mgr.ShowStats()
		if err != nil {
			fmt.Println("failed to get stats:", err)
			return
		}

		shellhistory.PrintStats(stats)

	case "ide":
		// IDE integration subcommands
		if len(os.Args) < 3 {
			fmt.Println("usage:")
			fmt.Println("  rewind ide start [port]")
			fmt.Println("  rewind ide stop")
			fmt.Println("  rewind ide status")
			fmt.Println("  rewind ide permissions [ide] [on|off]")
			fmt.Println("  rewind ide activity [ide] [limit]")
			fmt.Println("  rewind ide projects")
			fmt.Println("  rewind ide analyze [project]")
			return
		}

		ideCommand := os.Args[2]

		switch ideCommand {
		case "start":
			port := ide.ServerPort
			if len(os.Args) >= 4 {
				if p, err := strconv.Atoi(os.Args[3]); err == nil {
					port = p
				}
			}

			st, err := getSQLiteStore()
			if err != nil {
				fmt.Println("IDE server requires SQLite:", err)
				return
			}
			defer st.Close()

			recorder := ide.NewSQLiteRecorder(st)
			server := ide.NewServer(recorder, port)

			if err := server.Start(); err != nil {
				fmt.Printf("Server error: %v\n", err)
			}

		case "stop":
			fmt.Println("IDE server stop — send SIGINT/Ctrl+C to the running server process.")

		case "status":
			st, err := getSQLiteStore()
			if err != nil {
				fmt.Println("IDE status requires SQLite:", err)
				return
			}
			defer st.Close()

			recorder := ide.NewSQLiteRecorder(st)
			status, err := recorder.GetStatus()
			if err != nil {
				fmt.Println("failed to get status:", err)
				return
			}

			fmt.Println("")
			fmt.Println("IDE RECORDING STATUS")
			fmt.Println("=====================")
			fmt.Printf("Server running:  %v\n", status.ServerRunning)
			if status.ServerRunning {
				fmt.Printf("Server port:     %d\n", status.ServerPort)
			}
			fmt.Printf("Activity count:  %d\n", status.ActivityCount)
			fmt.Printf("Connected IDEs:  %v\n", status.ConnectedIDEs)
			if status.ActiveProject != "" {
				fmt.Printf("Active project:  %s\n", status.ActiveProject)
			}
			ide.PrintPermissions(status.Permissions)

		case "permissions":
			if len(os.Args) < 5 {
				// Display current permissions
				st, err := getSQLiteStore()
				if err != nil {
					fmt.Println("permissions require SQLite:", err)
					return
				}
				defer st.Close()

				recorder := ide.NewSQLiteRecorder(st)
				status, err := recorder.GetStatus()
				if err != nil {
					fmt.Println("failed:", err)
					return
				}
				ide.PrintPermissions(status.Permissions)
				fmt.Println("Usage: rewind ide permissions <ide> <on|off>")
				fmt.Println("Example: rewind ide permissions vscode on")
				return
			}

			ideName := os.Args[3]
			action := os.Args[4]
			projectPath := "*"
			if len(os.Args) >= 6 {
				projectPath = os.Args[5]
			}

			st, err := getSQLiteStore()
			if err != nil {
				fmt.Println("permissions require SQLite:", err)
				return
			}
			defer st.Close()

			recorder := ide.NewSQLiteRecorder(st)

			switch action {
			case "on":
				if err := ide.EnableRecording(recorder, ideName, projectPath); err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
				fmt.Printf(" Recording enabled for %s (project: %s)\n", ide.IDEToHumanName(ideName), projectPath)
				fmt.Println("   file recording: ON")
				fmt.Println("   terminal recording: ON")
				fmt.Println("   AI recording: ON")
			case "off":
				if err := ide.DisableRecording(recorder, ideName, projectPath); err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
				fmt.Printf(" Recording disabled for %s (project: %s)\n", ide.IDEToHumanName(ideName), projectPath)
			default:
				fmt.Println("Invalid action. Use 'on' or 'off'.")
			}

		case "activity":
			ideName := ""
			if len(os.Args) >= 4 {
				ideName = os.Args[3]
			}
			limit := 30
			if len(os.Args) >= 5 {
				if l, err := strconv.Atoi(os.Args[4]); err == nil && l > 0 {
					limit = l
				}
			}

			st, err := getSQLiteStore()
			if err != nil {
				fmt.Println("activity requires SQLite:", err)
				return
			}
			defer st.Close()

			recorder := ide.NewSQLiteRecorder(st)
			activities, err := recorder.QueryActivity(types.IDEActivityFilter{
				IDEName: ideName,
				Limit:   limit,
			})
			if err != nil {
				fmt.Println("failed:", err)
				return
			}
			ide.PrintActivity(activities)

		case "projects":
			st, err := getSQLiteStore()
			if err != nil {
				fmt.Println("projects require SQLite:", err)
				return
			}
			defer st.Close()

			recorder := ide.NewSQLiteRecorder(st)
			projects, err := recorder.GetProjects()
			if err != nil {
				fmt.Println("failed:", err)
				return
			}
			ide.PrintProjects(projects)

		case "analyze":
			if len(os.Args) < 4 {
				fmt.Println("usage: rewind ide analyze <project-path>")
				return
			}
			projectPath := os.Args[3]

			st, err := getSQLiteStore()
			if err != nil {
				fmt.Println("analyze requires SQLite:", err)
				return
			}
			defer st.Close()

			recorder := ide.NewSQLiteRecorder(st)
			insight, err := ide.AnalyzeProject(recorder, projectPath)
			if err != nil {
				fmt.Printf("Analysis error: %v\n", err)
				return
			}
			ide.PrintInsight(insight)

		default:
			fmt.Printf("Unknown ide command: %s\n", ideCommand)
			fmt.Println("Valid: start, stop, status, permissions, activity, projects, analyze")
		}

	default:
		fmt.Println("unknown command")
	}
}
