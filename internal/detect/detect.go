package detect

import (
	"fmt"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func DetectLoops(session types.Session) {

	fmt.Println("")
	fmt.Println("LOOP DETECTION")
	fmt.Println("----------------------")

	counts := map[string]int{}

	for _, event := range session.Events {

		counts[event.Content]++
	}

	found := false

	for content, count := range counts {

		if count >= 3 {

			found = true

			fmt.Println("")
			fmt.Println("LOOP DETECTED")
			fmt.Println("")
			fmt.Println("Pattern:")
			fmt.Println(content)
			fmt.Println("")
			fmt.Println("Repeated:")
			fmt.Printf("%d times\n", count)
		}
	}

	if !found {

		fmt.Println("")
		fmt.Println("No loops detected.")
	}
}
