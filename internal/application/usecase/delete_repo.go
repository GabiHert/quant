package usecase

// DeleteRepo defines the interface for repo deletion operations.
type DeleteRepo interface {
	DeleteRepo(id string) error
}
