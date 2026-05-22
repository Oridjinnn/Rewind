package detect

import (
	"fmt"

	"github.com/Oridjinnn/Rewind/internal/ui"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

func DetectLoops(session types.Session) {

	fmt.Println("")

	fmt.Println(
		ui.Header(
			"REWIND LOOP DETECTION",
		),
	)

	counts := map[string]int{}

	for _, event := range session.Events {

		if event.Content == "" {
			continue
		}

		counts[event.Content]++
	}

	found := false

	for content, count := range counts {

		if count >= 3 {

			found = true

			fmt.Println("")

			fmt.Println(
				ui.Error(
					"LOOP DETECTED",
				),
			)

			fmt.Println("")

			fmt.Println(
				ui.Highlight(
					"Pattern:",
				),
			)

			fmt.Println(content)

			fmt.Println("")

			fmt.Println(
				ui.Highlight(
					"Repeated:",
				),
			)

			fmt.Printf(
				"%s times\n",
				ui.Error(
					fmt.Sprintf("%d", count),
				),
			)
		}
	}

	if !found {

		fmt.Println("")

		fmt.Println(
			ui.Success(
				"No loops detected.",
			),
		)
	}

	fmt.Println("")
}
