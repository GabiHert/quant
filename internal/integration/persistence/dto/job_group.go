// Package dto contains persistence data transfer objects for SQLite row mapping.
package dto

import (
	"time"

	"quant/internal/domain/entity"
)

// JobGroupRow represents a job_groups row in the SQLite database.
type JobGroupRow struct {
	ID          string
	Name        string
	WorkspaceID string
	CreatedAt   string
	UpdatedAt   string
}

// ToEntity converts a JobGroupRow to a domain entity (without JobIDs, which are loaded separately).
func (r JobGroupRow) ToEntity() entity.JobGroup {
	createdAt, _ := time.Parse(time.RFC3339, r.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, r.UpdatedAt)

	return entity.JobGroup{
		ID:          r.ID,
		Name:        r.Name,
		WorkspaceID: r.WorkspaceID,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// JobGroupRowFromEntity converts a domain entity to a JobGroupRow.
func JobGroupRowFromEntity(group entity.JobGroup) JobGroupRow {
	return JobGroupRow{
		ID:          group.ID,
		Name:        group.Name,
		WorkspaceID: group.WorkspaceID,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
	}
}
