// Package dto contains data transfer objects for the entrypoint layer.
package dto

import "quant/internal/domain/entity"

// DiffFileResponse represents a changed file in the working directory of a session.
type DiffFileResponse struct {
	Path    string `json:"path"`
	Status  string `json:"status"`
	OldPath string `json:"oldPath"`
}

// DiffFileResponseFromEntity converts a domain entity to a DiffFileResponse DTO.
func DiffFileResponseFromEntity(f entity.DiffFile) DiffFileResponse {
	return DiffFileResponse{
		Path:    f.Path,
		Status:  f.Status,
		OldPath: f.OldPath,
	}
}

// DiffFileResponseListFromEntities converts a slice of domain entities to a slice of DiffFileResponse DTOs.
func DiffFileResponseListFromEntities(files []entity.DiffFile) []DiffFileResponse {
	responses := make([]DiffFileResponse, len(files))
	for i, f := range files {
		responses[i] = DiffFileResponseFromEntity(f)
	}
	return responses
}
