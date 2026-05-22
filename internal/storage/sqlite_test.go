package storage

import (
	"os"
	"testing"
	"time"

	"github.com/Oridjinnn/Rewind/pkg/types"
)

func TestSQLiteStore_SaveLoadRoundTrip(t *testing.T) {
	db, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	s := types.Session{
		ID:         "s1",
		Command:    "go test ./...",
		Model:      "test-model",
		Title:      "T",
		Summary:    "S",
		Tags:       []string{"tag"},
		Mood:       "neutral",
		StartedAt:  now,
		EndedAt:    now,
		ExitCode:   0,
		Events:     []types.Event{{Timestamp: now, Type: "user_message", Content: "hi"}},
	}

	if err := db.SaveSession(s); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	loaded, err := db.LoadSession("s1")
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}

	if loaded.ID != s.ID || loaded.Command != s.Command || loaded.Title != s.Title || loaded.Summary != s.Summary {
		t.Fatalf("loaded session mismatch: want %+v, got %+v", s, loaded)
	}
	if len(loaded.Events) != 1 || loaded.Events[0].Type != "user_message" {
		t.Fatalf("loaded events mismatch: %+v", loaded.Events)
	}
}

func TestSQLiteStore_LoadAllSessions_ReturnsAll(t *testing.T) {
	db, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for i := 0; i < 3; i++ {
		s := types.Session{
			ID:        "s" + string(rune('A'+i)),
			Command:   "cmd" + string(rune('A'+i)),
			Title:     "T" + string(rune('A'+i)),
			Summary:   "S" + string(rune('A'+i)),
			Tags:      []string{"t"},
			Mood:      "neutral",
			StartedAt: now,
			EndedAt:   now,
			ExitCode:  0,
			Events:    []types.Event{{Timestamp: now, Type: "user_message", Content: "x"}},
		}
		if err := db.SaveSession(s); err != nil {
			t.Fatalf("SaveSession: %v", err)
		}
	}

	all, err := db.LoadAllSessions()
	if err != nil {
		t.Fatalf("LoadAllSessions: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("LoadAllSessions count = %d, want 3", len(all))
	}
}

func TestSQLiteStore_DeleteSession_RemovesEvents(t *testing.T) {
	db, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	s := types.Session{
		ID:        "s1",
		Command:   "cmd",
		Title:     "T",
		Summary:   "S",
		Tags:      []string{"tag"},
		Mood:      "neutral",
		StartedAt: now,
		EndedAt:   now,
		ExitCode:  0,
		Events: []types.Event{{Timestamp: now, Type: "user_message", Content: "hi"}},
	}
	if err := db.SaveSession(s); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	if err := db.DeleteSession("s1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	loaded, err := db.LoadSession("s1")
	if err == nil {
		t.Fatalf("expected LoadSession error, got session: %+v", loaded)
	}
}

func TestSQLiteStore_SearchSessions_RelevantResults(t *testing.T) {
	db, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)

	_ = os.Setenv("TZ", "UTC")

	s1 := types.Session{ID: "s1", Command: "go build", Title: "T1", Summary: "debug stuff", Tags: []string{}, Mood: "neutral", StartedAt: now, EndedAt: now, ExitCode: 0, Events: []types.Event{}}
	s2 := types.Session{ID: "s2", Command: "npm test", Title: "T2", Summary: "unrelated", Tags: []string{}, Mood: "neutral", StartedAt: now, EndedAt: now, ExitCode: 0, Events: []types.Event{}}

	if err := db.SaveSession(s1); err != nil {
		t.Fatalf("SaveSession s1: %v", err)
	}
	if err := db.SaveSession(s2); err != nil {
		t.Fatalf("SaveSession s2: %v", err)
	}

	res, err := db.SearchSessions("debug")
	if err != nil {
		t.Fatalf("SearchSessions: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("expected results")
	}
	for _, r := range res {
		if r.ID != "s1" {
			// allow ordering, but only s1 should match
			t.Fatalf("unexpected session in results: %+v", r)
		}
	}
}

