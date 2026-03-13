// Package persistence contains SQLite implementations of persistence interfaces.
package persistence

import (
	"database/sql"
	"fmt"

	"quant/internal/domain/entity"
	"quant/internal/integration/adapter"
	pdto "quant/internal/integration/persistence/dto"
)

// actionPersistence implements the adapter.ActionPersistence interface using SQLite.
type actionPersistence struct {
	db *sql.DB
}

// NewActionPersistence creates a new SQLite action persistence implementation.
// Returns the adapter.ActionPersistence interface, not the concrete type.
func NewActionPersistence(db *sql.DB) adapter.ActionPersistence {
	return &actionPersistence{db: db}
}

// FindActionsBySessionID retrieves all actions for a given session.
func (p *actionPersistence) FindActionsBySessionID(sessionID string) ([]entity.Action, error) {
	query := `SELECT id, session_id, type, content, timestamp FROM actions WHERE session_id = ? ORDER BY timestamp ASC`

	rows, err := p.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find actions by session id: %w", err)
	}
	defer rows.Close()

	var actions []entity.Action
	for rows.Next() {
		var row pdto.ActionRow
		err := rows.Scan(
			&row.ID, &row.SessionID, &row.Type, &row.Content, &row.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action row: %w", err)
		}
		actions = append(actions, row.ToEntity())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating action rows: %w", err)
	}

	return actions, nil
}

// SaveAction persists a new action to the database.
func (p *actionPersistence) SaveAction(action entity.Action) error {
	row := pdto.ActionRowFromEntity(action)

	query := `INSERT INTO actions (id, session_id, type, content, timestamp) VALUES (?, ?, ?, ?, ?)`

	_, err := p.db.Exec(query, row.ID, row.SessionID, row.Type, row.Content, row.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save action: %w", err)
	}

	return nil
}
