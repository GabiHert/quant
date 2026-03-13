// Package dto contains persistence data transfer objects for SQLite row mapping.
package dto

import (
	"time"

	"quant/internal/domain/entity"
)

// ActionRow represents an action row in the SQLite database.
type ActionRow struct {
	ID        string
	SessionID string
	Type      string
	Content   string
	Timestamp string
}

// ToEntity converts an ActionRow to a domain entity.
func (r ActionRow) ToEntity() entity.Action {
	timestamp, _ := time.Parse(time.RFC3339, r.Timestamp)

	return entity.Action{
		ID:        r.ID,
		SessionID: r.SessionID,
		Type:      r.Type,
		Content:   r.Content,
		Timestamp: timestamp,
	}
}

// ActionRowFromEntity converts a domain entity to an ActionRow.
func ActionRowFromEntity(action entity.Action) ActionRow {
	return ActionRow{
		ID:        action.ID,
		SessionID: action.SessionID,
		Type:      action.Type,
		Content:   action.Content,
		Timestamp: action.Timestamp.Format(time.RFC3339),
	}
}
