package meta

import (
	"strings"

	"github.com/Oridjinnn/Rewind/internal/summarize"
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
	// Generate real summary via internal summarizer (Ollama).
	// If summarization fails/returns empty, fall back to a minimal summary.
	var conversation []string
	for _, e := range session.Events {
		if e.Type == "user_message" || e.Type == "assistant_message" {
			conversation = append(conversation, e.Type+": "+e.Content)
		}
	}
	conversationText := strings.Join(conversation, "\n")
	if conversationText != "" {
		summary := summarize.GenerateSummary(conversationText)
		if summary != "" {
			session.Summary = summary
		} else {
			session.Summary = "Session recorded."
		}
	} else {
		session.Summary = "Session recorded."
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
