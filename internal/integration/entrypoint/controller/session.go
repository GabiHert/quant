// Package controller contains entrypoint controllers bound to the Wails runtime.
package controller

import (
	"context"

	"quant/internal/application/adapter"
	intadapter "quant/internal/integration/adapter"
	"quant/internal/integration/entrypoint/dto"
)

// sessionController implements the integration adapter.SessionController interface.
// It is bound to the Wails runtime and exposes session management operations to the frontend.
type sessionController struct {
	ctx            context.Context
	sessionManager adapter.SessionManager
}

// NewSessionController creates a new session controller.
// Returns the intadapter.SessionController interface, not the concrete type.
func NewSessionController(sessionManager adapter.SessionManager) intadapter.SessionController {
	return &sessionController{
		sessionManager: sessionManager,
	}
}

// OnStartup is called when the Wails app starts. The context is saved for runtime method calls.
func (c *sessionController) OnStartup(ctx context.Context) {
	c.ctx = ctx
}

// OnShutdown is called when the Wails app is shutting down.
func (c *sessionController) OnShutdown(_ context.Context) {
	// Clean up any running sessions if needed.
}

// CreateSession creates a new session and returns its response DTO.
func (c *sessionController) CreateSession(request dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	session, err := c.sessionManager.CreateSession(request.Name, request.Description, request.RepoID, request.TaskID, request.UseWorktree, request.SkipPermissions)
	if err != nil {
		return nil, err
	}

	return dto.SessionResponseFromEntityPtr(session), nil
}

// StartSession starts a session by spawning a Claude process.
func (c *sessionController) StartSession(id string, rows int, cols int) error {
	return c.sessionManager.StartSession(id, rows, cols)
}

// StopSession stops a running session.
func (c *sessionController) StopSession(id string) error {
	return c.sessionManager.StopSession(id)
}

// ResumeSession resumes a paused session.
func (c *sessionController) ResumeSession(id string, rows int, cols int) error {
	return c.sessionManager.ResumeSession(id, rows, cols)
}

// DeleteSession deletes a session.
func (c *sessionController) DeleteSession(id string) error {
	return c.sessionManager.DeleteSession(id)
}

// ListSessions returns all sessions as response DTOs.
func (c *sessionController) ListSessions() ([]dto.SessionResponse, error) {
	sessions, err := c.sessionManager.ListSessions()
	if err != nil {
		return nil, err
	}

	return dto.SessionResponseListFromEntities(sessions), nil
}

// ListSessionsByRepo returns all sessions for a given repository as response DTOs.
func (c *sessionController) ListSessionsByRepo(repoID string) ([]dto.SessionResponse, error) {
	sessions, err := c.sessionManager.ListSessionsByRepo(repoID)
	if err != nil {
		return nil, err
	}

	return dto.SessionResponseListFromEntities(sessions), nil
}

// ListSessionsByTask returns all sessions for a given task as response DTOs.
func (c *sessionController) ListSessionsByTask(taskID string) ([]dto.SessionResponse, error) {
	sessions, err := c.sessionManager.ListSessionsByTask(taskID)
	if err != nil {
		return nil, err
	}

	return dto.SessionResponseListFromEntities(sessions), nil
}

// GetSession returns a single session by ID as a response DTO.
func (c *sessionController) GetSession(id string) (*dto.SessionResponse, error) {
	session, err := c.sessionManager.GetSession(id)
	if err != nil {
		return nil, err
	}

	return dto.SessionResponseFromEntityPtr(session), nil
}

// SendMessage sends a message to a running session's Claude process.
func (c *sessionController) SendMessage(id string, message string) error {
	return c.sessionManager.SendMessage(id, message)
}

// ResizeTerminal updates the PTY size for a running session.
func (c *sessionController) ResizeTerminal(id string, rows int, cols int) error {
	return c.sessionManager.ResizeTerminal(id, rows, cols)
}

// GetSessionOutput returns the persisted terminal output for a session.
func (c *sessionController) GetSessionOutput(id string) (string, error) {
	return c.sessionManager.GetSessionOutput(id)
}
