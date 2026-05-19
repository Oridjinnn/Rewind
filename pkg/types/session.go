package types

type Event struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

type Session struct {
	ID        string  `json:"id"`
	Command   string  `json:"command"`
	StartedAt string  `json:"started_at"`
	EndedAt   string  `json:"ended_at"`
	ExitCode  int     `json:"exit_code"`
	Events    []Event `json:"events"`
}
