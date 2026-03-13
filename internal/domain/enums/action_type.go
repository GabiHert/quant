// Package enums contains domain enumeration types.
package enums

import (
	"quant/internal/domain/enums/actiontype"
)

// ActionType represents the possible types of an action.
type ActionType struct {
	value string
}

// Value returns the string representation of the action type.
func (a ActionType) Value() string {
	return a.value
}

// String returns the string representation of the action type.
func (a ActionType) String() string {
	return a.value
}

// IsValid returns true if the action type is a recognized value.
func (a ActionType) IsValid() bool {
	switch a.value {
	case actiontype.UserMessage,
		actiontype.ClaudeRead,
		actiontype.ClaudeEdit,
		actiontype.ClaudeCreate,
		actiontype.ClaudeBash,
		actiontype.ClaudeResult:
		return true
	default:
		return false
	}
}

// NewActionType creates a new ActionType from a string value.
func NewActionType(value string) ActionType {
	return ActionType{value: value}
}

// ActionTypeUserMessage returns the user_message action type.
func ActionTypeUserMessage() ActionType {
	return ActionType{value: actiontype.UserMessage}
}

// ActionTypeClaudeRead returns the claude_read action type.
func ActionTypeClaudeRead() ActionType {
	return ActionType{value: actiontype.ClaudeRead}
}

// ActionTypeClaudeEdit returns the claude_edit action type.
func ActionTypeClaudeEdit() ActionType {
	return ActionType{value: actiontype.ClaudeEdit}
}

// ActionTypeClaudeCreate returns the claude_create action type.
func ActionTypeClaudeCreate() ActionType {
	return ActionType{value: actiontype.ClaudeCreate}
}

// ActionTypeClaudeBash returns the claude_bash action type.
func ActionTypeClaudeBash() ActionType {
	return ActionType{value: actiontype.ClaudeBash}
}

// ActionTypeClaudeResult returns the claude_result action type.
func ActionTypeClaudeResult() ActionType {
	return ActionType{value: actiontype.ClaudeResult}
}
