// Package worktree contains the git worktree manager implementation.
package worktree

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"quant/internal/application/usecase"
	"quant/internal/integration/adapter"
)

// worktreeManager implements the adapter.WorktreeManager interface using git CLI commands.
type worktreeManager struct{}

// NewWorktreeManager creates a new git worktree manager.
// Returns the adapter.WorktreeManager interface, not the concrete type.
func NewWorktreeManager() adapter.WorktreeManager {
	return &worktreeManager{}
}

// Create creates a new git worktree with the given branch name.
// The worktree is created as a sibling directory of the repository.
func (m *worktreeManager) Create(repoDir string, branchName string) (usecase.WorktreeInfo, error) {
	worktreePath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"-"+branchName)

	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return usecase.WorktreeInfo{}, fmt.Errorf("failed to create worktree: %s: %w", string(output), err)
	}

	return usecase.WorktreeInfo{
		Path:   worktreePath,
		Branch: branchName,
	}, nil
}

// Delete removes a git worktree.
func (m *worktreeManager) Delete(worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", worktreePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %s: %w", string(output), err)
	}

	return nil
}

// List returns all git worktrees for the given repository.
func (m *worktreeManager) List(repoDir string) ([]usecase.WorktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []usecase.WorktreeInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	var current usecase.WorktreeInfo
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = usecase.WorktreeInfo{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			// Strip refs/heads/ prefix.
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}
