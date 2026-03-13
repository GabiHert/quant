package usecase

// SpawnProcess defines the interface for managing Claude CLI processes.
type SpawnProcess interface {
	Spawn(sessionID string, directory string, conversationID string, skipPermissions bool) (int, error)
	Stop(sessionID string) error
	SendMessage(sessionID string, message string) error
}
