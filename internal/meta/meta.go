package meta

import (
	"strings"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

func GenerateMetadata(
	session *types.Session,
) {

	if len(session.Events) == 0 {
		return
	}

	firstMessage := ""

	for _, event := range session.Events {

		if event.Type == "user_message" {

			firstMessage = strings.ToLower(
				event.Content,
			)

			break
		}
	}

	if firstMessage == "" {
		return
	}

	// TITLE

	switch {

	case strings.Contains(firstMessage, "hello"),
		strings.Contains(firstMessage, "hi"):

		session.Title =
			"Greeting conversation"

		session.Tags = append(
			session.Tags,
			"greeting",
		)

		session.Mood = "casual"

	case strings.Contains(firstMessage, "fix"),
		strings.Contains(firstMessage, "error"),
		strings.Contains(firstMessage, "bug"):

		session.Title =
			"Debugging session"

		session.Tags = append(
			session.Tags,
			"debugging",
		)

		session.Mood = "focused"

	case strings.Contains(firstMessage, "design"),
		strings.Contains(firstMessage, "architecture"):

		session.Title =
			"System design discussion"

		session.Tags = append(
			session.Tags,
			"architecture",
		)

		session.Mood = "creative"

	default:

		session.Title =
			"AI conversation"

		session.Mood = "neutral"
	}

	// SUMMARY

	session.Summary =
		"Automatically generated conversation summary."
}
