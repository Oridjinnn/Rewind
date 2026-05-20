package score

import (
	"fmt"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/ui"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func PrintSessionScore(
	session types.Session,
) {

	score := 100

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

	score -= stderrCount * 10

	if session.ExitCode != 0 {
		score -= 25
	}

	startTime, _ := time.Parse(
		time.RFC3339,
		session.StartedAt,
	)

	endTime, _ := time.Parse(
		time.RFC3339,
		session.EndedAt,
	)

	duration := endTime.Sub(startTime)

	if duration.Seconds() > 10 {
		score -= 10
	}

	if duration.Seconds() > 30 {
		score -= 15
	}

	if score < 0 {
		score = 0
	}

	grade := "A"

	switch {
	case score >= 90:
		grade = "A"
	case score >= 75:
		grade = "B"
	case score >= 60:
		grade = "C"
	case score >= 40:
		grade = "D"
	default:
		grade = "F"
	}

	status := ui.Success("HEALTHY")

	if score < 60 {
		status = ui.Error("UNSTABLE")
	}

	fmt.Println("")

	fmt.Println(
		ui.Header(
			"REWIND EXECUTION SCORE",
		),
	)

	fmt.Println("SCORE         ", ui.Highlight(
		fmt.Sprintf("%d / 100", score),
	))

	fmt.Println("GRADE         ", grade)

	fmt.Println("")

	fmt.Println("STATUS        ", status)

	if duration.Seconds() < 5 {
		fmt.Println("DURATION      ", ui.Success("FAST"))
	} else {
		fmt.Println("DURATION      ", ui.Error("SLOW"))
	}

	if stderrCount == 0 {
		fmt.Println("STDERR        ", ui.Success("CLEAN"))
	} else {
		fmt.Println("STDERR        ", ui.Error(
			fmt.Sprintf("%d events", stderrCount),
		))
	}

	fmt.Println("")

	fmt.Println("ANALYSIS")

	if stderrCount == 0 {
		fmt.Println("• no stderr events detected")
	}

	if session.ExitCode == 0 {
		fmt.Println("• execution completed successfully")
	}

	if duration.Seconds() < 5 {
		fmt.Println("• runtime duration is healthy")
	}

	if score < 60 {
		fmt.Println("• execution quality degraded")
	}

	fmt.Println("")
	fmt.Println("EVENTS        ", len(session.Events))
	fmt.Println("STDOUT        ", stdoutCount)
	fmt.Println("STDERR        ", stderrCount)

	fmt.Println("")
}
