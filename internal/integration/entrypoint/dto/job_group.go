// Package dto contains data transfer objects for the entrypoint layer.
package dto

import (
	"quant/internal/domain/entity"
)

// CreateJobGroupRequest represents the request payload for creating a new job group.
type CreateJobGroupRequest struct {
	Name        string   `json:"name"`
	JobIDs      []string `json:"jobIds"`
	WorkspaceID string   `json:"workspaceId"`
}

// UpdateJobGroupRequest represents the request payload for updating an existing job group.
type UpdateJobGroupRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	JobIDs      []string `json:"jobIds"`
	WorkspaceID string   `json:"workspaceId"`
}

// JobGroupResponse represents the response payload for job group data.
type JobGroupResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	JobIDs      []string `json:"jobIds"`
	WorkspaceID string   `json:"workspaceId"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// JobGroupResponseFromEntity converts a domain entity to a JobGroupResponse DTO.
func JobGroupResponseFromEntity(group entity.JobGroup) JobGroupResponse {
	jobIDs := group.JobIDs
	if jobIDs == nil {
		jobIDs = []string{}
	}
	return JobGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		JobIDs:      jobIDs,
		WorkspaceID: group.WorkspaceID,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// JobGroupResponseFromEntityPtr converts a domain entity pointer to a JobGroupResponse DTO pointer.
func JobGroupResponseFromEntityPtr(group *entity.JobGroup) *JobGroupResponse {
	if group == nil {
		return nil
	}
	response := JobGroupResponseFromEntity(*group)
	return &response
}

// JobGroupResponseListFromEntities converts a slice of domain entities to a slice of JobGroupResponse DTOs.
func JobGroupResponseListFromEntities(groups []entity.JobGroup) []JobGroupResponse {
	responses := make([]JobGroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = JobGroupResponseFromEntity(group)
	}
	return responses
}
