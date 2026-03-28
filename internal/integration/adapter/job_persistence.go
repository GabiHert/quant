// Package adapter contains integration adapter interfaces that combine multiple usecase interfaces.
package adapter

import (
	"quant/internal/application/usecase"
)

// JobPersistence combines all job-related persistence usecase interfaces.
// Integration persistence implementations must implement this interface.
type JobPersistence interface {
	usecase.FindJob
	usecase.SaveJob
	usecase.UpdateJob
	usecase.DeleteJob
	usecase.FindJobTrigger
	usecase.SaveJobTrigger
	usecase.FindJobRun
	usecase.SaveJobRun
}
