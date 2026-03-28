// Package mcp implements an MCP server that exposes Quant's job management
// to Claude Code and other MCP clients.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	appAdapter "quant/internal/application/adapter"
	"quant/internal/domain/entity"
)

// QuantMCPServer wraps an MCP server that exposes job management tools.
type QuantMCPServer struct {
	jobManager appAdapter.JobManager
	httpServer *http.Server
}

// NewQuantMCPServer creates a new MCP server with all job management tools registered.
func NewQuantMCPServer(jobManager appAdapter.JobManager) *QuantMCPServer {
	mcpServer := server.NewMCPServer("quant", "1.0.0")

	s := &QuantMCPServer{jobManager: jobManager}

	s.registerTools(mcpServer)

	streamable := server.NewStreamableHTTPServer(mcpServer)

	mux := http.NewServeMux()
	mux.Handle("/mcp", streamable)

	s.httpServer = &http.Server{
		Addr:    ":52945",
		Handler: mux,
	}

	return s
}

// Start begins listening for MCP requests in a background goroutine.
func (s *QuantMCPServer) Start() error {
	go s.httpServer.ListenAndServe()
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *QuantMCPServer) Stop() error {
	return s.httpServer.Shutdown(context.Background())
}

func (s *QuantMCPServer) registerTools(mcpServer *server.MCPServer) {
	// 1. list_jobs
	mcpServer.AddTool(
		mcp.NewTool("list_jobs",
			mcp.WithDescription("List all configured jobs"),
		),
		s.handleListJobs,
	)

	// 2. get_job
	mcpServer.AddTool(
		mcp.NewTool("get_job",
			mcp.WithDescription("Get a job by ID"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID")),
		),
		s.handleGetJob,
	)

	// 3. create_job
	mcpServer.AddTool(
		mcp.NewTool("create_job",
			mcp.WithDescription("Create a new job"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Job name")),
			mcp.WithString("description", mcp.Description("Job description")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Job type: claude or bash")),
			mcp.WithString("workingDirectory", mcp.Description("Working directory for the job")),
			mcp.WithBoolean("scheduleEnabled", mcp.Description("Whether scheduling is enabled")),
			mcp.WithString("scheduleType", mcp.Description("Schedule type: recurring or one_time")),
			mcp.WithString("cronExpression", mcp.Description("Cron expression for scheduling")),
			mcp.WithNumber("scheduleInterval", mcp.Description("Schedule interval in minutes")),
			mcp.WithNumber("timeoutSeconds", mcp.Description("Timeout in seconds")),
			mcp.WithString("prompt", mcp.Description("Claude prompt")),
			mcp.WithBoolean("allowBypass", mcp.Description("Allow bypass of confirmation prompts")),
			mcp.WithBoolean("autonomousMode", mcp.Description("Run in autonomous mode")),
			mcp.WithNumber("maxRetries", mcp.Description("Maximum number of retries")),
			mcp.WithString("model", mcp.Description("Claude model to use")),
			mcp.WithString("claudeCommand", mcp.Description("Custom Claude command")),
			mcp.WithString("interpreter", mcp.Description("Script interpreter for bash jobs")),
			mcp.WithString("scriptContent", mcp.Description("Script content for bash jobs")),
		),
		s.handleCreateJob,
	)

	// 4. update_job
	mcpServer.AddTool(
		mcp.NewTool("update_job",
			mcp.WithDescription("Update an existing job"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID")),
			mcp.WithString("name", mcp.Description("Job name")),
			mcp.WithString("description", mcp.Description("Job description")),
			mcp.WithString("type", mcp.Description("Job type: claude or bash")),
			mcp.WithString("workingDirectory", mcp.Description("Working directory for the job")),
			mcp.WithBoolean("scheduleEnabled", mcp.Description("Whether scheduling is enabled")),
			mcp.WithString("scheduleType", mcp.Description("Schedule type: recurring or one_time")),
			mcp.WithString("cronExpression", mcp.Description("Cron expression for scheduling")),
			mcp.WithNumber("scheduleInterval", mcp.Description("Schedule interval in minutes")),
			mcp.WithNumber("timeoutSeconds", mcp.Description("Timeout in seconds")),
			mcp.WithString("prompt", mcp.Description("Claude prompt")),
			mcp.WithBoolean("allowBypass", mcp.Description("Allow bypass of confirmation prompts")),
			mcp.WithBoolean("autonomousMode", mcp.Description("Run in autonomous mode")),
			mcp.WithNumber("maxRetries", mcp.Description("Maximum number of retries")),
			mcp.WithString("model", mcp.Description("Claude model to use")),
			mcp.WithString("claudeCommand", mcp.Description("Custom Claude command")),
			mcp.WithString("interpreter", mcp.Description("Script interpreter for bash jobs")),
			mcp.WithString("scriptContent", mcp.Description("Script content for bash jobs")),
		),
		s.handleUpdateJob,
	)

	// 5. delete_job
	mcpServer.AddTool(
		mcp.NewTool("delete_job",
			mcp.WithDescription("Delete a job by ID"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID")),
		),
		s.handleDeleteJob,
	)

	// 6. run_job
	mcpServer.AddTool(
		mcp.NewTool("run_job",
			mcp.WithDescription("Run a job immediately"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID to run")),
		),
		s.handleRunJob,
	)

	// 7. get_run
	mcpServer.AddTool(
		mcp.NewTool("get_run",
			mcp.WithDescription("Get a job run by ID"),
			mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID")),
		),
		s.handleGetRun,
	)

	// 8. list_runs
	mcpServer.AddTool(
		mcp.NewTool("list_runs",
			mcp.WithDescription("List all runs for a job"),
			mcp.WithString("jobId", mcp.Required(), mcp.Description("Job ID")),
		),
		s.handleListRuns,
	)

	// 9. get_run_output
	mcpServer.AddTool(
		mcp.NewTool("get_run_output",
			mcp.WithDescription("Get the output of a job run"),
			mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID")),
		),
		s.handleGetRunOutput,
	)

	// 10. cancel_run
	mcpServer.AddTool(
		mcp.NewTool("cancel_run",
			mcp.WithDescription("Cancel a running job"),
			mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID to cancel")),
		),
		s.handleCancelRun,
	)
}

// ---------------------------------------------------------------------------
// Tool handlers
// ---------------------------------------------------------------------------

func (s *QuantMCPServer) handleListJobs(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobs, err := s.jobManager.ListJobs()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := make([]map[string]any, 0, len(jobs))
	for i := range jobs {
		result = append(result, jobToMap(&jobs[i]))
	}

	return marshalResult(result)
}

func (s *QuantMCPServer) handleGetJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	job, err := s.jobManager.GetJob(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(jobToMap(job))
}

func (s *QuantMCPServer) handleCreateJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	job := entity.Job{
		Name:             stringArg(args, "name"),
		Description:      stringArg(args, "description"),
		Type:             stringArg(args, "type"),
		WorkingDirectory: stringArg(args, "workingDirectory"),
		ScheduleEnabled:  boolArg(args, "scheduleEnabled"),
		ScheduleType:     stringArg(args, "scheduleType"),
		CronExpression:   stringArg(args, "cronExpression"),
		ScheduleInterval: intArg(args, "scheduleInterval"),
		TimeoutSeconds:   intArg(args, "timeoutSeconds"),
		Prompt:           stringArg(args, "prompt"),
		AllowBypass:      boolArg(args, "allowBypass"),
		AutonomousMode:   boolArg(args, "autonomousMode"),
		MaxRetries:       intArg(args, "maxRetries"),
		Model:            stringArg(args, "model"),
		ClaudeCommand:    stringArg(args, "claudeCommand"),
		Interpreter:      stringArg(args, "interpreter"),
		ScriptContent:    stringArg(args, "scriptContent"),
	}

	onSuccess := stringSliceArg(args, "onSuccess")
	onFailure := stringSliceArg(args, "onFailure")

	created, err := s.jobManager.CreateJob(job, onSuccess, onFailure)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(jobToMap(created))
}

func (s *QuantMCPServer) handleUpdateJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	existing, err := s.jobManager.GetJob(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Merge provided fields onto existing job.
	if v, ok := args["name"]; ok {
		existing.Name = v.(string)
	}
	if v, ok := args["description"]; ok {
		existing.Description = v.(string)
	}
	if v, ok := args["type"]; ok {
		existing.Type = v.(string)
	}
	if v, ok := args["workingDirectory"]; ok {
		existing.WorkingDirectory = v.(string)
	}
	if v, ok := args["scheduleEnabled"]; ok {
		existing.ScheduleEnabled, _ = v.(bool)
	}
	if v, ok := args["scheduleType"]; ok {
		existing.ScheduleType = v.(string)
	}
	if v, ok := args["cronExpression"]; ok {
		existing.CronExpression = v.(string)
	}
	if v, ok := args["scheduleInterval"]; ok {
		existing.ScheduleInterval = toInt(v)
	}
	if v, ok := args["timeoutSeconds"]; ok {
		existing.TimeoutSeconds = toInt(v)
	}
	if v, ok := args["prompt"]; ok {
		existing.Prompt = v.(string)
	}
	if v, ok := args["allowBypass"]; ok {
		existing.AllowBypass, _ = v.(bool)
	}
	if v, ok := args["autonomousMode"]; ok {
		existing.AutonomousMode, _ = v.(bool)
	}
	if v, ok := args["maxRetries"]; ok {
		existing.MaxRetries = toInt(v)
	}
	if v, ok := args["model"]; ok {
		existing.Model = v.(string)
	}
	if v, ok := args["claudeCommand"]; ok {
		existing.ClaudeCommand = v.(string)
	}
	if v, ok := args["interpreter"]; ok {
		existing.Interpreter = v.(string)
	}
	if v, ok := args["scriptContent"]; ok {
		existing.ScriptContent = v.(string)
	}

	onSuccess := stringSliceArg(args, "onSuccess")
	onFailure := stringSliceArg(args, "onFailure")

	updated, err := s.jobManager.UpdateJob(*existing, onSuccess, onFailure)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(jobToMap(updated))
}

func (s *QuantMCPServer) handleDeleteJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.jobManager.DeleteJob(id); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Job %s deleted successfully", id)), nil
}

func (s *QuantMCPServer) handleRunJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	run, err := s.jobManager.RunJob(id, "")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(runToMap(run))
}

func (s *QuantMCPServer) handleGetRun(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := requiredString(request, "runId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	run, err := s.jobManager.GetRun(runID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(runToMap(run))
}

func (s *QuantMCPServer) handleListRuns(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID, err := requiredString(request, "jobId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	runs, err := s.jobManager.ListRunsByJob(jobID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := make([]map[string]any, 0, len(runs))
	for i := range runs {
		result = append(result, runToMap(&runs[i]))
	}

	return marshalResult(result)
}

func (s *QuantMCPServer) handleGetRunOutput(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := requiredString(request, "runId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	output, err := s.jobManager.GetRunOutput(runID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(output), nil
}

func (s *QuantMCPServer) handleCancelRun(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := requiredString(request, "runId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.jobManager.CancelRun(runID); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Run %s cancelled successfully", runID)), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func jobToMap(job *entity.Job) map[string]any {
	if job == nil {
		return nil
	}
	m := map[string]any{
		"id":               job.ID,
		"name":             job.Name,
		"description":      job.Description,
		"type":             job.Type,
		"workingDirectory": job.WorkingDirectory,
		"scheduleEnabled":  job.ScheduleEnabled,
		"scheduleType":     job.ScheduleType,
		"cronExpression":   job.CronExpression,
		"scheduleInterval": job.ScheduleInterval,
		"timeoutSeconds":   job.TimeoutSeconds,
		"prompt":           job.Prompt,
		"allowBypass":      job.AllowBypass,
		"autonomousMode":   job.AutonomousMode,
		"maxRetries":       job.MaxRetries,
		"model":            job.Model,
		"claudeCommand":    job.ClaudeCommand,
		"interpreter":      job.Interpreter,
		"scriptContent":    job.ScriptContent,
		"createdAt":        job.CreatedAt,
		"updatedAt":        job.UpdatedAt,
	}
	if job.ScheduleStartTime != nil {
		m["scheduleStartTime"] = *job.ScheduleStartTime
	}
	if job.EnvVariables != nil {
		m["envVariables"] = job.EnvVariables
	}
	return m
}

func runToMap(run *entity.JobRun) map[string]any {
	if run == nil {
		return nil
	}
	m := map[string]any{
		"id":           run.ID,
		"jobId":        run.JobID,
		"status":       run.Status,
		"triggeredBy":  run.TriggeredBy,
		"sessionId":    run.SessionID,
		"durationMs":   run.DurationMs,
		"tokensUsed":   run.TokensUsed,
		"result":       run.Result,
		"errorMessage": run.ErrorMessage,
		"startedAt":    run.StartedAt,
	}
	if run.FinishedAt != nil {
		m["finishedAt"] = *run.FinishedAt
	}
	return m
}

func marshalResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to serialize result: %s", err.Error())), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func requiredString(request mcp.CallToolRequest, key string) (string, error) {
	args := request.GetArguments()
	v, ok := args[key]
	if !ok || v == nil {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return s, nil
}

func stringArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func boolArg(args map[string]any, key string) bool {
	v, ok := args[key]
	if !ok || v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

func intArg(args map[string]any, key string) int {
	v, ok := args[key]
	if !ok || v == nil {
		return 0
	}
	return toInt(v)
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func stringSliceArg(args map[string]any, key string) []string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
