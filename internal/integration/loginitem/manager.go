// Package loginitem manages macOS LaunchAgent plist for start-on-login functionality.
package loginitem

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"quant/internal/application/usecase"
)

const (
	plistLabel = "com.quant.app"
	plistName  = "com.quant.app.plist"
)

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{ .Label }}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{ .AppPath }}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`))

// manager implements the usecase.SetLoginItem interface.
type manager struct{}

// NewManager creates a new login item manager.
// Returns the usecase.SetLoginItem interface, not the concrete type.
func NewManager() usecase.SetLoginItem {
	return &manager{}
}

// SetLoginItem enables or disables the LaunchAgent for start-on-login.
func (m *manager) SetLoginItem(enabled bool) error {
	return setEnabled(enabled)
}

// setEnabled enables or disables the LaunchAgent for start-on-login.
func setEnabled(enabled bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	plistPath := filepath.Join(launchAgentsDir, plistName)

	if !enabled {
		err := os.Remove(plistPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove launch agent: %w", err)
		}
		return nil
	}

	// Find the app binary path.
	appPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Ensure LaunchAgents directory exists.
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("failed to create plist file: %w", err)
	}
	defer f.Close()

	err = plistTemplate.Execute(f, struct {
		Label   string
		AppPath string
	}{
		Label:   plistLabel,
		AppPath: appPath,
	})
	if err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	return nil
}
