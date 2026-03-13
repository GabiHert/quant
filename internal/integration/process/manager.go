// Package process contains the Claude CLI process manager implementation.
package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	outputDir string                    // base dir for output files (~/.quant/sessions/)
}

// NewProcessManager creates a new process manager for Claude CLI processes.
func NewProcessManager() adapter.ProcessManager {
	homeDir, _ := os.UserHomeDir()
	outputDir := filepath.Join(homeDir, ".quant", "sessions")
	_ = os.MkdirAll(outputDir, 0755)

	return &processManager{
		processes: make(map[string]*claudeProcess),
		outputDir: outputDir,
	}
}

// SetContext sets the Wails runtime context for emitting events.
func (m *processManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// outputPath returns the path to the output file for a session.
func (m *processManager) outputPath(sessionID string) string {
	return filepath.Join(m.outputDir, sessionID+".log")
}

// Spawn starts a new Claude CLI process in a PTY in the given directory.
func (m *processManager) Spawn(sessionID string, directory string, conversationID string, skipPermissions bool, rows uint16, cols uint16) (int, error) {
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

	// Set PTY size from the frontend's terminal dimensions.
	if rows == 0 {
		rows = 24
	}
	if cols == 0 {
		cols = 80
	}
	_ = pty.Setsize(ptm, &pty.Winsize{Rows: rows, Cols: cols})

	pid := cmd.Process.Pid
	cp := &claudeProcess{cmd: cmd, ptm: ptm}

	m.mu.Lock()
	m.processes[sessionID] = cp
	m.mu.Unlock()

	// Open output file for appending.
	outputFile, err := os.OpenFile(m.outputPath(sessionID), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		outputFile = nil // non-fatal, just skip persistence
	}

	// Stream PTY output to the Wails frontend and to disk.
	go m.streamPTY(sessionID, ptm, outputFile)

	// Monitor the process in a goroutine.
	go func() {
		_ = cmd.Wait()
		_ = ptm.Close()
		if outputFile != nil {
			_ = outputFile.Close()
		}

		m.mu.Lock()
		delete(m.processes, sessionID)
		m.mu.Unlock()

		if m.ctx != nil {
			wailsRuntime.EventsEmit(m.ctx, "session:exited", sessionID)
		}
	}()

	return pid, nil
}

// streamPTY reads raw bytes from the PTY and sends them as Wails events + writes to disk.
func (m *processManager) streamPTY(sessionID string, ptm *os.File, outputFile *os.File) {
	buf := make([]byte, 4096)
	for {
		n, err := ptm.Read(buf)
		if n > 0 {
			data := buf[:n]

			// Write to disk for persistence.
			if outputFile != nil {
				_, _ = outputFile.Write(data)
			}

			// Send to frontend via Wails event.
			if m.ctx != nil {
				wailsRuntime.EventsEmit(m.ctx, "session:output", map[string]string{
					"sessionId": sessionID,
					"data":      string(data),
				})
			}
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

// SendMessage sends raw data to a running Claude process via the PTY.
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

	_, err := cp.ptm.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to pty: %w", err)
	}

	return nil
}

// Resize updates the PTY window size for a running session.
func (m *processManager) Resize(sessionID string, rows uint16, cols uint16) error {
	m.mu.RLock()
	cp, exists := m.processes[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no process running for session: %s", sessionID)
	}

	if cp.ptm == nil {
		return fmt.Errorf("pty not available for session: %s", sessionID)
	}

	return pty.Setsize(cp.ptm, &pty.Winsize{Rows: rows, Cols: cols})
}

// GetOutput returns the persisted output for a session from disk.
func (m *processManager) GetOutput(sessionID string) ([]byte, error) {
	data, err := os.ReadFile(m.outputPath(sessionID))
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}
	return data, nil
}
