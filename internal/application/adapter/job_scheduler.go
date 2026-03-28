// Package adapter contains interfaces that application services implement.
package adapter

// JobScheduler defines the interface for the job scheduling service.
type JobScheduler interface {
	Start()
	Stop()
}
