import { useCallback, useEffect, useState } from "react";
import type { Job, JobRun } from "../types";
import * as api from "../api";

type JobTab = "settings" | "history";
type RunTab = "session" | "result";
type FilterTab = "all" | "active" | "failed";

interface Props {
  jobs: Job[];
  onCreateJob: () => void;
  onEditJob: (job: Job) => void;
  onRefreshJobs: () => void;
}

const font = "'JetBrains Mono', monospace";

function relativeTime(dateStr: string): string {
  if (!dateStr) return "---";
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = now - then;
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function formatDuration(ms: number): string {
  if (!ms) return "---";
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainSec = seconds % 60;
  if (minutes < 60) return `${minutes}m ${remainSec}s`;
  const hours = Math.floor(minutes / 60);
  const remainMin = minutes % 60;
  return `${hours}h ${remainMin}m`;
}

function statusColor(status: string): string {
  switch (status) {
    case "success": return "#10B981";
    case "running": return "#3B82F6";
    case "pending": return "#F59E0B";
    case "failed": return "#EF4444";
    case "cancelled": return "#6B7280";
    case "timed_out": return "#EF4444";
    default: return "#6B7280";
  }
}

const pulseKeyframes = `
@keyframes job-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
`;

export function JobsView({ jobs, onCreateJob, onEditJob, onRefreshJobs }: Props) {
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const [selectedTab, setSelectedTab] = useState<JobTab>("settings");
  const [runs, setRuns] = useState<JobRun[]>([]);
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [selectedRunTab, setSelectedRunTab] = useState<RunTab>("session");
  const [runOutput, setRunOutput] = useState<string>("");
  const [filter, setFilter] = useState<FilterTab>("all");
  const [copied, setCopied] = useState(false);

  const selectedJob = jobs.find((j) => j.id === selectedJobId) ?? null;
  const selectedRun = runs.find((r) => r.id === selectedRunId) ?? null;

  const fetchRuns = useCallback(async (jobId: string) => {
    try {
      const result = await api.listRunsByJob(jobId);
      setRuns(result ?? []);
    } catch (err) {
      console.error("failed to fetch runs:", err);
      setRuns([]);
    }
  }, []);

  const fetchOutput = useCallback(async (runId: string) => {
    try {
      const output = await api.getRunOutput(runId);
      setRunOutput(output ?? "");
    } catch (err) {
      console.error("failed to fetch run output:", err);
      setRunOutput("");
    }
  }, []);

  useEffect(() => {
    if (selectedJobId) {
      fetchRuns(selectedJobId);
      setSelectedRunId(null);
      setRunOutput("");
    } else {
      setRuns([]);
    }
  }, [selectedJobId, fetchRuns]);

  useEffect(() => {
    if (selectedRunId && selectedRunTab === "session") {
      fetchOutput(selectedRunId);
    }
  }, [selectedRunId, selectedRunTab, fetchOutput]);

  // Poll for updates when a running run is selected
  useEffect(() => {
    if (!selectedRun || selectedRun.status !== "running") return;
    if (!selectedJobId) return;

    const interval = setInterval(async () => {
      await fetchRuns(selectedJobId);
      if (selectedRunId && selectedRunTab === "session") {
        await fetchOutput(selectedRunId);
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [selectedRun?.status, selectedJobId, selectedRunId, selectedRunTab, fetchRuns, fetchOutput]);

  // Auto-select first job if none selected
  useEffect(() => {
    if (!selectedJobId && jobs.length > 0) {
      setSelectedJobId(jobs[0].id);
    }
  }, [jobs, selectedJobId]);

  const filteredJobs = jobs.filter((job) => {
    if (filter === "all") return true;
    if (filter === "active") return job.scheduleEnabled;
    if (filter === "failed") return !job.scheduleEnabled;
    return true;
  });

  async function handleRunNow() {
    if (!selectedJobId) return;
    try {
      const run = await api.runJob(selectedJobId);
      await fetchRuns(selectedJobId);
      setSelectedTab("history");
      setSelectedRunId(run.id);
      setSelectedRunTab("session");
      onRefreshJobs();
    } catch (err) {
      console.error("failed to run job:", err);
    }
  }

  async function handleStopRun() {
    if (!selectedRunId || !selectedJobId) return;
    try {
      await api.cancelRun(selectedRunId);
      await fetchRuns(selectedJobId);
    } catch (err) {
      console.error("failed to stop run:", err);
    }
  }

  // Check if there's a running run for the selected job
  const hasRunningRun = runs.some((r) => r.status === "running");

  async function handleDelete() {
    if (!selectedJobId) return;
    try {
      await api.deleteJob(selectedJobId);
      setSelectedJobId(null);
      onRefreshJobs();
    } catch (err) {
      console.error("failed to delete job:", err);
    }
  }

  function renderFilterTabs() {
    const tabs: FilterTab[] = ["all", "active", "failed"];
    return (
      <div className="flex" style={{ borderBottom: "1px solid #2a2a2a" }}>
        {tabs.map((t) => (
          <button
            key={t}
            onClick={() => setFilter(t)}
            className="flex-1 flex items-center justify-center py-2 text-[10px] lowercase transition-colors"
            style={{
              fontFamily: font,
              fontWeight: filter === t ? 500 : "normal",
              color: filter === t ? "#10B981" : "#6B7280",
              borderBottom: filter === t ? "2px solid #10B981" : "2px solid transparent",
            }}
          >
            {t}
          </button>
        ))}
      </div>
    );
  }

  function renderJobList() {
    return (
      <div className="flex-1 overflow-y-auto">
        {filteredJobs.map((job) => {
          const active = job.id === selectedJobId;
          return (
            <button
              key={job.id}
              onClick={() => setSelectedJobId(job.id)}
              className="flex items-center gap-2 w-full px-4 py-2 text-left transition-colors"
              style={{
                backgroundColor: active ? "#1F1F1F" : "transparent",
                fontFamily: font,
              }}
              onMouseEnter={(e) => { if (!active) e.currentTarget.style.backgroundColor = "#1F1F1F"; }}
              onMouseLeave={(e) => { if (!active) e.currentTarget.style.backgroundColor = "transparent"; }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: "50%",
                  backgroundColor: job.scheduleEnabled ? "#10B981" : "#6B7280",
                  flexShrink: 0,
                }}
              />
              <span
                style={{
                  color: active ? "#FAFAFA" : "#9CA3AF",
                  fontSize: 11,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {job.name}
              </span>
            </button>
          );
        })}
      </div>
    );
  }

  function renderKeyValue(key: string, value: string | number | boolean | string[] | undefined | null) {
    let displayValue: string;
    let color = "#FAFAFA";

    if (value === undefined || value === null || value === "") {
      displayValue = "---";
      color = "#6B7280";
    } else if (typeof value === "boolean") {
      displayValue = value ? "true" : "false";
      color = value ? "#10B981" : "#EF4444";
    } else if (Array.isArray(value)) {
      displayValue = value.length > 0 ? value.join(", ") : "---";
      if (value.length === 0) color = "#6B7280";
    } else {
      displayValue = String(value);
    }

    return (
      <div
        key={key}
        style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", fontSize: 11, fontFamily: font }}
      >
        <span style={{ color: "#6B7280" }}>{key}:</span>
        <span style={{ color, textAlign: "right", wordBreak: "break-all" }}>{displayValue}</span>
      </div>
    );
  }

  function renderSection(title: string, rows: React.ReactNode) {
    return (
      <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
        <span style={{ color: "#4B5563", fontSize: 10, fontFamily: font }}>
          # {title}
        </span>
        <div style={{ height: 1, backgroundColor: "#2a2a2a" }} />
        <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
          {rows}
        </div>
      </div>
    );
  }

  function renderSettingsTab() {
    if (!selectedJob) return null;

    return (
      <div
        className="flex-1 overflow-y-auto"
        style={{ padding: "16px 20px", display: "flex", flexDirection: "column", gap: 16 }}
      >
        {renderSection("general", <>
          {renderKeyValue("type", selectedJob.type)}
          {renderKeyValue("name", selectedJob.name)}
          {renderKeyValue("description", selectedJob.description)}
          {renderKeyValue("working_dir", selectedJob.workingDirectory)}
        </>)}

        {renderSection("schedule", <>
          {renderKeyValue("enabled", selectedJob.scheduleEnabled)}
          {renderKeyValue("type", selectedJob.scheduleType)}
          {renderKeyValue("interval", selectedJob.scheduleInterval ? `${selectedJob.scheduleInterval}m` : "")}
          {renderKeyValue("cron", selectedJob.cronExpression)}
          {renderKeyValue("timeout", selectedJob.timeoutSeconds ? `${selectedJob.timeoutSeconds}s` : "")}
        </>)}

        {renderSection("triggers", <>
          {renderKeyValue("on_success", selectedJob.onSuccess)}
          {renderKeyValue("on_failure", selectedJob.onFailure)}
          {renderKeyValue("triggered_by", selectedJob.triggeredBy.length > 0
            ? selectedJob.triggeredBy.map((ref) => {
                const j = jobs.find((jj) => jj.id === ref.jobId);
                return `${j?.name ?? ref.jobId.slice(0, 8)} (${ref.triggerOn})`;
              }).join(", ")
            : "")}
        </>)}

        {selectedJob.type === "claude"
          ? renderSection("session", <>
              {renderKeyValue("prompt", selectedJob.prompt)}
              {renderKeyValue("model", selectedJob.model)}
              {renderKeyValue("allow_bypass", selectedJob.allowBypass)}
              {renderKeyValue("autonomous_mode", selectedJob.autonomousMode)}
              {renderKeyValue("max_retries", selectedJob.maxRetries)}
              {renderKeyValue("override_repo_command", selectedJob.overrideRepoCommand)}
              {renderKeyValue("claude_command", selectedJob.claudeCommand)}
            </>)
          : renderSection("script", <>
              {renderKeyValue("interpreter", selectedJob.interpreter)}
              {renderKeyValue("script_content", selectedJob.scriptContent)}
            </>)
        }
      </div>
    );
  }

  function renderHistoryTab() {
    return (
      <div className="flex flex-1 overflow-hidden">
        {/* Runs sub-sidebar */}
        <div
          className="flex flex-col h-full shrink-0 overflow-y-auto"
          style={{ width: 220, borderRight: "1px solid #2a2a2a" }}
        >
          {runs.length === 0 ? (
            <div className="flex items-center justify-center p-4">
              <span style={{ color: "#6B7280", fontSize: 11, fontFamily: font }}>no runs yet</span>
            </div>
          ) : (
            runs.map((run) => {
              const active = run.id === selectedRunId;
              return (
                <button
                  key={run.id}
                  onClick={() => { setSelectedRunId(run.id); setSelectedRunTab("session"); }}
                  className="flex items-center gap-2 w-full px-3 py-2 text-left transition-colors"
                  style={{
                    backgroundColor: active ? "#1F1F1F" : "transparent",
                    fontFamily: font,
                  }}
                  onMouseEnter={(e) => { if (!active) e.currentTarget.style.backgroundColor = "#1F1F1F"; }}
                  onMouseLeave={(e) => { if (!active) e.currentTarget.style.backgroundColor = "transparent"; }}
                >
                  <span
                    style={{
                      width: 6,
                      height: 6,
                      borderRadius: "50%",
                      backgroundColor: statusColor(run.status),
                      flexShrink: 0,
                      animation: run.status === "running" ? "job-pulse 1.5s ease-in-out infinite" : "none",
                    }}
                  />
                  <div className="flex flex-col overflow-hidden" style={{ gap: 2 }}>
                    <span style={{ color: "#FAFAFA", fontSize: 11, fontFamily: font }}>
                      {run.id.slice(0, 8)}
                    </span>
                    <span style={{ color: "#6B7280", fontSize: 9, fontFamily: font }}>
                      {relativeTime(run.startedAt)}
                    </span>
                  </div>
                </button>
              );
            })
          )}
        </div>

        {/* Run detail area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {!selectedRun ? (
            <div className="flex items-center justify-center flex-1">
              <span style={{ color: "#6B7280", fontSize: 11, fontFamily: font }}>select a run</span>
            </div>
          ) : (
            <>
              {/* Run sub-tabs */}
              <div className="flex" style={{ borderBottom: "1px solid #2a2a2a" }}>
                {(["session", "result"] as RunTab[]).map((t) => (
                  <button
                    key={t}
                    onClick={() => setSelectedRunTab(t)}
                    className="flex items-center justify-center px-4 py-2 text-[10px] lowercase transition-colors"
                    style={{
                      fontFamily: font,
                      fontWeight: selectedRunTab === t ? 500 : "normal",
                      color: selectedRunTab === t ? "#10B981" : "#6B7280",
                      borderBottom: selectedRunTab === t ? "2px solid #10B981" : "2px solid transparent",
                    }}
                  >
                    {t}
                  </button>
                ))}
              </div>

              {/* Sub-tab content */}
              {selectedRunTab === "session" ? (
                <div className="flex-1 overflow-y-auto p-4" style={{ position: "relative" }}>
                  {runOutput && (
                    <button
                      onClick={() => {
                        navigator.clipboard.writeText(runOutput);
                        setCopied(true);
                        setTimeout(() => setCopied(false), 2000);
                      }}
                      style={{
                        position: "sticky",
                        top: 0,
                        float: "right",
                        background: "none",
                        border: "1px solid #2a2a2a",
                        borderRadius: 4,
                        padding: "4px 8px",
                        cursor: "pointer",
                        color: copied ? "#10B981" : "#6B7280",
                        fontSize: 10,
                        fontFamily: font,
                        zIndex: 1,
                      }}
                      onMouseEnter={(e) => { if (!copied) e.currentTarget.style.color = "#FAFAFA"; }}
                      onMouseLeave={(e) => { if (!copied) e.currentTarget.style.color = "#6B7280"; }}
                      title="copy output"
                    >
                      {copied ? "✓ copied" : "⧉ copy"}
                    </button>
                  )}
                  {selectedRun.status === "running" && !runOutput && (
                    <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 12 }}>
                      <span
                        style={{
                          width: 8,
                          height: 8,
                          borderRadius: "50%",
                          backgroundColor: "#3B82F6",
                          animation: "job-pulse 1.5s ease-in-out infinite",
                          display: "inline-block",
                        }}
                      />
                      <span style={{ color: "#3B82F6", fontSize: 11, fontFamily: font }}>
                        running...
                      </span>
                    </div>
                  )}
                  {selectedRun.status === "running" && runOutput && (
                    <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 12 }}>
                      <span
                        style={{
                          width: 8,
                          height: 8,
                          borderRadius: "50%",
                          backgroundColor: "#3B82F6",
                          animation: "job-pulse 1.5s ease-in-out infinite",
                          display: "inline-block",
                        }}
                      />
                      <span style={{ color: "#3B82F6", fontSize: 11, fontFamily: font }}>
                        running... output updating every 3s
                      </span>
                    </div>
                  )}
                  <pre
                    style={{
                      color: "#FAFAFA",
                      fontSize: 11,
                      fontFamily: font,
                      whiteSpace: "pre-wrap",
                      wordBreak: "break-word",
                      margin: 0,
                    }}
                  >
                    {runOutput || (selectedRun.status !== "running" ? "no output available" : "")}
                  </pre>
                </div>
              ) : (
                <div
                  className="flex-1 overflow-y-auto"
                  style={{ padding: "16px 20px", display: "flex", flexDirection: "column", gap: 16 }}
                >
                  {renderSection("execution", <>
                    {renderKeyValue("run_id", selectedRun.id)}
                    {renderKeyValue("status", selectedRun.status)}
                    {renderKeyValue("triggered_by", selectedRun.triggeredBy || "manual")}
                    {renderKeyValue("started", selectedRun.startedAt ? new Date(selectedRun.startedAt).toLocaleString() : "---")}
                    {selectedRun.finishedAt && renderKeyValue("finished", new Date(selectedRun.finishedAt).toLocaleString())}
                    {renderKeyValue("duration", formatDuration(selectedRun.durationMs))}
                    {selectedRun.tokensUsed > 0 && renderKeyValue("tokens_used", selectedRun.tokensUsed.toLocaleString())}
                  </>)}

                  {selectedRun.sessionId && renderSection("triggered_sessions",
                    <div style={{ fontSize: 11, fontFamily: font }}>
                      <span style={{ color: "#10B981", cursor: "pointer" }}>
                        {selectedRun.sessionId}
                      </span>
                    </div>
                  )}

                  {selectedRun.errorMessage && renderSection("error",
                    <span style={{ color: "#EF4444", fontSize: 11, fontFamily: font }}>
                      {selectedRun.errorMessage}
                    </span>
                  )}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen w-screen" style={{ backgroundColor: "#0A0A0A", fontFamily: font }}>
      <style>{pulseKeyframes}</style>
      {/* Left sidebar */}
      <div
        className="flex flex-col h-full"
        style={{ width: 240, borderRight: "1px solid #2a2a2a", backgroundColor: "#0A0A0A" }}
      >
        {/* Header */}
        <div
          className="flex items-center px-4 py-3"
          style={{ height: 48, borderBottom: "1px solid #2a2a2a" }}
        >
          <h1 className="text-sm font-bold lowercase">
            <span style={{ color: "#10B981" }}>{">"}</span>{" "}
            <span style={{ color: "#FAFAFA" }}>jobs</span>
          </h1>
        </div>

        {/* Filter tabs */}
        {renderFilterTabs()}

        {/* Job list */}
        {renderJobList()}

        {/* Bottom bar */}
        <div className="flex items-center gap-2 p-3" style={{ borderTop: "1px solid #2a2a2a" }}>
          <button
            onClick={onCreateJob}
            className="flex-1 flex items-center justify-center gap-1 px-3 py-2 text-sm lowercase transition-colors"
            style={{
              backgroundColor: "#10B981",
              color: "#0A0A0A",
              fontFamily: font,
              fontWeight: 500,
            }}
            onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = "#059669")}
            onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = "#10B981")}
          >
            $ new_job
          </button>
        </div>
      </div>

      {/* Main content area */}
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        {!selectedJob ? (
          <div className="flex items-center justify-center flex-1" style={{ backgroundColor: "#0A0A0A" }}>
            <div className="text-center max-w-lg" style={{ fontFamily: font }}>
              <p className="text-3xl mb-3" style={{ color: "#10B981" }}>{">"}_  jobs</p>
              <p className="text-sm mb-8" style={{ color: "#6B7280" }}>
                automate tasks. schedule work. chain results.
              </p>

              <div className="text-xs text-left space-y-4" style={{ color: "#4B5563" }}>
                <div>
                  <p style={{ color: "#6B7280" }}>// what are jobs?</p>
                  <p>
                    <span style={{ color: "#10B981" }}>$</span> jobs run <span style={{ color: "#FAFAFA" }}>claude sessions</span> or{" "}
                    <span style={{ color: "#FAFAFA" }}>bash scripts</span> automatically
                  </p>
                  <p>
                    <span style={{ color: "#10B981" }}>$</span> schedule them to run{" "}
                    <span style={{ color: "#FAFAFA" }}>recurring</span>, <span style={{ color: "#FAFAFA" }}>one-time</span>, or{" "}
                    <span style={{ color: "#FAFAFA" }}>manually</span>
                  </p>
                  <p>
                    <span style={{ color: "#10B981" }}>$</span> chain jobs together with{" "}
                    <span style={{ color: "#FAFAFA" }}>on_success</span> and{" "}
                    <span style={{ color: "#FAFAFA" }}>on_failure</span> triggers
                  </p>
                </div>

                <div>
                  <p style={{ color: "#6B7280" }}>// getting started</p>
                  <p>
                    <span style={{ color: "#10B981" }}>1.</span> click{" "}
                    <span style={{ color: "#FAFAFA" }}>$ new_job</span> to create a job
                  </p>
                  <p>
                    <span style={{ color: "#10B981" }}>2.</span> configure a{" "}
                    <span style={{ color: "#FAFAFA" }}>prompt</span> or{" "}
                    <span style={{ color: "#FAFAFA" }}>script</span> to execute
                  </p>
                  <p>
                    <span style={{ color: "#10B981" }}>3.</span> set a{" "}
                    <span style={{ color: "#FAFAFA" }}>schedule</span> or trigger{" "}
                    <span style={{ color: "#FAFAFA" }}>manually</span>
                  </p>
                  <p>
                    <span style={{ color: "#10B981" }}>4.</span> view{" "}
                    <span style={{ color: "#FAFAFA" }}>run history</span> with full output and results
                  </p>
                </div>

                <div>
                  <p style={{ color: "#6B7280" }}>// examples</p>
                  <p>
                    <span style={{ color: "#06B6D4" }}>$</span> deploy monitor — check staging health every 30 min
                  </p>
                  <p>
                    <span style={{ color: "#06B6D4" }}>$</span> code review bot — review open PRs on schedule
                  </p>
                  <p>
                    <span style={{ color: "#06B6D4" }}>$</span> db backup — validate backups daily via bash script
                  </p>
                  <p>
                    <span style={{ color: "#06B6D4" }}>$</span> chain: deploy {"->"} test {"->"} notify on success or rollback on failure
                  </p>
                </div>

                <div>
                  <p style={{ color: "#6B7280" }}>// tips</p>
                  <p>
                    <span style={{ color: "#F59E0B" }}>$</span> claude jobs run with{" "}
                    <span style={{ color: "#FAFAFA" }}>-p</span> flag (non-interactive, exits when done)
                  </p>
                  <p>
                    <span style={{ color: "#F59E0B" }}>$</span> failed jobs auto-retry with previous output as context
                  </p>
                  <p>
                    <span style={{ color: "#F59E0B" }}>$</span> use{" "}
                    <span style={{ color: "#FAFAFA" }}>override repo command</span> to set a custom claude alias
                  </p>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <>
            {/* Top bar */}
            <div
              className="flex items-center justify-between px-6 shrink-0"
              style={{ height: 48, borderBottom: "1px solid #2a2a2a" }}
            >
              <div className="flex items-center gap-2">
                <span
                  style={{
                    width: 6,
                    height: 6,
                    borderRadius: "50%",
                    backgroundColor: selectedJob.scheduleEnabled ? "#10B981" : "#6B7280",
                  }}
                />
                <span style={{ color: "#FAFAFA", fontSize: 13, fontWeight: 500, fontFamily: font }}>
                  {selectedJob.name}
                </span>
              </div>
              <div className="flex items-center gap-3">
                {hasRunningRun ? (
                  <button
                    onClick={handleStopRun}
                    className="flex items-center gap-1 px-3 py-1 text-[11px] lowercase transition-colors"
                    style={{ color: "#EF4444", fontFamily: font }}
                    onMouseEnter={(e) => (e.currentTarget.style.color = "#DC2626")}
                    onMouseLeave={(e) => (e.currentTarget.style.color = "#EF4444")}
                  >
                    &#9632; stop
                  </button>
                ) : (
                  <button
                    onClick={handleRunNow}
                    className="flex items-center gap-1 px-3 py-1 text-[11px] lowercase transition-colors"
                    style={{ color: "#10B981", fontFamily: font }}
                    onMouseEnter={(e) => (e.currentTarget.style.color = "#059669")}
                    onMouseLeave={(e) => (e.currentTarget.style.color = "#10B981")}
                  >
                    &#9654; run now
                  </button>
                )}
                <button
                  onClick={() => selectedJob && onEditJob(selectedJob)}
                  className="flex items-center gap-1 px-3 py-1 text-[11px] lowercase transition-colors"
                  style={{ color: "#6B7280", fontFamily: font }}
                  onMouseEnter={(e) => (e.currentTarget.style.color = "#FAFAFA")}
                  onMouseLeave={(e) => (e.currentTarget.style.color = "#6B7280")}
                >
                  &#10000; edit
                </button>
                <button
                  onClick={handleDelete}
                  className="flex items-center gap-1 px-3 py-1 text-[11px] lowercase transition-colors"
                  style={{ color: "#EF4444", fontFamily: font }}
                  onMouseEnter={(e) => (e.currentTarget.style.color = "#DC2626")}
                  onMouseLeave={(e) => (e.currentTarget.style.color = "#EF4444")}
                >
                  &#10005; delete
                </button>
              </div>
            </div>

            {/* Tab bar */}
            <div className="flex" style={{ borderBottom: "1px solid #2a2a2a" }}>
              {(["settings", "history"] as JobTab[]).map((t) => (
                <button
                  key={t}
                  onClick={() => setSelectedTab(t)}
                  className="flex items-center justify-center px-6 py-2 text-[11px] lowercase transition-colors"
                  style={{
                    fontFamily: font,
                    fontWeight: selectedTab === t ? 500 : "normal",
                    color: selectedTab === t ? "#10B981" : "#6B7280",
                    borderBottom: selectedTab === t ? "2px solid #10B981" : "2px solid transparent",
                  }}
                >
                  {t}
                </button>
              ))}
            </div>

            {/* Tab content */}
            {selectedTab === "settings" && renderSettingsTab()}
            {selectedTab === "history" && renderHistoryTab()}
          </>
        )}
      </div>
    </div>
  );
}
