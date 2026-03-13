// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// TaskPersistence combines all task-related persistence usecase interfaces.
// Integration persistence implementations must implement this interface.
type TaskPersistence interface {
	usecase.FindTask
	usecase.SaveTask
	usecase.DeleteTask
	usecase.UpdateTask
}
