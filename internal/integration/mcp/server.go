// Package mcp implements an MCP server that exposes Quant's job management
// to Claude Code and other MCP clients.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	appAdapter "quant/internal/application/adapter"
	"quant/internal/domain/entity"
)

// QuantMCPServer wraps an MCP server that exposes job and agent management tools.
type QuantMCPServer struct {
	jobManager   appAdapter.JobManager
	agentManager appAdapter.AgentManager
	httpServer   *http.Server
}

// NewQuantMCPServer creates a new MCP server with all job and agent management tools registered.
func NewQuantMCPServer(jobManager appAdapter.JobManager, agentManager appAdapter.AgentManager) *QuantMCPServer {
	mcpServer := server.NewMCPServer("quant", "1.0.0")

	s := &QuantMCPServer{jobManager: jobManager, agentManager: agentManager}

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

// Stop gracefully shuts down the HTTP server with a 2-second timeout.
func (s *QuantMCPServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func (s *QuantMCPServer) registerTools(mcpServer *server.MCPServer) {
	// 1. list_jobs
	mcpServer.AddTool(
		mcp.NewTool("list_jobs",
			mcp.WithDescription(`List all configured jobs with their full configuration.

Returns an array of job objects. Each job may have an agentId linking to an agent persona. When a job runs, the agent's configuration (role, goal, boundaries, skills) is injected as a system prompt into the Claude CLI session. The agent also provides env vars and a fallback model. Use get_agent(agentId) to see the full agent config.

Key fields in the response:
- id: UUID — use this to reference the job in run_job, update_job, delete_job, list_runs
- name: human-readable name shown in the canvas UI
- type: 'claude' (Claude CLI session) or 'bash' (shell script)
- agentId: UUID of the assigned agent (empty = no agent). The agent defines WHO executes the job (persona, rules, skills). Use get_agent to see details
- agentName/agentRole: inline summary of the assigned agent (empty if no agent)
- claudeCommand: which Claude CLI binary/alias to invoke (e.g. 'claude', 'claude-bl')
- prompt: the task instructions sent to Claude (what to do)
- successPrompt/failurePrompt: evaluation criteria run after the main task to determine success/failure
- metadataPrompt: instructions for extracting structured data passed to triggered downstream jobs
- scheduleEnabled: whether the job runs on a schedule (cron or interval)
- onSuccess/onFailure: trigger chains — downstream job IDs that fire when this job completes
- envVariables: key-value env vars injected at runtime (secrets, tokens, config)
- overrideRepoCommand: custom repo command override for the Claude CLI (advanced)
- workingDirectory: the directory where the job executes (supports ~/path)

Use this as the starting point to discover job IDs before calling run_job, update_job, or list_runs.`),
		),
		s.handleListJobs,
	)

	// 2. get_job
	mcpServer.AddTool(
		mcp.NewTool("get_job",
			mcp.WithDescription(`Get a job by ID with full configuration. Returns a single job object with all fields.

If the job has an agentId, the response includes agentName and agentRole inline for quick reference. Use get_agent(agentId) for the full agent config (boundaries, skills, env vars, MCP servers).

Use this to inspect a job's prompt, schedule, triggers, and agent assignment before running or modifying it.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID (get from list_jobs)")),
		),
		s.handleGetJob,
	)

	// 3. create_job
	mcpServer.AddTool(
		mcp.NewTool("create_job",
			mcp.WithDescription(`Create a new automated job in Quant. Jobs can be:
- 'claude' type: runs a Claude CLI session with a prompt (use for code reviews, analysis, complex tasks)
- 'bash' type: runs a shell script (use for health checks, deployments, notifications, data pipelines)

Jobs run autonomously with permissions bypassed by default. After creating, use update_job to wire trigger chains (onSuccess/onFailure arrays with target job IDs).

Trigger chains: when a job finishes, it can trigger other jobs based on the outcome. Use onSuccess/onFailure to build pipelines like: health-check → deploy (on success) → notify (on deploy success), health-check → incident-report (on failure).

For claude jobs, after execution a second evaluation prompt runs to determine success/failure and extract structured metadata that gets passed to triggered jobs, saving tokens vs passing raw output.

IMPORTANT: timeoutSeconds must be at least 60. For claude jobs, use 600 (10 min) as the default — they need time for the main task plus a follow-up evaluation call. 5 min is tight, 10 min is safe. For bash jobs, 60-120s is usually enough.

The Quant canvas UI auto-refreshes every 10 seconds and auto-layouts new jobs that don't have positions yet. After creating multiple jobs, they will appear organized on the canvas automatically.

Workflow for building pipelines:
1. Create all jobs first (they appear on canvas automatically)
2. Wire triggers with update_job onSuccess/onFailure
3. Use run_job to test the entry point — downstream jobs fire automatically

Returns the created job object with the generated ID. Use this ID for run_job, update_job, list_runs, etc.`),
			mcp.WithString("name", mcp.Required(), mcp.Description("Unique job name (e.g. health-check, deploy-staging, code-review-bot). Shown on canvas nodes")),
			mcp.WithString("description", mcp.Description("What the job does — shown in the canvas UI tooltip and job details")),
			mcp.WithString("type", mcp.Required(), mcp.Description("'claude' for Claude CLI sessions, 'bash' for shell scripts")),
			mcp.WithString("workingDirectory", mcp.Description("Working directory (supports ~/path). Leave empty for home dir. This is where Claude or the script runs")),
			// Schedule
			mcp.WithBoolean("scheduleEnabled", mcp.Description("Enable scheduled execution. False = manual/trigger only. Default: false")),
			mcp.WithString("scheduleType", mcp.Description("'recurring' (repeats on interval/cron) or 'one_time' (runs once then auto-disables). Default: 'recurring'")),
			mcp.WithString("cronExpression", mcp.Description("Cron expression (e.g. '0 9 * * 1-5' for weekdays 9am, '*/30 * * * *' for every 30 min). Alternative to scheduleInterval. Standard 5-field cron format")),
			mcp.WithNumber("scheduleInterval", mcp.Description("Repeat interval in minutes (e.g. 30 for every 30min). Alternative to cronExpression. Simpler but less flexible")),
			mcp.WithNumber("timeoutSeconds", mcp.Description("Max execution time in seconds. Claude jobs: use 600 (10 min, safe default). Bash jobs: 60-120s. Minimum 60. Job process is killed after this")),
			// Claude config
			mcp.WithString("prompt", mcp.Description("Main task prompt for claude jobs. Be specific about what to do, which files to read, which tools to use, and what output format to produce. This is piped to Claude via stdin with -p flag")),
			mcp.WithNumber("maxRetries", mcp.Description("Retry count on failure (claude only). Each retry includes the previous attempt's output as context so Claude can learn from errors. Default: 0")),
			mcp.WithString("model", mcp.Description("Claude model (e.g. 'claude-sonnet-4-6', 'claude-opus-4-6', 'claude-haiku-4-5-20251001'). Empty = CLI default. Agent model is used as fallback if set")),
			mcp.WithString("claudeCommand", mcp.Description("Claude CLI command/alias (e.g. 'claude', 'claude-bl'). Supports shell aliases from ~/.zshrc. Default: 'claude'")),
			mcp.WithString("agentId", mcp.Description("Agent ID to use for this job. The agent's role, goal, boundaries, skills, MCP servers, and env vars are injected as a system prompt. Use list_agents to get IDs")),
			mcp.WithString("overrideRepoCommand", mcp.Description("Custom repo command override for Claude CLI (advanced). Overrides the default command used to interact with the repository")),
			mcp.WithString("successPrompt", mcp.Description("How to evaluate success after the main task completes (max 300 chars). E.g. 'All tests passed and PR was created'. Claude runs a second evaluation call using this. Optional — if omitted, exit code determines success")),
			mcp.WithString("failurePrompt", mcp.Description("How to evaluate failure after the main task completes (max 300 chars). E.g. 'Tests failed, build errors, or no PR created'. Used in the evaluation call. Optional")),
			mcp.WithString("metadataPrompt", mcp.Description("What structured data to extract for triggered downstream jobs (max 500 chars). E.g. 'Extract PR URL, test count, error summary as JSON'. This metadata is passed as context to downstream triggered jobs, saving tokens vs passing raw output. Optional")),
			// Bash config
			mcp.WithString("interpreter", mcp.Description("Shell interpreter for bash jobs: '/bin/bash' (default), '/bin/zsh', 'python3', 'node', etc. The scriptContent is piped to this via stdin")),
			mcp.WithString("scriptContent", mcp.Description("Script content for bash jobs. Piped to the interpreter via stdin. Exit 0 = success (fires onSuccess triggers), non-zero = failure (fires onFailure triggers). Stdout/stderr are captured as run output")),
			// Environment
			mcp.WithString("envVariables", mcp.Description("JSON object of environment variables injected at runtime. E.g. '{\"API_KEY\":\"xxx\",\"ENV\":\"prod\"}'. For claude jobs, these are set before the CLI runs. For bash jobs, available in the script. Agent env vars are merged (job takes precedence)")),
			// Triggers
			mcp.WithString("onSuccess", mcp.Description("JSON array of job IDs to trigger on success. E.g. '[\"job-id-1\",\"job-id-2\"]'. All listed jobs run in parallel. Use list_jobs to get IDs")),
			mcp.WithString("onFailure", mcp.Description("JSON array of job IDs to trigger on failure. E.g. '[\"job-id-1\"]'. All listed jobs run in parallel. Use list_jobs to get IDs")),
			// Flags
			mcp.WithBoolean("allowBypass", mcp.Description("Allow --dangerously-skip-permissions flag for claude jobs. Default: true. Set to false to require manual permission grants during execution")),
			mcp.WithBoolean("autonomousMode", mcp.Description("Run in autonomous mode without stopping to ask the user. Default: true. Set to false for interactive jobs that need human approval")),
		),
		s.handleCreateJob,
	)

	// 4. update_job
	mcpServer.AddTool(
		mcp.NewTool("update_job",
			mcp.WithDescription(`Update a job's configuration. Only provided fields are changed — omitted fields keep their current values. Also use this to wire trigger chains by setting onSuccess/onFailure with arrays of target job IDs.

Returns the full updated job object.

Common workflows:
- Wire triggers: update_job(id, onSuccess='["target-job-id"]')
- Change prompt: update_job(id, prompt="new prompt")
- Enable schedule: update_job(id, scheduleEnabled=true, scheduleInterval=30)
- Add evaluation: update_job(id, successPrompt="...", failurePrompt="...")
- Assign agent: update_job(id, agentId="agent-uuid")
- Unassign agent: update_job(id, agentId="")
- Add env vars: update_job(id, envVariables='{"KEY":"value"}')
- Disable bypass: update_job(id, allowBypass=false)`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID to update (get from list_jobs)")),
			mcp.WithString("name", mcp.Description("Job name")),
			mcp.WithString("description", mcp.Description("Job description")),
			mcp.WithString("type", mcp.Description("'claude' or 'bash'")),
			mcp.WithString("workingDirectory", mcp.Description("Working directory (supports ~/path)")),
			mcp.WithBoolean("scheduleEnabled", mcp.Description("Enable/disable scheduled execution")),
			mcp.WithString("scheduleType", mcp.Description("'recurring' or 'one_time'")),
			mcp.WithString("cronExpression", mcp.Description("Cron expression (5-field format)")),
			mcp.WithNumber("scheduleInterval", mcp.Description("Interval in minutes")),
			mcp.WithNumber("timeoutSeconds", mcp.Description("Timeout in seconds (min 60)")),
			mcp.WithString("prompt", mcp.Description("Task prompt for claude jobs")),
			mcp.WithNumber("maxRetries", mcp.Description("Retry count on failure (claude jobs)")),
			mcp.WithString("model", mcp.Description("Claude model (e.g. 'claude-sonnet-4-6')")),
			mcp.WithString("claudeCommand", mcp.Description("Claude CLI command/alias")),
			mcp.WithString("agentId", mcp.Description("Agent ID. Use list_agents to get IDs. Set to empty string to unassign")),
			mcp.WithString("overrideRepoCommand", mcp.Description("Custom repo command override (advanced)")),
			mcp.WithString("successPrompt", mcp.Description("Success evaluation criteria (max 300 chars)")),
			mcp.WithString("failurePrompt", mcp.Description("Failure evaluation criteria (max 300 chars)")),
			mcp.WithString("metadataPrompt", mcp.Description("Metadata extraction instructions (max 500 chars)")),
			mcp.WithString("interpreter", mcp.Description("Script interpreter for bash jobs")),
			mcp.WithString("scriptContent", mcp.Description("Script content for bash jobs")),
			mcp.WithString("envVariables", mcp.Description("JSON object of env vars. E.g. '{\"KEY\":\"value\"}'. Replaces existing env vars")),
			mcp.WithString("onSuccess", mcp.Description("JSON array of job IDs to trigger on success. E.g. '[\"id1\",\"id2\"]'. Replaces existing triggers")),
			mcp.WithString("onFailure", mcp.Description("JSON array of job IDs to trigger on failure. E.g. '[\"id1\"]'. Replaces existing triggers")),
			mcp.WithBoolean("allowBypass", mcp.Description("Allow --dangerously-skip-permissions for claude jobs")),
			mcp.WithBoolean("autonomousMode", mcp.Description("Run without stopping to ask the user")),
		),
		s.handleUpdateJob,
	)

	// 5. delete_job
	mcpServer.AddTool(
		mcp.NewTool("delete_job",
			mcp.WithDescription(`Delete a job and all its trigger chains and run history. This is irreversible.

Deleting a job also removes:
- All trigger connections TO and FROM this job (other jobs' onSuccess/onFailure entries referencing this job are cleaned up)
- All run records and their output logs
- The job's position on the canvas

Returns a confirmation message. Use list_jobs to verify deletion.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID to delete (get from list_jobs)")),
		),
		s.handleDeleteJob,
	)

	// 6. run_job
	mcpServer.AddTool(
		mcp.NewTool("run_job",
			mcp.WithDescription(`Trigger a job to run immediately. Returns the run object with a run ID and initial status 'pending'.

The job executes asynchronously in a background goroutine:
1. Status transitions: pending → running → success/failed/timed_out
2. If the job has onSuccess/onFailure trigger chains, downstream jobs fire automatically when this run completes
3. For claude jobs with maxRetries > 0, failed runs are automatically retried with the previous output as context

To monitor execution:
- get_run(runId) — check status, duration, tokens used
- get_run_output(runId) — get the live output (updates while running, polled every few seconds)
- get_pipeline_status(runId) — trace the full cascade of triggered downstream jobs

The Quant canvas UI shows running jobs with a pulsing green border and highlights the active pipeline path in real-time.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Job ID to run (get from list_jobs)")),
		),
		s.handleRunJob,
	)

	// 7. get_run
	mcpServer.AddTool(
		mcp.NewTool("get_run",
			mcp.WithDescription(`Get details of a specific job run.

Returns a run object with:
- id: unique run ID
- jobId: which job this run belongs to
- status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled' | 'timed_out'
- triggeredBy: run ID of the upstream job that triggered this run (empty if manual)
- sessionId: Claude session ID (for claude-type jobs — can be used to view the session)
- modelUsed: which Claude model actually executed (extracted from stream output)
- durationMs: execution time in milliseconds
- tokensUsed: total tokens consumed (input + output, claude jobs only)
- result: the evaluation result text (if successPrompt/failurePrompt configured)
- errorMessage: error details if the run failed
- startedAt: ISO timestamp when execution began
- finishedAt: ISO timestamp when execution completed (null while running)

Use after run_job to check if execution completed, or to inspect historical run details.`),
			mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID (returned by run_job or list_runs)")),
		),
		s.handleGetRun,
	)

	// 8. list_runs
	mcpServer.AddTool(
		mcp.NewTool("list_runs",
			mcp.WithDescription(`List all runs for a job, sorted by most recent first. Returns an array of run objects.

Each run includes: id, status, triggeredBy, sessionId, modelUsed, durationMs, tokensUsed, result, errorMessage, startedAt, finishedAt.

Use to:
- Check job execution history and recent status
- Find specific run IDs for get_run_output or get_pipeline_status
- See which runs were triggered by upstream jobs (triggeredBy field contains the parent run ID)
- Monitor how many tokens jobs are consuming over time`),
			mcp.WithString("jobId", mcp.Required(), mcp.Description("Job ID (get from list_jobs)")),
		),
		s.handleListRuns,
	)

	// 9. get_run_output
	mcpServer.AddTool(
		mcp.NewTool("get_run_output",
			mcp.WithDescription(`Get the full output/logs of a job run. Returns raw text.

For claude jobs: the full Claude CLI session output including all tool calls, reasoning, and final response. While running, output updates incrementally (poll to see progress).

For bash jobs: combined stdout/stderr output from the script execution.

Also includes (appended at the end):
- Evaluation results: if successPrompt/failurePrompt were configured, the evaluation outcome
- Extracted metadata: if metadataPrompt was configured, the structured data that was passed to downstream triggered jobs

This can return large amounts of text for long-running jobs. Use get_run first to check status before fetching output.`),
			mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID (from list_runs or run_job)")),
		),
		s.handleGetRunOutput,
	)

	// 10. cancel_run
	mcpServer.AddTool(
		mcp.NewTool("cancel_run",
			mcp.WithDescription(`Cancel a currently running job. Kills the process immediately (SIGKILL).

Effects:
- Run status is set to 'cancelled'
- No onSuccess/onFailure triggers are fired (the pipeline stops here)
- Duration is recorded up to the cancellation point
- Any partial output is preserved and accessible via get_run_output

Only works on runs with status 'running' or 'pending'. Returns a confirmation message.`),
			mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID to cancel (from list_runs or run_job)")),
		),
		s.handleCancelRun,
	)

	// 11. get_triggers
	mcpServer.AddTool(
		mcp.NewTool("get_triggers",
			mcp.WithDescription(`Get the full trigger graph showing how all jobs are connected. Returns an array of trigger info objects.

Each object contains:
- job_id: the job's UUID
- job_name: human-readable name
- on_success: array of job names this job triggers on success
- on_failure: array of job names this job triggers on failure
- triggered_by: array of job names that can trigger this job

Only includes jobs that have at least one trigger connection. Use this to:
- Understand the pipeline topology before wiring new connections
- Verify trigger chains are correctly configured after update_job
- Identify entry points (jobs with no triggered_by) and terminal nodes (jobs with no on_success/on_failure)
- Debug why a downstream job didn't fire (check the trigger graph)`),
		),
		s.handleGetTriggers,
	)

	// -----------------------------------------------------------------------
	// Agent tools
	// -----------------------------------------------------------------------

	// 13. list_agents
	mcpServer.AddTool(
		mcp.NewTool("list_agents",
			mcp.WithDescription(`List all configured agents. Returns an array of agent objects with full configuration.

Agents define the persona and constraints for Claude jobs:
- identity: name, color (for UI), role (who), goal (what to achieve)
- access: which MCP servers and env vars the agent can use
- boundaries: anti-prompt rules the agent must never violate
- skills: which Claude skills (from ~/.claude/skills/) are enabled
- model: fallback Claude model when the job doesn't specify one

Assign an agent to a job with update_job(id, agentId="agent-uuid") to give the job a persona. Use this to find agent IDs.`),
		),
		s.handleListAgents,
	)

	// 14. get_agent
	mcpServer.AddTool(
		mcp.NewTool("get_agent",
			mcp.WithDescription(`Get an agent by ID. Returns full configuration including:
- id, name, color: identity and UI representation
- role: who the agent is (identity, tone, expertise)
- goal: what the agent should achieve (success criteria)
- model: Claude model fallback
- autonomousMode: whether the agent runs without stopping to ask
- boundaries: array of hard rules the agent must never violate
- skills: map of skill name → enabled (from ~/.claude/skills/)
- mcpServers: map of MCP server name → enabled
- envVariables: map of env var name → value (secrets, tokens)
- createdAt, updatedAt: timestamps

Use get_agent_system_prompt(id) to see the actual system prompt that gets injected.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Agent ID (get from list_agents)")),
		),
		s.handleGetAgent,
	)

	// 15. create_agent
	mcpServer.AddTool(
		mcp.NewTool("create_agent",
			mcp.WithDescription(`Create a new agent persona for Claude jobs. Agents are task-specific — create many small focused agents, not monoliths.

An agent's config is injected as a system prompt when a job runs. The system prompt includes:
- role: who the agent is (identity, tone, expertise) — max 500 chars, be semantically dense
- goal: success criteria — max 500 chars
- boundaries: hard rules the agent must never violate (e.g. "never push to main", "never modify production databases")
- skills: which Claude skills the agent can use (from ~/.claude/skills/) — these provide domain-specific knowledge and patterns
- mcpServers: which MCP servers the agent can access (e.g. database, Linear, GitHub)
- envVariables: private secrets only this agent knows (e.g. API tokens, database URLs)
- autonomousMode: true (default) = agent executes without stopping to ask

After creating, assign to a job with update_job(id, agentId="agent-uuid").

Design tips:
- One agent per role: "code_reviewer", "devops_engineer", "data_analyst" — not "do_everything"
- Role should describe expertise and communication style: "Senior Go engineer focused on clean architecture. Direct, concise."
- Goal should be measurable: "Review PR for architectural violations and security issues. Report findings in markdown."
- Boundaries are hard stops, not suggestions: "never push to main", "never run DROP statements"
- Skills provide patterns the agent follows: architecture rules, testing conventions, etc.

Returns the created agent object with generated ID.`),
			mcp.WithString("name", mcp.Required(), mcp.Description("Agent name (e.g. 'code_reviewer', 'devops_engineer', 'data_analyst'). Used in job canvas UI")),
			mcp.WithString("color", mcp.Description("Hex color for UI (e.g. '#10B981' green, '#3B82F6' blue, '#EF4444' red). Default: green")),
			mcp.WithString("role", mcp.Description("Who is this agent? Identity, expertise, and tone. Max 500 chars. Be semantically dense. E.g. 'Senior Go engineer focused on clean architecture and DDD. Direct, concise, prefers code over explanations.'")),
			mcp.WithString("goal", mcp.Description("What does this agent achieve? Measurable success criteria. Max 500 chars. E.g. 'Review changes for architectural violations, security issues, and test coverage. Report findings in structured markdown.'")),
			mcp.WithString("model", mcp.Description("Claude model (e.g. 'claude-opus-4-6', 'claude-sonnet-4-6', 'claude-haiku-4-5-20251001'). Used as fallback when the job doesn't specify a model")),
			mcp.WithBoolean("autonomousMode", mcp.Description("Execute without stopping to ask the user. Default: true. Set false for agents that should pause for human approval")),
			mcp.WithString("boundaries", mcp.Description("JSON array of anti-prompt rules. Hard constraints the agent must never violate. E.g. '[\"never push to main\",\"never delete databases\",\"never modify files outside src/\"]'")),
			mcp.WithString("skills", mcp.Description("JSON object of skill toggles. Skills from ~/.claude/skills/ provide domain knowledge. E.g. '{\"architecture\":true,\"bdd-testing\":true}'. Use list_available_skills to see what's available")),
			mcp.WithString("mcpServers", mcp.Description("JSON object of MCP server toggles. Controls which external tools the agent can access. E.g. '{\"dbhub\":true,\"linear\":true,\"figma\":false}'. Use list_available_mcp_servers to see what's configured")),
			mcp.WithString("envVariables", mcp.Description("JSON object of private env vars. Secrets injected into the agent's environment at runtime. E.g. '{\"GITHUB_TOKEN\":\"ghp_xxx\",\"DATABASE_URL\":\"postgres://...\"}'. Only this agent sees these values")),
		),
		s.handleCreateAgent,
	)

	// 16. update_agent
	mcpServer.AddTool(
		mcp.NewTool("update_agent",
			mcp.WithDescription(`Update an agent's configuration. Only provided fields are changed — omitted fields keep their current values.

Note: for map/array fields (boundaries, skills, mcpServers, envVariables), the provided value REPLACES the entire field. To add a single boundary, include all existing ones plus the new one.

Returns the full updated agent object.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Agent ID to update (get from list_agents)")),
			mcp.WithString("name", mcp.Description("Agent name")),
			mcp.WithString("color", mcp.Description("Hex color for UI (e.g. '#10B981')")),
			mcp.WithString("role", mcp.Description("Role description — who the agent is (max 500 chars)")),
			mcp.WithString("goal", mcp.Description("Goal description — success criteria (max 500 chars)")),
			mcp.WithString("model", mcp.Description("Claude model fallback")),
			mcp.WithBoolean("autonomousMode", mcp.Description("Execute without stopping to ask")),
			mcp.WithString("boundaries", mcp.Description("JSON array of anti-prompt rules. REPLACES all existing boundaries")),
			mcp.WithString("skills", mcp.Description("JSON object of skill toggles. REPLACES all existing skills")),
			mcp.WithString("mcpServers", mcp.Description("JSON object of MCP server toggles. REPLACES all existing MCP servers")),
			mcp.WithString("envVariables", mcp.Description("JSON object of env vars. REPLACES all existing env vars")),
		),
		s.handleUpdateAgent,
	)

	// 17. delete_agent
	mcpServer.AddTool(
		mcp.NewTool("delete_agent",
			mcp.WithDescription(`Delete an agent permanently. This is irreversible.

Side effects:
- Jobs using this agent will have their agentId cleared (they become agent-less)
- The agent's system prompt will no longer be injected into those jobs
- Existing run history is preserved (historical runs still reference the agent ID)

Returns a confirmation message.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Agent ID to delete (get from list_agents)")),
		),
		s.handleDeleteAgent,
	)

	// 18. list_available_skills
	mcpServer.AddTool(
		mcp.NewTool("list_available_skills",
			mcp.WithDescription(`List all Claude skills available in ~/.claude/skills/. Returns an array of skill name strings.

Skills are markdown files or directories that provide domain-specific knowledge and patterns to Claude. Examples:
- architecture: clean architecture patterns and layer separation rules
- bdd-testing: BDD/Gherkin test writing conventions
- code-review: code review checklist and standards

Use these names as keys in the agent 'skills' parameter. E.g. create_agent(skills='{"architecture":true,"bdd-testing":true}').

Skills are read from the filesystem at agent creation time. The agent's system prompt includes the content of all enabled skills.`),
		),
		s.handleListAvailableSkills,
	)

	// 19. list_available_mcp_servers
	mcpServer.AddTool(
		mcp.NewTool("list_available_mcp_servers",
			mcp.WithDescription(`List all MCP servers configured in ~/.mcp.json. Returns an array of server name strings.

MCP servers provide external tool access to agents. Common examples:
- dbhub: database querying (SQL)
- linear: project management (issues, tasks)
- figma: design file access
- context7: documentation lookup

Use these names as keys in the agent 'mcpServers' parameter. E.g. create_agent(mcpServers='{"dbhub":true,"linear":true}').

When an agent has MCP servers enabled, the Claude CLI session is started with access to those servers' tools.`),
		),
		s.handleListAvailableMcpServers,
	)

	// 20. get_agent_system_prompt
	mcpServer.AddTool(
		mcp.NewTool("get_agent_system_prompt",
			mcp.WithDescription(`Preview the exact system prompt that would be injected for a given agent when a job runs.

The system prompt is constructed from the agent's configuration:
- Role and goal are included as identity context
- Boundaries are listed as hard rules
- Enabled skills have their full markdown content injected
- MCP server access is documented

Use this to debug agent behavior:
- Verify the system prompt reads correctly before running a job
- Check that skills are being included properly
- Ensure boundaries are clear and unambiguous
- Test prompt changes after update_agent

Returns the full system prompt text, or a message indicating the prompt is empty.`),
			mcp.WithString("id", mcp.Required(), mcp.Description("Agent ID (get from list_agents)")),
		),
		s.handleGetAgentSystemPrompt,
	)

	// 12. get_pipeline_status
	mcpServer.AddTool(
		mcp.NewTool("get_pipeline_status",
			mcp.WithDescription(`Given a run ID, trace the full chain of triggered runs downstream using BFS. Shows the complete cascade of the pipeline execution.

Returns an array of pipeline step objects, each with:
- run_id: the run's UUID
- job_name: human-readable name of the job
- status: pending/running/success/failed/cancelled/timed_out
- duration_ms: execution time in milliseconds
- tokens_used: total tokens consumed (claude jobs only)
- triggered_by_run: the upstream run ID that triggered this step
- error: error message if the step failed (omitted if empty)

Use after run_job to see the full pipeline execution result without manually checking each job. The first element is always the initial run, followed by downstream triggered runs in BFS order.

Example flow: run_job("health-check") → get_pipeline_status(runId) shows:
1. health-check (success, 45s, 12k tokens)
2. deploy-staging (success, 120s, 8k tokens, triggered by health-check)
3. notify-slack (success, 5s, 0 tokens, triggered by deploy-staging)`),
			mcp.WithString("runId", mcp.Required(), mcp.Description("The initial run ID to trace from (returned by run_job or list_runs)")),
		),
		s.handleGetPipelineStatus,
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
		m := jobToMap(&jobs[i])
		s.enrichJobWithAgent(m, jobs[i].AgentID)
		result = append(result, m)
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

	m := jobToMap(job)
	s.enrichJobWithAgent(m, job.AgentID)
	return marshalResult(m)
}

func (s *QuantMCPServer) handleCreateJob(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	job := entity.Job{
		Name:                stringArg(args, "name"),
		Description:         stringArg(args, "description"),
		Type:                stringArg(args, "type"),
		WorkingDirectory:    stringArg(args, "workingDirectory"),
		ScheduleEnabled:     boolArg(args, "scheduleEnabled"),
		ScheduleType:        stringArg(args, "scheduleType"),
		CronExpression:      stringArg(args, "cronExpression"),
		ScheduleInterval:    intArg(args, "scheduleInterval"),
		TimeoutSeconds:      intArg(args, "timeoutSeconds"),
		Prompt:              stringArg(args, "prompt"),
		AllowBypass:         true,
		AutonomousMode:      true,
		MaxRetries:          intArg(args, "maxRetries"),
		Model:               stringArg(args, "model"),
		ClaudeCommand:       stringArg(args, "claudeCommand"),
		AgentID:             stringArg(args, "agentId"),
		OverrideRepoCommand: stringArg(args, "overrideRepoCommand"),
		SuccessPrompt:       stringArg(args, "successPrompt"),
		FailurePrompt:       stringArg(args, "failurePrompt"),
		MetadataPrompt:      stringArg(args, "metadataPrompt"),
		Interpreter:         stringArg(args, "interpreter"),
		ScriptContent:       stringArg(args, "scriptContent"),
		EnvVariables:        mapStringArg(args, "envVariables"),
	}

	// Allow explicit override of flags (default to true)
	if v, ok := args["allowBypass"]; ok {
		job.AllowBypass, _ = v.(bool)
	}
	if v, ok := args["autonomousMode"]; ok {
		job.AutonomousMode, _ = v.(bool)
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
	if v, ok := args["agentId"]; ok {
		existing.AgentID, _ = v.(string)
	}
	if v, ok := args["successPrompt"]; ok {
		existing.SuccessPrompt = v.(string)
	}
	if v, ok := args["failurePrompt"]; ok {
		existing.FailurePrompt = v.(string)
	}
	if v, ok := args["metadataPrompt"]; ok {
		existing.MetadataPrompt = v.(string)
	}
	if v, ok := args["interpreter"]; ok {
		existing.Interpreter = v.(string)
	}
	if v, ok := args["scriptContent"]; ok {
		existing.ScriptContent = v.(string)
	}
	if v, ok := args["overrideRepoCommand"]; ok {
		existing.OverrideRepoCommand, _ = v.(string)
	}
	if _, ok := args["envVariables"]; ok {
		existing.EnvVariables = mapStringArg(args, "envVariables")
	}

	// Only update triggers if explicitly provided — nil means "don't change"
	onSuccess := stringSliceArg(args, "onSuccess")
	onFailure := stringSliceArg(args, "onFailure")

	// If neither provided, preserve existing triggers
	if onSuccess == nil && onFailure == nil {
		existingSuccess, existingFailure, _, _ := s.jobManager.GetTriggersForJob(id)
		onSuccess = make([]string, len(existingSuccess))
		for i, t := range existingSuccess {
			onSuccess[i] = t.TargetJobID
		}
		onFailure = make([]string, len(existingFailure))
		for i, t := range existingFailure {
			onFailure[i] = t.TargetJobID
		}
	}

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

func (s *QuantMCPServer) handleGetTriggers(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobs, err := s.jobManager.ListJobs()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	type triggerInfo struct {
		JobID       string   `json:"job_id"`
		JobName     string   `json:"job_name"`
		OnSuccess   []string `json:"on_success"`
		OnFailure   []string `json:"on_failure"`
		TriggeredBy []string `json:"triggered_by"`
	}

	// Build name lookup
	nameMap := make(map[string]string)
	for _, j := range jobs {
		nameMap[j.ID] = j.Name
	}

	var result []triggerInfo
	for _, j := range jobs {
		onSuccess, onFailure, triggeredBy, err := s.jobManager.GetTriggersForJob(j.ID)
		if err != nil {
			continue
		}

		info := triggerInfo{
			JobID:   j.ID,
			JobName: j.Name,
		}
		for _, t := range onSuccess {
			name := nameMap[t.TargetJobID]
			if name == "" {
				name = t.TargetJobID
			}
			info.OnSuccess = append(info.OnSuccess, name)
		}
		for _, t := range onFailure {
			name := nameMap[t.TargetJobID]
			if name == "" {
				name = t.TargetJobID
			}
			info.OnFailure = append(info.OnFailure, name)
		}
		for _, t := range triggeredBy {
			name := nameMap[t.SourceJobID]
			if name == "" {
				name = t.SourceJobID
			}
			info.TriggeredBy = append(info.TriggeredBy, name)
		}

		// Only include jobs that have any trigger connections
		if len(info.OnSuccess) > 0 || len(info.OnFailure) > 0 || len(info.TriggeredBy) > 0 {
			result = append(result, info)
		}
	}

	return marshalResult(result)
}

func (s *QuantMCPServer) handleGetPipelineStatus(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, err := requiredString(request, "runId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Build name lookup
	jobs, _ := s.jobManager.ListJobs()
	nameMap := make(map[string]string)
	for _, j := range jobs {
		nameMap[j.ID] = j.Name
	}

	type pipelineStep struct {
		RunID       string `json:"run_id"`
		JobName     string `json:"job_name"`
		Status      string `json:"status"`
		DurationMs  int64  `json:"duration_ms"`
		TokensUsed  int    `json:"tokens_used"`
		TriggeredBy string `json:"triggered_by_run"`
		Error       string `json:"error,omitempty"`
	}

	var steps []pipelineStep
	visited := make(map[string]bool)

	// BFS from the initial run, following triggered_by references
	queue := []string{runID}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		run, err := s.jobManager.GetRun(currentID)
		if err != nil {
			continue
		}

		step := pipelineStep{
			RunID:       run.ID,
			JobName:     nameMap[run.JobID],
			Status:      run.Status,
			DurationMs:  run.DurationMs,
			TokensUsed:  run.TokensUsed,
			TriggeredBy: run.TriggeredBy,
			Error:       run.ErrorMessage,
		}
		if step.JobName == "" {
			step.JobName = run.JobID
		}
		steps = append(steps, step)

		// Find runs that were triggered by this run
		for _, j := range jobs {
			runs, err := s.jobManager.ListRunsByJob(j.ID)
			if err != nil {
				continue
			}
			for _, r := range runs {
				if r.TriggeredBy == currentID && !visited[r.ID] {
					queue = append(queue, r.ID)
				}
			}
		}
	}

	return marshalResult(steps)
}

// ---------------------------------------------------------------------------
// Agent handlers
// ---------------------------------------------------------------------------

func (s *QuantMCPServer) handleListAgents(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agents, err := s.agentManager.ListAgents()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := make([]map[string]any, 0, len(agents))
	for i := range agents {
		result = append(result, agentToMap(&agents[i]))
	}

	return marshalResult(result)
}

func (s *QuantMCPServer) handleGetAgent(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	agent, err := s.agentManager.GetAgent(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if agent == nil {
		return mcp.NewToolResultError(fmt.Sprintf("agent not found: %s", id)), nil
	}

	return marshalResult(agentToMap(agent))
}

func (s *QuantMCPServer) handleCreateAgent(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	agent := entity.Agent{
		Name:           stringArg(args, "name"),
		Color:          stringArg(args, "color"),
		Role:           stringArg(args, "role"),
		Goal:           stringArg(args, "goal"),
		Model:          stringArg(args, "model"),
		AutonomousMode: true,
	}

	if v, ok := args["autonomousMode"]; ok {
		agent.AutonomousMode, _ = v.(bool)
	}

	agent.Boundaries = stringSliceArg(args, "boundaries")
	if agent.Boundaries == nil {
		agent.Boundaries = []string{}
	}

	agent.Skills = mapBoolArg(args, "skills")
	agent.McpServers = mapBoolArg(args, "mcpServers")
	agent.EnvVariables = mapStringArg(args, "envVariables")

	created, err := s.agentManager.CreateAgent(agent)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(agentToMap(created))
}

func (s *QuantMCPServer) handleUpdateAgent(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	existing, err := s.agentManager.GetAgent(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if existing == nil {
		return mcp.NewToolResultError(fmt.Sprintf("agent not found: %s", id)), nil
	}

	if v, ok := args["name"]; ok {
		existing.Name, _ = v.(string)
	}
	if v, ok := args["color"]; ok {
		existing.Color, _ = v.(string)
	}
	if v, ok := args["role"]; ok {
		existing.Role, _ = v.(string)
	}
	if v, ok := args["goal"]; ok {
		existing.Goal, _ = v.(string)
	}
	if v, ok := args["model"]; ok {
		existing.Model, _ = v.(string)
	}
	if v, ok := args["autonomousMode"]; ok {
		existing.AutonomousMode, _ = v.(bool)
	}
	if _, ok := args["boundaries"]; ok {
		existing.Boundaries = stringSliceArg(args, "boundaries")
		if existing.Boundaries == nil {
			existing.Boundaries = []string{}
		}
	}
	if _, ok := args["skills"]; ok {
		existing.Skills = mapBoolArg(args, "skills")
	}
	if _, ok := args["mcpServers"]; ok {
		existing.McpServers = mapBoolArg(args, "mcpServers")
	}
	if _, ok := args["envVariables"]; ok {
		existing.EnvVariables = mapStringArg(args, "envVariables")
	}

	updated, err := s.agentManager.UpdateAgent(*existing)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return marshalResult(agentToMap(updated))
}

func (s *QuantMCPServer) handleDeleteAgent(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.agentManager.DeleteAgent(id); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Agent %s deleted successfully", id)), nil
}

func (s *QuantMCPServer) handleListAvailableSkills(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	skillsDir := filepath.Join(home, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return marshalResult([]string{})
	}

	var skills []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			skillFile := filepath.Join(skillsDir, name, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				skills = append(skills, name)
			}
			continue
		}
		if strings.HasSuffix(name, ".md") {
			skills = append(skills, strings.TrimSuffix(name, ".md"))
		}
	}

	return marshalResult(skills)
}

func (s *QuantMCPServer) handleListAvailableMcpServers(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mcpPath := filepath.Join(home, ".mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		return marshalResult([]string{})
	}

	var config map[string]interface{}
	if json.Unmarshal(data, &config) != nil {
		return marshalResult([]string{})
	}

	servers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return marshalResult([]string{})
	}

	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}

	return marshalResult(names)
}

func (s *QuantMCPServer) handleGetAgentSystemPrompt(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := requiredString(request, "id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	prompt, err := s.agentManager.BuildSystemPrompt(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if prompt == "" {
		return mcp.NewToolResultText("(empty system prompt — agent has no role, goal, boundaries, or skills configured)"), nil
	}

	return mcp.NewToolResultText(prompt), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// enrichJobWithAgent adds agentName and agentRole to a job map when an agent is assigned.
func (s *QuantMCPServer) enrichJobWithAgent(m map[string]any, agentID string) {
	if agentID == "" || s.agentManager == nil {
		return
	}
	agent, err := s.agentManager.GetAgent(agentID)
	if err != nil || agent == nil {
		return
	}
	m["agentName"] = agent.Name
	m["agentRole"] = agent.Role
}

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
		"claudeCommand":        job.ClaudeCommand,
		"agentId":              job.AgentID,
		"overrideRepoCommand":  job.OverrideRepoCommand,
		"successPrompt":        job.SuccessPrompt,
		"failurePrompt":    job.FailurePrompt,
		"metadataPrompt":   job.MetadataPrompt,
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
		"modelUsed":    run.ModelUsed,
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

func agentToMap(agent *entity.Agent) map[string]any {
	if agent == nil {
		return nil
	}
	return map[string]any{
		"id":             agent.ID,
		"name":           agent.Name,
		"color":          agent.Color,
		"role":           agent.Role,
		"goal":           agent.Goal,
		"model":          agent.Model,
		"autonomousMode": agent.AutonomousMode,
		"mcpServers":     agent.McpServers,
		"envVariables":   agent.EnvVariables,
		"boundaries":     agent.Boundaries,
		"skills":         agent.Skills,
		"createdAt":      agent.CreatedAt,
		"updatedAt":      agent.UpdatedAt,
	}
}

func mapBoolArg(args map[string]any, key string) map[string]bool {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	// Native JSON object
	if m, ok := v.(map[string]any); ok {
		result := make(map[string]bool, len(m))
		for k, val := range m {
			result[k], _ = val.(bool)
		}
		return result
	}
	// JSON string
	if s, ok := v.(string); ok && s != "" {
		var result map[string]bool
		if json.Unmarshal([]byte(s), &result) == nil {
			return result
		}
	}
	return nil
}

func mapStringArg(args map[string]any, key string) map[string]string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	// Native JSON object
	if m, ok := v.(map[string]any); ok {
		result := make(map[string]string, len(m))
		for k, val := range m {
			result[k], _ = val.(string)
		}
		return result
	}
	// JSON string
	if s, ok := v.(string); ok && s != "" {
		var result map[string]string
		if json.Unmarshal([]byte(s), &result) == nil {
			return result
		}
	}
	return nil
}

func stringSliceArg(args map[string]any, key string) []string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}

	// Handle []any (from native JSON arrays)
	if arr, ok := v.([]any); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	// Handle string (JSON-encoded array from MCP string field)
	if s, ok := v.(string); ok && s != "" {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "[") {
			var result []string
			if err := json.Unmarshal([]byte(s), &result); err == nil {
				return result
			}
		}
	}

	return nil
}
