package recorder

import (
	"bufio"
	"os/exec"
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

	stdoutScanner := bufio.NewScanner(stdout)

	for stdoutScanner.Scan() {
		session.Events = append(
			session.Events,
			types.Event{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "stdout",
				Content:   stdoutScanner.Text(),
			},
		)
	}

	stderrScanner := bufio.NewScanner(stderr)

	for stderrScanner.Scan() {
		session.Events = append(
			session.Events,
			types.Event{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "stderr",
				Content:   stderrScanner.Text(),
			},
		)
	}

	err = cmd.Wait()

	session.EndedAt = time.Now().Format(time.RFC3339)

	if err != nil {
		session.ExitCode = 1
	} else {
		session.ExitCode = 0
	}

	return session, nil
}
