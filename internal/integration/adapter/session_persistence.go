// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// SessionPersistence combines all session-related persistence usecase interfaces.
// Integration persistence implementations must implement this interface.
type SessionPersistence interface {
	usecase.FindSession
	usecase.SaveSession
	usecase.DeleteSession
	usecase.UpdateSession
}
