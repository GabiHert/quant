package usecase

// DeleteSession defines the interface for session deletion operations.
type DeleteSession interface {
	Delete(id string) error
}
