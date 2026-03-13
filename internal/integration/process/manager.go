// Package process contains the Claude CLI process manager implementation.
package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"quant/internal/integration/adapter"
)

// claudeProcess holds the running process and its PTY file.
type claudeProcess struct {
	cmd *exec.Cmd
	ptm *os.File // PTY master — read for output, write for input
}

// processManager implements the adapter.ProcessManager interface using os/exec + PTY.
type processManager struct {
	ctx       context.Context
	mu        sync.RWMutex
	processes map[string]*claudeProcess // keyed by sessionID
}

// NewProcessManager creates a new process manager for Claude CLI processes.
func NewProcessManager() adapter.ProcessManager {
	return &processManager{
		processes: make(map[string]*claudeProcess),
	}
}

// SetContext sets the Wails runtime context for emitting events.
func (m *processManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// Spawn starts a new Claude CLI process in a PTY in the given directory.
func (m *processManager) Spawn(sessionID string, directory string, conversationID string, skipPermissions bool) (int, error) {
	// Stop any existing process for this session.
	m.mu.RLock()
	_, exists := m.processes[sessionID]
	m.mu.RUnlock()
	if exists {
		_ = m.Stop(sessionID)
	}

	args := []string{}
	if conversationID != "" {
		args = append(args, "--resume", conversationID)
	}
	if skipPermissions {
		args = append(args, "--dangerously-skip-permissions")
	}

	cmd := exec.Command("claude", args...)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Start the process in a PTY.
	ptm, err := pty.Start(cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to start claude process in pty: %w", err)
	}

	// Set PTY size to a reasonable default.
	_ = pty.Setsize(ptm, &pty.Winsize{Rows: 40, Cols: 120})

	pid := cmd.Process.Pid
	cp := &claudeProcess{cmd: cmd, ptm: ptm}

	m.mu.Lock()
	m.processes[sessionID] = cp
	m.mu.Unlock()

	// Stream PTY output to the Wails frontend.
	go m.streamPTY(sessionID, ptm)

	// Monitor the process in a goroutine.
	go func() {
		_ = cmd.Wait()
		_ = ptm.Close()

		m.mu.Lock()
		delete(m.processes, sessionID)
		m.mu.Unlock()

		if m.ctx != nil {
			wailsRuntime.EventsEmit(m.ctx, "session:exited", sessionID)
		}
	}()

	return pid, nil
}

// streamPTY reads raw bytes from the PTY and sends them as Wails events.
func (m *processManager) streamPTY(sessionID string, ptm *os.File) {
	buf := make([]byte, 4096)
	for {
		n, err := ptm.Read(buf)
		if n > 0 && m.ctx != nil {
			// Send raw terminal data (including ANSI escape codes) to the frontend.
			wailsRuntime.EventsEmit(m.ctx, "session:output", map[string]string{
				"sessionId": sessionID,
				"data":      string(buf[:n]),
			})
		}
		if err != nil {
			break
		}
	}
}

// Stop terminates a running Claude process by session ID.
func (m *processManager) Stop(sessionID string) error {
	m.mu.RLock()
	cp, exists := m.processes[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no process running for session: %s", sessionID)
	}

	if cp.cmd.Process != nil {
		err := cp.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			return cp.cmd.Process.Kill()
		}
	}

	return nil
}

// SendMessage sends a message to a running Claude process via the PTY.
func (m *processManager) SendMessage(sessionID string, message string) error {
	m.mu.RLock()
	cp, exists := m.processes[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no process running for session: %s", sessionID)
	}

	if cp.ptm == nil {
		return fmt.Errorf("pty not available for session: %s", sessionID)
	}

	_, err := cp.ptm.Write([]byte(message + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write to pty: %w", err)
	}

	return nil
}
