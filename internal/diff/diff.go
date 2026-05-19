package diff

import (
	"fmt"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func CompareSessions(
	left types.Session,
	right types.Session,
) {

	fmt.Println("")
	fmt.Println("SESSION DIFF")
	fmt.Println("----------------------")
	fmt.Println("LEFT :", left.ID)
	fmt.Println("RIGHT:", right.ID)
	fmt.Println("")

	maxEvents := len(left.Events)

	if len(right.Events) > maxEvents {
		maxEvents = len(right.Events)
	}

	for i := 0; i < maxEvents; i++ {

		if i >= len(left.Events) {

			fmt.Println("+", right.Events[i].Content)
			continue
		}

		if i >= len(right.Events) {

			fmt.Println("-", left.Events[i].Content)
			continue
		}

		leftEvent := left.Events[i]
		rightEvent := right.Events[i]

		if leftEvent.Content == rightEvent.Content {

			fmt.Println("=", leftEvent.Content)

		} else {

			fmt.Println("-", leftEvent.Content)
			fmt.Println("+", rightEvent.Content)
		}
	}
}
