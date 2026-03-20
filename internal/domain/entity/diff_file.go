// Package entity contains domain entities representing core business objects.
package entity

// DiffFile represents a file that has changed in the working directory of a session.
// Status follows git's short format: "M" (modified), "A" (added), "D" (deleted),
// "R" (renamed), "?" (untracked).
type DiffFile struct {
	Path    string
	Status  string
	OldPath string // populated only when Status is "R"
}
