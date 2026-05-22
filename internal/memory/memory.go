package memory

import (
	"github.com/Oridjinnn/Rewind/pkg/types"
)

// DEPRECATED: Use v2.go for semantic memory and hybrid ranking.
// This file is kept temporarily for backward compatibility but will be removed in v0.3.0.
func BuildMemoryContext(
	sessions []types.Session,
	query string,
) string {
	return "" 
}
