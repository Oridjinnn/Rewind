package meta

import (
	"testing"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

func TestGenerateMetadata_NoEvents_NoMutation(t *testing.T) {
	s := types.Session{Title: "KeepMe", Summary: "KeepSummary", Mood: "keep"}

	GenerateMetadata(&s)

	if s.Title != "KeepMe" || s.Summary != "KeepSummary" || s.Mood != "keep" {
		t.Fatalf("session mutated unexpectedly: %+v", s)
	}
}

func TestGenerateMetadata_KeywordBug_TitleDebuggingSession(t *testing.T) {
	s := types.Session{Tags: []string{}}
	s.Events = []types.Event{{
		Type:    "user_message",
		Content: "There is a bug in the code",
	}}

	GenerateMetadata(&s)

	if s.Title != "Debugging session" {
		t.Fatalf("Title = %q, want %q", s.Title, "Debugging session")
	}
}

func TestGenerateMetadata_Unknown_TitleAIConversation(t *testing.T) {
	s := types.Session{Tags: []string{}}
	s.Events = []types.Event{{
		Type:    "user_message",
		Content: "Just talking about random stuff",
	}}

	GenerateMetadata(&s)

	if s.Title != "AI conversation" {
		t.Fatalf("Title = %q, want %q", s.Title, "AI conversation")
	}
}

