// Package dto contains persistence data transfer objects for SQLite row mapping.
package dto

import (
	"database/sql"
	"time"

	"quant/internal/domain/entity"
)

// SessionRow represents a session row in the SQLite database.
type SessionRow struct {
	ID              string
	Name            string
	Description     sql.NullString
	Status          string
	Directory       string
	WorktreePath    sql.NullString
	BranchName      sql.NullString
	ClaudeConvID    sql.NullString
	PID             int
	RepoID          sql.NullString
	TaskID          sql.NullString
	SkipPermissions bool
	CreatedAt       string
	UpdatedAt       string
	LastActiveAt    string
}

// ToEntity converts a SessionRow to a domain entity.
func (r SessionRow) ToEntity() entity.Session {
	createdAt, _ := time.Parse(time.RFC3339, r.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, r.UpdatedAt)
	lastActiveAt, _ := time.Parse(time.RFC3339, r.LastActiveAt)

	return entity.Session{
		ID:              r.ID,
		Name:            r.Name,
		Description:     r.Description.String,
		Status:          r.Status,
		Directory:       r.Directory,
		WorktreePath:    r.WorktreePath.String,
		BranchName:      r.BranchName.String,
		ClaudeConvID:    r.ClaudeConvID.String,
		PID:             r.PID,
		RepoID:          r.RepoID.String,
		TaskID:          r.TaskID.String,
		SkipPermissions: r.SkipPermissions,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		LastActiveAt:    lastActiveAt,
	}
}

// SessionRowFromEntity converts a domain entity to a SessionRow.
func SessionRowFromEntity(session entity.Session) SessionRow {
	return SessionRow{
		ID:              session.ID,
		Name:            session.Name,
		Description:     toNullString(session.Description),
		Status:          session.Status,
		Directory:       session.Directory,
		WorktreePath:    toNullString(session.WorktreePath),
		BranchName:      toNullString(session.BranchName),
		ClaudeConvID:    toNullString(session.ClaudeConvID),
		PID:             session.PID,
		RepoID:          toNullString(session.RepoID),
		TaskID:          toNullString(session.TaskID),
		SkipPermissions: session.SkipPermissions,
		CreatedAt:       session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
		LastActiveAt:    session.LastActiveAt.Format(time.RFC3339),
	}
}

// toNullString converts a string to sql.NullString.
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
