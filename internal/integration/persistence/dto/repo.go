// Package dto contains persistence data transfer objects for SQLite row mapping.
package dto

import (
	"time"

	"quant/internal/domain/entity"
)

// RepoRow represents a repo row in the SQLite database.
type RepoRow struct {
	ID        string
	Name      string
	Path      string
	CreatedAt string
	UpdatedAt string
}

// ToEntity converts a RepoRow to a domain entity.
func (r RepoRow) ToEntity() entity.Repo {
	createdAt, _ := time.Parse(time.RFC3339, r.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, r.UpdatedAt)

	return entity.Repo{
		ID:        r.ID,
		Name:      r.Name,
		Path:      r.Path,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// RepoRowFromEntity converts a domain entity to a RepoRow.
func RepoRowFromEntity(repo entity.Repo) RepoRow {
	return RepoRow{
		ID:        repo.ID,
		Name:      repo.Name,
		Path:      repo.Path,
		CreatedAt: repo.CreatedAt.Format(time.RFC3339),
		UpdatedAt: repo.UpdatedAt.Format(time.RFC3339),
	}
}
