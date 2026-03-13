// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// ActionPersistence combines all action-related persistence usecase interfaces.
// Integration persistence implementations must implement this interface.
type ActionPersistence interface {
	usecase.FindAction
	usecase.SaveAction
}
