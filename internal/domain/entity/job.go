// Package entity contains domain entities representing core business objects.
package entity

import (
	"time"
)

// Job represents an automated task that can be scheduled or triggered manually.
type Job struct {
	ID               string
	Name             string
	Description      string
	Type             string // "claude" or "bash"
	WorkingDirectory string

	// Schedule configuration
	ScheduleEnabled   bool
	ScheduleType      string // "recurring" or "one_time"
	CronExpression    string
	ScheduleInterval  int // minutes
	ScheduleStartTime *time.Time
	TimeoutSeconds    int

	// Claude session configuration
	Prompt              string
	AllowBypass         bool
	AutonomousMode      bool
	MaxRetries          int
	Model               string
	OverrideRepoCommand string
	ClaudeCommand       string

	// Bash script configuration
	Interpreter   string
	ScriptContent string
	EnvVariables  map[string]string

	CreatedAt time.Time
	UpdatedAt time.Time
	LastRunAt *time.Time
}
