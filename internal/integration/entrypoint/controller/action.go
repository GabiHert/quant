// Package controller contains entrypoint controllers bound to the Wails runtime.
package controller

import (
	"context"

	"quant/internal/application/adapter"
	intadapter "quant/internal/integration/adapter"
	"quant/internal/integration/entrypoint/dto"
)

// actionController implements the integration adapter.ActionController interface.
// It is bound to the Wails runtime and exposes action retrieval operations to the frontend.
type actionController struct {
	ctx          context.Context
	actionLogger adapter.ActionLogger
}

// NewActionController creates a new action controller.
// Returns the intadapter.ActionController interface, not the concrete type.
func NewActionController(actionLogger adapter.ActionLogger) intadapter.ActionController {
	return &actionController{
		actionLogger: actionLogger,
	}
}

// OnStartup is called when the Wails app starts. The context is saved for runtime method calls.
func (c *actionController) OnStartup(ctx context.Context) {
	c.ctx = ctx
}

// OnShutdown is called when the Wails app is shutting down.
func (c *actionController) OnShutdown(_ context.Context) {
	// Clean up if needed.
}

// GetActions returns all actions for a given session as response DTOs.
func (c *actionController) GetActions(sessionID string) ([]dto.ActionResponse, error) {
	actions, err := c.actionLogger.GetActions(sessionID)
	if err != nil {
		return nil, err
	}

	return dto.ActionResponseListFromEntities(actions), nil
}
