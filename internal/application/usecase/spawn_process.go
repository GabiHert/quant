package usecase

// SpawnProcess defines the interface for managing Claude CLI processes.
type SpawnProcess interface {
	Spawn(sessionID string, directory string, conversationID string, skipPermissions bool, rows uint16, cols uint16) (int, error)
	Stop(sessionID string) error
	SendMessage(sessionID string, message string) error
	Resize(sessionID string, rows uint16, cols uint16) error
	GetOutput(sessionID string) ([]byte, error)
}
