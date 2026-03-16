package usecase

// SetNewLineKey defines the interface for configuring the Claude CLI newline keybinding.
type SetNewLineKey interface {
	SetNewLineKey(key string) error
}
