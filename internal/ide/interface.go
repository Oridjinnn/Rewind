package ide

import (
	"github.com/Oridjinnn/Rewind/pkg/types"
)

// Recorder stores IDE activity events and manages permissions.
type Recorder interface {
	// RecordEvent saves a single IDE event.
	RecordEvent(event types.IDEEvent) error

	// RecordBatch saves multiple IDE events in a transaction.
	RecordBatch(events []types.IDEEvent) error

	// CheckPermission verifies if recording is enabled for this IDE/project.
	CheckPermission(ideName, projectPath string) (bool, error)

	// GetPermission retrieves full permission settings.
	GetPermission(ideName, projectPath string) (types.IDEPermission, error)

	// SetPermission updates permission settings.
	SetPermission(perm types.IDEPermission) error

	// GetStatus returns current IDE recording status.
	GetStatus() (types.IDEStatus, error)

	// GetProjects returns all tracked projects.
	GetProjects() ([]types.IDEProject, error)

	// QueryActivity retrieves filtered IDE activities.
	QueryActivity(filter types.IDEActivityFilter) ([]types.IDEActivity, error)

	// Close closes the underlying storage.
	Close() error
}