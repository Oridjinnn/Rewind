package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("rewind - time-travel debugging for AI workflows")
		fmt.Println("")
		fmt.Println("usage:")
		fmt.Println("  rewind record")
		fmt.Println("  rewind replay")
		fmt.Println("  rewind diff")
		return
	}

	command := os.Args[1]

	switch command {
	case "record":
		fmt.Println("starting recorder...")
	case "replay":
		fmt.Println("starting replay...")
	case "diff":
		fmt.Println("starting diff...")
	default:
		fmt.Println("unknown command:", command)
	}
}
