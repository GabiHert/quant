// Package sessionstatus contains string constants for session status values.
package sessionstatus

const (
	// Idle indicates the session is created but not running.
	Idle = "idle"

	// Running indicates the session has an active Claude process.
	Running = "running"

	// Paused indicates the session was running but has been temporarily stopped.
	Paused = "paused"

	// Done indicates the session completed successfully.
	Done = "done"

	// Error indicates the session encountered an error.
	Error = "error"
)
