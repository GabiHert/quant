package usecase

// DeleteWorkspace defines the interface for deleting a workspace.
type DeleteWorkspace interface {
	DeleteWorkspace(id string) error
}
