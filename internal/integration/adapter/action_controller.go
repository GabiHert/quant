package adapter

import (
	"context"

	"quant/internal/integration/entrypoint/dto"
)

// ActionController defines the interface for the action entrypoint controller.
// This interface is what the Wails app binds to.
type ActionController interface {
	OnStartup(ctx context.Context)
	OnShutdown(ctx context.Context)
	GetActions(sessionID string) ([]dto.ActionResponse, error)
}
