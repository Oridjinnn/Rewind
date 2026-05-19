package timeline

import (
	"fmt"
	"time"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func RenderTimeline(session types.Session) {

	fmt.Println("")
	fmt.Println("TIMELINE")
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

	fmt.Printf("00.000s  START\n")

	for _, event := range session.Events {

		eventTime, err := time.Parse(
			time.RFC3339,
			event.Timestamp,
		)

		if err != nil {
			continue
		}

		elapsed := eventTime.Sub(startTime)

		fmt.Printf(
			"%07.3fs  %-7s %s\n",
			elapsed.Seconds(),
			event.Type,
			event.Content,
		)
	}

	if session.EndedAt != "" {

		endTime, err := time.Parse(
			time.RFC3339,
			session.EndedAt,
		)

		if err == nil {

			total := endTime.Sub(startTime)

			fmt.Printf(
				"%07.3fs  END\n",
				total.Seconds(),
			)
		}
	}

	fmt.Println("")
}
