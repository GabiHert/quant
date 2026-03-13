// Package persistence contains SQLite implementations of persistence interfaces.
package persistence

import (
	"database/sql"
	"fmt"

	"quant/internal/domain/entity"
	"quant/internal/integration/adapter"
	pdto "quant/internal/integration/persistence/dto"
)

// taskPersistence implements the adapter.TaskPersistence interface using SQLite.
type taskPersistence struct {
	db *sql.DB
}

// NewTaskPersistence creates a new SQLite task persistence implementation.
// Returns the adapter.TaskPersistence interface, not the concrete type.
func NewTaskPersistence(db *sql.DB) adapter.TaskPersistence {
	return &taskPersistence{db: db}
}

// FindTaskByID retrieves a task by its ID.
func (p *taskPersistence) FindTaskByID(id string) (*entity.Task, error) {
	query := `SELECT id, repo_id, tag, name, created_at, updated_at FROM tasks WHERE id = ?`

	var row pdto.TaskRow
	err := p.db.QueryRow(query, id).Scan(
		&row.ID, &row.RepoID, &row.Tag, &row.Name, &row.CreatedAt, &row.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find task by id: %w", err)
	}

	task := row.ToEntity()
	return &task, nil
}

// FindTasksByRepoID retrieves all tasks for a given repository.
func (p *taskPersistence) FindTasksByRepoID(repoID string) ([]entity.Task, error) {
	query := `SELECT id, repo_id, tag, name, created_at, updated_at FROM tasks WHERE repo_id = ? ORDER BY created_at DESC`

	rows, err := p.db.Query(query, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by repo id: %w", err)
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var row pdto.TaskRow
		err := rows.Scan(
			&row.ID, &row.RepoID, &row.Tag, &row.Name, &row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, row.ToEntity())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

// FindAllTasks retrieves all tasks.
func (p *taskPersistence) FindAllTasks() ([]entity.Task, error) {
	query := `SELECT id, repo_id, tag, name, created_at, updated_at FROM tasks ORDER BY created_at DESC`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to find all tasks: %w", err)
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var row pdto.TaskRow
		err := rows.Scan(
			&row.ID, &row.RepoID, &row.Tag, &row.Name, &row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, row.ToEntity())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

// SaveTask persists a new task to the database.
func (p *taskPersistence) SaveTask(task entity.Task) error {
	row := pdto.TaskRowFromEntity(task)

	query := `INSERT INTO tasks (id, repo_id, tag, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := p.db.Exec(query, row.ID, row.RepoID, row.Tag, row.Name, row.CreatedAt, row.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return nil
}

// DeleteTask removes a task by its ID.
func (p *taskPersistence) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return nil
}

// UpdateTask updates all fields of a task.
func (p *taskPersistence) UpdateTask(task entity.Task) error {
	row := pdto.TaskRowFromEntity(task)

	query := `UPDATE tasks SET repo_id = ?, tag = ?, name = ?, updated_at = ? WHERE id = ?`

	result, err := p.db.Exec(query, row.RepoID, row.Tag, row.Name, row.UpdatedAt, row.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return nil
}
