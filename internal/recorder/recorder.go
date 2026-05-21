package recorder

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/Oridjinnn/Rewind/internal/redact"
	"github.com/Oridjinnn/Rewind/internal/utils"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

func RecordCommand(command string, args []string) (types.Session, error) {

	session := types.Session{
		ID:        utils.GenerateSessionID(),
		Command:   command,
		StartedAt: time.Now().Format(time.RFC3339),
		Events:    []types.Event{},
	}

	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return session, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return session, err
	}

	err = cmd.Start()
	if err != nil {
		return session, err
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {

		defer wg.Done()

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {

			line := scanner.Text()

			// Apply redaction before storing or printing
			redacted := redact.RedactCommand(line)
			if redacted == "" && line != "" {
				// If redact mode is 'skip', we don't even print it to the console
				continue
			}

			fmt.Println(redacted)

			session.Events = append(
				session.Events,
				types.Event{
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "stdout",
					Content:   redacted,
				},
			)
		}
	}()

	go func() {

		defer wg.Done()

		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {

			line := scanner.Text()

			redacted := redact.RedactCommand(line)
			if redacted == "" && line != "" {
				continue
			}

			fmt.Println(redacted)

			session.Events = append(
				session.Events,
				types.Event{
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "stderr",
					Content:   redacted,
				},
			)
		}
	}()

	wg.Wait()

	err = cmd.Wait()

	session.EndedAt = time.Now().Format(time.RFC3339)

	if err != nil {
		session.ExitCode = 1
	} else {
		session.ExitCode = 0
	}

	return session, nil
}
