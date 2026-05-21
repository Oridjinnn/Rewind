package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/habeldavidson007-glitch/rewind/internal/recall"
	"github.com/habeldavidson007-glitch/rewind/internal/markdown"
	"github.com/habeldavidson007-glitch/rewind/internal/chat"
	"github.com/habeldavidson007-glitch/rewind/internal/score"
	"github.com/habeldavidson007-glitch/rewind/internal/inspect"
	"github.com/habeldavidson007-glitch/rewind/internal/export"
	"github.com/habeldavidson007-glitch/rewind/internal/stats"
	"github.com/habeldavidson007-glitch/rewind/internal/timeline"
	"github.com/habeldavidson007-glitch/rewind/internal/detect"
	"github.com/habeldavidson007-glitch/rewind/internal/diff"
	"github.com/habeldavidson007-glitch/rewind/internal/recorder"
	"github.com/habeldavidson007-glitch/rewind/internal/replay"
	"github.com/habeldavidson007-glitch/rewind/internal/storage"
	"github.com/habeldavidson007-glitch/rewind/internal/shell"
	"github.com/habeldavidson007-glitch/rewind/internal/web"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("  rewind run <command>")
		fmt.Println("  rewind replay <session_id>")
		fmt.Println("  rewind web [port]         # Start the Rewind web UI")
		fmt.Println("  rewind setup              # Setup auto-recording for your shell")
		fmt.Println("  rewind chat <model>       # Chat with memory")
		fmt.Println("  rewind recall <query>     # Search sessions")
		fmt.Println("  rewind list               # List all sessions")
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

		session, err := recorder.RecordCommand(
			targetCommand,
			targetArgs,
		)

		if err != nil {
			fmt.Println("record failed:", err)
			return
		}

		sessionPath := filepath.Join(
			"sessions",
			session.ID+".json",
		)

		err = storage.SaveSession(
			session,
			sessionPath,
		)

		if err != nil {
			fmt.Println("save failed:", err)
			return
		}

		fmt.Println("session recorded:", session.ID)

	case "replay":

		if len(os.Args) < 3 {
			fmt.Println("missing session id")
			return
		}

		sessionID := os.Args[2]

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		leftPath := filepath.Join(
			"sessions",
			leftID+".json",
		)

		rightPath := filepath.Join(
			"sessions",
			rightID+".json",
		)

		leftSession, err := storage.LoadSession(
			leftPath,
		)

		if err != nil {
			fmt.Println("failed loading left session:", err)
			return
		}

		rightSession, err := storage.LoadSession(
			rightPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		sessionPath := filepath.Join(
			"sessions",
			sessionID+".json",
		)

		session, err := storage.LoadSession(
			sessionPath,
		)

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

		files, err := storage.ListSessions("sessions")

		if err != nil {
			fmt.Println("list failed:", err)
			return
		}

		fmt.Println("")
		fmt.Println("AVAILABLE SESSIONS")
		fmt.Println("------------------")

		for _, file := range files {

			name := file.Name()

			if strings.HasSuffix(name, ".json") {

				fmt.Println(
					strings.TrimSuffix(name, ".json"),
				)
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

	default:
		fmt.Println("unknown command")
	}
}
