package usecase

// DeleteJob defines the interface for deleting a job.
type DeleteJob interface {
	DeleteJob(id string) error
}
