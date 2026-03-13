// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// RepoPersistence combines all repo-related persistence usecase interfaces.
// Integration persistence implementations must implement this interface.
type RepoPersistence interface {
	usecase.FindRepo
	usecase.SaveRepo
	usecase.DeleteRepo
	usecase.UpdateRepo
}
