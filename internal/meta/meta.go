package meta

import (
	"strings"

	"github.com/Oridjinnn/Rewind/pkg/types"
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

	case containsAny(firstMessage, "hello", "hi"):
		session.Title = "Greeting conversation"
		session.Tags = append(session.Tags, "greeting")
		session.Mood = "casual"

	case containsAny(firstMessage, "fix", "error", "bug"):
		session.Title = "Debugging session"
		session.Tags = append(session.Tags, "debugging")
		session.Mood = "focused"

	case containsAny(firstMessage, "design", "architecture"):
		session.Title = "System design discussion"
		session.Tags = append(session.Tags, "architecture")
		session.Mood = "creative"

	default:

		session.Title =
			"AI conversation"

		session.Mood = "neutral"
	}

	// SUMMARY

	// SUMMARY
	// Use a simple, deterministic summary based on the first user intent.
	// (Previously this was a placeholder and broke downstream expectations.)
	if session.Title != "" {
		session.Summary = "Summary: " + session.Title + "."
	} else {
		session.Summary = "Summary: AI conversation." 
	}

}

func containsAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
