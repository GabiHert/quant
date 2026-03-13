// Package actiontype contains string constants for action type values.
package actiontype

const (
	// UserMessage indicates a message sent by the user.
	UserMessage = "user_message"

	// ClaudeRead indicates Claude read a file.
	ClaudeRead = "claude_read"

	// ClaudeEdit indicates Claude edited a file.
	ClaudeEdit = "claude_edit"

	// ClaudeCreate indicates Claude created a file.
	ClaudeCreate = "claude_create"

	// ClaudeBash indicates Claude executed a bash command.
	ClaudeBash = "claude_bash"

	// ClaudeResult indicates Claude returned a result.
	ClaudeResult = "claude_result"
)
