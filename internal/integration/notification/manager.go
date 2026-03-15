// Package notification provides macOS system notification support.
package notification

import (
	"fmt"
	"os/exec"

	"quant/internal/application/usecase"
)

// manager implements the usecase.SendNotification interface using macOS osascript.
type manager struct{}

// NewManager creates a new notification manager.
// Returns the usecase.SendNotification interface, not the concrete type.
func NewManager() usecase.SendNotification {
	return &manager{}
}

// SendNotification displays a macOS system notification with the given title and message.
func (m *manager) SendNotification(title, message string) error {
	script := fmt.Sprintf(`display notification %q with title %q`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	return nil
}
