package stats

import (
	"fmt"
	"time"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func RenderStats(session types.Session) {

	fmt.Println("")
	fmt.Println("SESSION STATS")
	fmt.Println("------------------------------------------------")
	fmt.Println("")

	startTime, err := time.Parse(
		time.RFC3339,
		session.StartedAt,
	)

	if err != nil {
		fmt.Println("invalid session start time")
		return
	}

	endTime, err := time.Parse(
		time.RFC3339,
		session.EndedAt,
	)

	if err != nil {
		fmt.Println("invalid session end time")
		return
	}

	duration := endTime.Sub(startTime)

	totalEvents := len(session.Events)

	stdoutCount := 0
	stderrCount := 0

	for _, event := range session.Events {

		switch event.Type {

		case "stdout":
			stdoutCount++

		case "stderr":
			stderrCount++
		}
	}

	density := 0.0

	if duration.Seconds() > 0 {

		density = float64(totalEvents) / duration.Seconds()
	}

	status := "SUCCESS"

	if session.ExitCode != 0 {
		status = "FAILED"
	}

	fmt.Println("Session ID:", session.ID)
	fmt.Println("Command:", session.Command)
	fmt.Println("")

	fmt.Printf("Duration: %.3fs\n", duration.Seconds())
	fmt.Println("Total Events:", totalEvents)
	fmt.Println("")

	fmt.Println("Stdout Events:", stdoutCount)
	fmt.Println("Stderr Events:", stderrCount)
	fmt.Println("")

	fmt.Printf(
		"Output Density:\n%.2f events/sec\n",
		density,
	)

	fmt.Println("")
	fmt.Println("Status:")
	fmt.Println(status)

	fmt.Println("")
}
