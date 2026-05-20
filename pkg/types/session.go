package types

import "time"

type Event struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type Session struct {
	ID        string  `json:"id"`
	Command   string  `json:"command"`
	Model     string  `json:"model,omitempty"`

	Title     string   `json:"title,omitempty"`
	Summary   string   `json:"summary,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Mood      string   `json:"mood,omitempty"`

	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`

	ExitCode int `json:"exit_code"`

	Events []Event `json:"events"`
}
