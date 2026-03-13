// Package dto contains data transfer objects for the entrypoint layer.
package dto

import (
	"quant/internal/domain/entity"
)

// ActionResponse represents the response payload for action data.
type ActionResponse struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// ActionResponseFromEntity converts a domain entity to an ActionResponse DTO.
func ActionResponseFromEntity(action entity.Action) ActionResponse {
	return ActionResponse{
		ID:        action.ID,
		SessionID: action.SessionID,
		Type:      action.Type,
		Content:   action.Content,
		Timestamp: action.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ActionResponseListFromEntities converts a slice of domain entities to a slice of ActionResponse DTOs.
func ActionResponseListFromEntities(actions []entity.Action) []ActionResponse {
	responses := make([]ActionResponse, len(actions))
	for i, action := range actions {
		responses[i] = ActionResponseFromEntity(action)
	}
	return responses
}
