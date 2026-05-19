package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	default:
		fmt.Println("unknown command")
	}
}
