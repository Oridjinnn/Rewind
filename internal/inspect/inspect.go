package inspect

import (
	"fmt"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/ui"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func PrintSessionInspect(
	session types.Session,
) {

	fmt.Println("")

	fmt.Println(
		ui.Header(
			"REWIND SESSION INSPECT",
		),
	)

	startTime, _ := time.Parse(
		time.RFC3339,
		session.StartedAt,
	)

	endTime, _ := time.Parse(
		time.RFC3339,
		session.EndedAt,
	)

	duration := endTime.Sub(startTime)

	stdoutCount := 0
	stderrCount := 0

	for _, event := range session.Events {

		if event.Type == "stdout" {
			stdoutCount++
		}

		if event.Type == "stderr" {
			stderrCount++
		}
	}

	status := ui.Success("success")

	if session.ExitCode != 0 {
		status = ui.Error("failure")
	}

	fmt.Println("COMMAND      ", ui.Highlight(session.Command))
	fmt.Println("STATUS       ", status)
	fmt.Println("EXIT CODE    ", session.ExitCode)
	fmt.Println("DURATION     ", duration)

	fmt.Println("")

	fmt.Println("EVENTS       ", len(session.Events))
	fmt.Println("STDOUT       ", stdoutCount)
	fmt.Println("STDERR       ", stderrCount)

	fmt.Println("")

	fmt.Println("STARTED      ", ui.Muted(session.StartedAt))
	fmt.Println("ENDED        ", ui.Muted(session.EndedAt))

	fmt.Println("")
}	
