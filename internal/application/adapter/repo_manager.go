// Package adapter contains interfaces that application services implement.
package adapter

import (
	"quant/internal/domain/entity"
)

// RepoManager defines the service interface for repository management operations.
// This is the application adapter that the repoManagerService implements.
type RepoManager interface {
	OpenRepo(name string, path string) (*entity.Repo, error)
	ListRepos() ([]entity.Repo, error)
	GetRepo(id string) (*entity.Repo, error)
	RemoveRepo(id string) error
}
