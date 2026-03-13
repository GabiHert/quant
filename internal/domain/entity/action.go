// Package entity contains domain entities representing core business objects.
package entity

import (
	"time"
)

// Action represents a discrete action taken during a session.
type Action struct {
	ID        string
	SessionID string
	Type      string
	Content   string
	Timestamp time.Time
}
