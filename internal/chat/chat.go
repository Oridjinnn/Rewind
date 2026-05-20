package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func LoadSession(
	path string,
) (types.Session, error) {

	var session types.Session

	data, err := os.ReadFile(path)

	if err != nil {
		return session, err
	}

	err = json.Unmarshal(
		data,
		&session,
	)

	if err != nil {
		return session, err
	}

	return session, nil
}

func LoadAllSessions() ([]types.Session, error) {

	var sessions []types.Session

	files, err := os.ReadDir(
		"sessions",
	)

	if err != nil {
		return sessions, err
	}

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(
			file.Name(),
			".json",
		) {
			continue
		}

		path := filepath.Join(
			"sessions",
			file.Name(),
		)

		session, err := LoadSession(
			path,
		)

		if err != nil {
			fmt.Println(
				"failed loading:",
				file.Name(),
			)
			continue
		}

		sessions = append(
			sessions,
			session,
		)
	}

	return sessions, nil
}

func StartChat(model string) {
	fmt.Println("Starting chat with model:", model)
	fmt.Println("Type 'exit' to quit")
	fmt.Println("")

	reader := bufio.NewReader(os.Stdin)

	var messages []types.Message

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "" {
			continue
		}

		userMsg := types.Message{
			Role:    "user",
			Content: input,
			Time:    time.Now(),
		}

		messages = append(messages, userMsg)

		fmt.Println("[AI response would appear here - Ollama integration pending]")
		fmt.Println("")
	}
}
