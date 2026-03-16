// Package entity contains domain entities representing core business objects.
package entity

import (
	"time"
)

// Session represents a working session (Claude or terminal).
type Session struct {
	ID              string
	Name            string
	Description     string
	SessionType     string // "claude" or "terminal"
	Status          string
	Directory       string
	WorktreePath    string
	BranchName      string
	ClaudeConvID    string
	PID             int
	RepoID          string
	TaskID          string
	SkipPermissions bool
	Model           string
	ExtraCliArgs    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastActiveAt    time.Time
	ArchivedAt      *time.Time
}
