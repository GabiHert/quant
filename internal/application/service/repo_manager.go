// Package service contains application service implementations with business logic.
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"quant/internal/application/adapter"
	"quant/internal/application/usecase"
	"quant/internal/domain/entity"
)

// repoManagerService implements the adapter.RepoManager interface.
type repoManagerService struct {
	findRepo   usecase.FindRepo
	saveRepo   usecase.SaveRepo
	deleteRepo usecase.DeleteRepo
}

// NewRepoManagerService creates a new RepoManager service.
// Returns the adapter.RepoManager interface, not the concrete type.
func NewRepoManagerService(
	findRepo usecase.FindRepo,
	saveRepo usecase.SaveRepo,
	deleteRepo usecase.DeleteRepo,
) adapter.RepoManager {
	return &repoManagerService{
		findRepo:   findRepo,
		saveRepo:   saveRepo,
		deleteRepo: deleteRepo,
	}
}

// OpenRepo registers a new repository with the given name and path.
func (s *repoManagerService) OpenRepo(name string, path string) (*entity.Repo, error) {
	now := time.Now()
	repo := entity.Repo{
		ID:        uuid.New().String(),
		Name:      name,
		Path:      path,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.saveRepo.SaveRepo(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to save repo: %w", err)
	}

	return &repo, nil
}

// ListRepos returns all registered repositories.
func (s *repoManagerService) ListRepos() ([]entity.Repo, error) {
	repos, err := s.findRepo.FindAllRepos()
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	return repos, nil
}

// GetRepo returns a repository by ID.
func (s *repoManagerService) GetRepo(id string) (*entity.Repo, error) {
	repo, err := s.findRepo.FindRepoByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	if repo == nil {
		return nil, fmt.Errorf("repo not found: %s", id)
	}

	return repo, nil
}

// RemoveRepo removes a repository registration by ID.
func (s *repoManagerService) RemoveRepo(id string) error {
	repo, err := s.findRepo.FindRepoByID(id)
	if err != nil {
		return fmt.Errorf("failed to find repo: %w", err)
	}

	if repo == nil {
		return fmt.Errorf("repo not found: %s", id)
	}

	err = s.deleteRepo.DeleteRepo(id)
	if err != nil {
		return fmt.Errorf("failed to delete repo: %w", err)
	}

	return nil
}
