package types

// IDEActivity represents a single IDE event recorded by any supported IDE.
type IDEActivity struct {
	ID              int64            `json:"id,omitempty"`
	IDEName         string           `json:"ide_name"`         // 'vscode','cursor','intellij-idea','goland','pycharm','webstorm','android-studio','eclipse','sublime','nvim','vim'
	ProjectName     string           `json:"project_name"`     // project name
	ProjectPath     string           `json:"project_path"`     // full path
	ActivityType    string           `json:"activity_type"`    // 'file_open','file_save','file_edit','file_close','terminal_cmd','build_start','build_end','build_error','test_run','test_pass','test_fail','git_commit','git_push','git_pull','git_branch','debug_start','debug_breakpoint','debug_step','debug_end','ai_chat','ai_completion','ai_accept','ai_reject','refactor','search','run_config','dependency_change'
	FilePath        string           `json:"file_path"`        // affected file path
	Language        string           `json:"language"`         // go, typescript, python, etc.
	Metadata        map[string]any   `json:"metadata"`         // flexible metadata
	ContentSnapshot string           `json:"content_snapshot"` // optional content diff/snapshot
	ExecutedAt      string           `json:"executed_at"`      // timestamp
	SessionID       string           `json:"session_id"`       // linked terminal session ID
}

// IDEEvent is the JSON payload sent from IDE extensions to rewind server.
type IDEEvent struct {
	Protocol  string      `json:"protocol"`  // 'rewind-ide-v1'
	IDE       string      `json:"ide"`       // ide name
	Version   string      `json:"version"`   // ide version
	Project   string      `json:"project"`   // project name
	ProjectPath string    `json:"project_path"`
	Event     IDEEventData `json:"event"`
}

// IDEEventData is the inner event payload from an IDE.
type IDEEventData struct {
	Type            string         `json:"type"`            // activity type
	Timestamp       string         `json:"timestamp"`       // ISO 8601
	File            string         `json:"file,omitempty"`  // affected file
	Language        string         `json:"language,omitempty"`
	LinesAdded      int            `json:"lines_added,omitempty"`
	LinesRemoved    int            `json:"lines_removed,omitempty"`
	CursorLine      int            `json:"cursor_line,omitempty"`
	CursorColumn    int            `json:"cursor_column,omitempty"`
	DurationMs      int            `json:"duration_ms,omitempty"`
	ExitCode        int            `json:"exit_code,omitempty"`
	Message         string         `json:"message,omitempty"`
	Tags            []string       `json:"tags,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	ContentSnapshot string         `json:"content_snapshot,omitempty"`
	SessionID       string         `json:"session_id,omitempty"`
}

// IDEPermission controls recording permissions per IDE and project.
type IDEPermission struct {
	IDEName           string `json:"ide_name"`
	ProjectPath       string `json:"project_path"`
	RecordingEnabled  bool   `json:"recording_enabled"`  // master switch
	FileRecording     bool   `json:"file_recording"`     // record file contents
	TerminalRecording bool   `json:"terminal_recording"` // record terminal commands
	AiRecording       bool   `json:"ai_recording"`       // record AI conversations
	LastToggled       string `json:"last_toggled"`
}

// IDEStatus represents the current state of IDE recording.
type IDEStatus struct {
	ServerRunning   bool              `json:"server_running"`
	ServerPort      int               `json:"server_port"`
	ConnectedIDEs   []string          `json:"connected_ides"`
	Permissions     []IDEPermission   `json:"permissions"`
	ActivityCount   int64             `json:"activity_count"`
	ActiveProject   string            `json:"active_project,omitempty"`
}

// IDEActivityFilter is used for querying IDE activities.
type IDEActivityFilter struct {
	IDEName      string `json:"ide_name,omitempty"`
	ProjectPath  string `json:"project_path,omitempty"`
	ActivityType string `json:"activity_type,omitempty"`
	FilePath     string `json:"file_path,omitempty"`
	Language     string `json:"language,omitempty"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// IDEProject represents a tracked IDE project.
type IDEProject struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IDEName      string `json:"ide_name"`
	LastActivity string `json:"last_activity"`
	EventCount   int64  `json:"event_count"`
	IsRecording  bool   `json:"is_recording"`
}

// ProductivityInsight represents AI-generated insights from IDE activity.
type ProductivityInsight struct {
	Project          string           `json:"project"`
	TotalEvents      int64            `json:"total_events"`
	ActiveHours      float64          `json:"active_hours"`
	TopLanguages     []string         `json:"top_languages"`
	TopFiles         []string         `json:"top_files"`
	BuildSuccessRate float64          `json:"build_success_rate"`
	TestPassRate     float64          `json:"test_pass_rate"`
	AIUsageCount     int64            `json:"ai_usage_count"`
	AIRejectRate     float64          `json:"ai_reject_rate"`
	MostUsedCommands []string         `json:"most_used_commands"`
	Suggestions      []string         `json:"suggestions"`
}