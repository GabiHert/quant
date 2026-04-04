// Package entity contains domain entities representing core business objects.
package entity

import (
	"time"
)

// JobGroup represents a visual grouping of jobs on the canvas.
type JobGroup struct {
	ID          string
	Name        string
	WorkspaceID string
	JobIDs      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
