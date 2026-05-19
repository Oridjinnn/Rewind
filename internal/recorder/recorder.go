package recorder

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/utils"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
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

			fmt.Println(line)

			session.Events = append(
				session.Events,
				types.Event{
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "stdout",
					Content:   line,
				},
			)
		}
	}()

	go func() {

		defer wg.Done()

		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {

			line := scanner.Text()

			fmt.Println(line)

			session.Events = append(
				session.Events,
				types.Event{
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "stderr",
					Content:   line,
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
