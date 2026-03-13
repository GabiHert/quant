// Package service contains application service implementations with business logic.
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"quant/internal/application/adapter"
	"quant/internal/application/usecase"
	"quant/internal/domain/entity"
)

// actionLoggerService implements the adapter.ActionLogger interface.
type actionLoggerService struct {
	findAction usecase.FindAction
	saveAction usecase.SaveAction
}

// NewActionLoggerService creates a new ActionLogger service.
// Returns the adapter.ActionLogger interface, not the concrete type.
func NewActionLoggerService(
	findAction usecase.FindAction,
	saveAction usecase.SaveAction,
) adapter.ActionLogger {
	return &actionLoggerService{
		findAction: findAction,
		saveAction: saveAction,
	}
}

// LogAction creates and persists a new action for a session.
func (s *actionLoggerService) LogAction(sessionID string, actionType string, content string) (*entity.Action, error) {
	action := entity.Action{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      actionType,
		Content:   content,
		Timestamp: time.Now(),
	}

	err := s.saveAction.SaveAction(action)
	if err != nil {
		return nil, fmt.Errorf("failed to save action: %w", err)
	}

	return &action, nil
}

// GetActions returns all actions for a given session.
func (s *actionLoggerService) GetActions(sessionID string) ([]entity.Action, error) {
	actions, err := s.findAction.FindActionsBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	return actions, nil
}
