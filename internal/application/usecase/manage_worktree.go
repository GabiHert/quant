package usecase

// WorktreeInfo holds information about a git worktree.
type WorktreeInfo struct {
	Path   string
	Branch string
}

// ManageWorktree defines the interface for git worktree operations.
type ManageWorktree interface {
	Create(repoDir string, branchName string) (WorktreeInfo, error)
	Delete(worktreePath string) error
	List(repoDir string) ([]WorktreeInfo, error)
}
