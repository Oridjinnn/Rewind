package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/habeldavidson007-glitch/rewind/internal/diff"
	"github.com/habeldavidson007-glitch/rewind/internal/recorder"
	"github.com/habeldavidson007-glitch/rewind/internal/replay"
	"github.com/habeldavidson007-glitch/rewind/internal/storage"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("  rewind run <command>")
		fmt.Println("  rewind replay <session_id>")
		return
	}

	command := os.Args[1]

	switch command {

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
	default:
		fmt.Println("unknown command")
	}
}
