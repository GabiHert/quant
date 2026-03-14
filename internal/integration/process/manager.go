// Package process contains the Claude CLI process manager implementation.
package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"unicode/utf8"

	"github.com/creack/pty"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"quant/internal/integration/adapter"
)

// claudeProcess holds the running process and its PTY master.
type claudeProcess struct {
	cmd *exec.Cmd
	ptm *os.File // PTY master
}

// processManager implements the adapter.ProcessManager interface using PTY.
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

// Spawn starts claude in a PTY and streams output to the frontend.
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

	ptm, err := pty.Start(cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to start claude in PTY: %w", err)
	}

	// Set initial PTY size.
	_ = pty.Setsize(ptm, &pty.Winsize{Rows: rows, Cols: cols})

	cp := &claudeProcess{cmd: cmd, ptm: ptm}

	m.mu.Lock()
	m.processes[sessionID] = cp
	m.mu.Unlock()

	pid := cmd.Process.Pid

	// Open output file for appending raw terminal output.
	outputFile, err := os.OpenFile(m.outputPath(sessionID), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		outputFile = nil // non-fatal, just skip persistence
	}

	// Stream PTY output in a goroutine.
	go func() {
		buf := make([]byte, 32*1024)
		var carry []byte // buffer for incomplete UTF-8 sequences at chunk boundaries

		for {
			n, readErr := ptm.Read(buf)
			if n > 0 {
				data := buf[:n]

				// Prepend any carry from previous read.
				if len(carry) > 0 {
					data = append(carry, data...)
					carry = nil
				}

				// Check for incomplete UTF-8 at the end.
				// Find the last valid UTF-8 boundary.
				validEnd := len(data)
				for validEnd > 0 && !utf8.Valid(data[:validEnd]) {
					validEnd--
				}

				// If the tail is an incomplete sequence, carry it over.
				if validEnd < len(data) {
					carry = make([]byte, len(data)-validEnd)
					copy(carry, data[validEnd:])
					data = data[:validEnd]
				}

				if len(data) > 0 {
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
			}

			if readErr != nil {
				break
			}
		}

		// Wait for process to finish.
		_ = cmd.Wait()

		if outputFile != nil {
			_ = outputFile.Close()
		}

		m.mu.Lock()
		delete(m.processes, sessionID)
		m.mu.Unlock()

		// Notify frontend that the process exited.
		if m.ctx != nil {
			wailsRuntime.EventsEmit(m.ctx, "session:exited", map[string]string{
				"sessionId": sessionID,
			})
		}
	}()

	return pid, nil
}

// Stop terminates a running Claude process by session ID.
func (m *processManager) Stop(sessionID string) error {
	m.mu.RLock()
	cp, exists := m.processes[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no process running for session: %s", sessionID)
	}

	// Close PTY master — this sends SIGHUP to the process.
	_ = cp.ptm.Close()

	// Also kill the process explicitly in case it doesn't respond to SIGHUP.
	if cp.cmd.Process != nil {
		_ = cp.cmd.Process.Kill()
	}

	return nil
}

// SendMessage writes raw data to the PTY (for terminal input).
func (m *processManager) SendMessage(sessionID string, message string) error {
	m.mu.RLock()
	cp, exists := m.processes[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no process running for session: %s", sessionID)
	}

	_, err := cp.ptm.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to PTY: %w", err)
	}

	return nil
}

// Resize resizes the PTY for the given session.
func (m *processManager) Resize(sessionID string, rows uint16, cols uint16) error {
	m.mu.RLock()
	cp, exists := m.processes[sessionID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no process running for session: %s", sessionID)
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
