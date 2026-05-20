package diff

import (
	"fmt"

	"github.com/habeldavidson007-glitch/rewind/internal/ui"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func CompareSessions(
	left types.Session,
	right types.Session,
) {

	fmt.Println("")

	fmt.Println(
		ui.Header(
			"REWIND SESSION DIFF",
		),
	)

	fmt.Println("LEFT :", left.ID)
	fmt.Println("RIGHT:", right.ID)
	fmt.Println("")

	maxEvents := len(left.Events)

	if len(right.Events) > maxEvents {
		maxEvents = len(right.Events)
	}

	for i := 0; i < maxEvents; i++ {

		if i >= len(left.Events) {

			fmt.Println(
				ui.Success(
					"[+] " + right.Events[i].Content,
				),
			)

			continue
		}

		if i >= len(right.Events) {

			fmt.Println(
				ui.Error(
					"[-] " + left.Events[i].Content,
				),
			)

			continue
		}

		leftContent := left.Events[i].Content
		rightContent := right.Events[i].Content

		if leftContent != rightContent {

			fmt.Println(
				ui.Error(
					"[-] " + leftContent,
				),
			)

			fmt.Println(
				ui.Success(
					"[+] " + rightContent,
				),
			)
		}
	}
}
