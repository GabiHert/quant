// Package entity contains domain entities representing core business objects.
package entity

import (
	"time"
)

// JobRun represents a single execution of a job.
type JobRun struct {
	ID           string
	JobID        string
	Status       string // pending, running, success, failed, cancelled, timed_out
	TriggeredBy  string // run ID that triggered this run (empty if manual/scheduled)
	SessionID    string // linked Claude session ID (for claude-type jobs)
	DurationMs   int64
	TokensUsed   int
	Result       string
	ErrorMessage string
	StartedAt    time.Time
	FinishedAt   *time.Time
}
