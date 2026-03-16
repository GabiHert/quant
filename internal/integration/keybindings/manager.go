// Package keybindings manages Claude Code keybindings configuration.
package keybindings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"quant/internal/application/usecase"
)

type keybindingsFile struct {
	Schema   string    `json:"$schema,omitempty"`
	Bindings []binding `json:"bindings"`
}

type binding struct {
	Context  string            `json:"context"`
	Bindings map[string]string `json:"bindings"`
}

// manager implements the usecase.SetNewLineKey interface.
type manager struct{}

// NewManager creates a new keybindings manager.
func NewManager() usecase.SetNewLineKey {
	return &manager{}
}

// SetNewLineKey configures the Claude Code newline keybinding.
// When "shift+enter" is selected, it writes the keybinding to ~/.claude/keybindings.json.
// When "backslash+enter" (default), it removes the override.
func (m *manager) SetNewLineKey(key string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	keybindingsPath := filepath.Join(homeDir, ".claude", "keybindings.json")

	if key == "shift+enter" {
		kb := keybindingsFile{
			Bindings: []binding{
				{
					Context: "Chat",
					Bindings: map[string]string{
						"shift+enter": "chat:newline",
					},
				},
			},
		}

		data, err := json.MarshalIndent(kb, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal keybindings: %w", err)
		}

		if err := os.WriteFile(keybindingsPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write keybindings: %w", err)
		}
	} else {
		// Remove the keybindings file to restore defaults.
		err := os.Remove(keybindingsPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove keybindings: %w", err)
		}
	}

	return nil
}
