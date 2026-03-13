// Package adapter contains interfaces that application services implement.
package adapter

import (
	"quant/internal/domain/entity"
)

// ActionLogger defines the service interface for action logging operations.
// This is the application adapter that the actionLoggerService implements.
type ActionLogger interface {
	LogAction(sessionID string, actionType string, content string) (*entity.Action, error)
	GetActions(sessionID string) ([]entity.Action, error)
}
