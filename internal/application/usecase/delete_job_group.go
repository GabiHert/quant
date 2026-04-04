package usecase

// DeleteJobGroup defines the interface for deleting a job group.
type DeleteJobGroup interface {
	DeleteJobGroup(id string) error
}
