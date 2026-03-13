package usecase

// DeleteTask defines the interface for task deletion operations.
type DeleteTask interface {
	DeleteTask(id string) error
}
