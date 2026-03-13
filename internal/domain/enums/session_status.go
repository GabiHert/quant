// Package enums contains domain enumeration types.
package enums

import (
	"quant/internal/domain/enums/sessionstatus"
)

// SessionStatus represents the possible states of a session.
type SessionStatus struct {
	value string
}

// Value returns the string representation of the session status.
func (s SessionStatus) Value() string {
	return s.value
}

// String returns the string representation of the session status.
func (s SessionStatus) String() string {
	return s.value
}

// IsValid returns true if the session status is a recognized value.
func (s SessionStatus) IsValid() bool {
	switch s.value {
	case sessionstatus.Idle,
		sessionstatus.Running,
		sessionstatus.Paused,
		sessionstatus.Done,
		sessionstatus.Error:
		return true
	default:
		return false
	}
}

// NewSessionStatus creates a new SessionStatus from a string value.
func NewSessionStatus(value string) SessionStatus {
	return SessionStatus{value: value}
}

// SessionStatusIdle returns the idle session status.
func SessionStatusIdle() SessionStatus {
	return SessionStatus{value: sessionstatus.Idle}
}

// SessionStatusRunning returns the running session status.
func SessionStatusRunning() SessionStatus {
	return SessionStatus{value: sessionstatus.Running}
}

// SessionStatusPaused returns the paused session status.
func SessionStatusPaused() SessionStatus {
	return SessionStatus{value: sessionstatus.Paused}
}

// SessionStatusDone returns the done session status.
func SessionStatusDone() SessionStatus {
	return SessionStatus{value: sessionstatus.Done}
}

// SessionStatusError returns the error session status.
func SessionStatusError() SessionStatus {
	return SessionStatus{value: sessionstatus.Error}
}
