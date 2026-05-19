package replay

import (
	"fmt"
	"time"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func ReplaySession(session types.Session) {

	fmt.Println("")
	fmt.Println("REPLAY SESSION")
	fmt.Println("----------------------")
	fmt.Println("ID:", session.ID)
	fmt.Println("COMMAND:", session.Command)
	fmt.Println("EXIT CODE:", session.ExitCode)
	fmt.Println("")

	for _, event := range session.Events {

		fmt.Printf(
			"[%s] %s\n",
			event.Timestamp,
			event.Type,
		)

		fmt.Println(event.Content)
		fmt.Println("")

		time.Sleep(700 * time.Millisecond)
	}
}
