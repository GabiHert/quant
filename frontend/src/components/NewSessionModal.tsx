import { useState, useRef, useEffect } from "react";
import type { Repo, Task, CreateSessionRequest, SessionType } from "../types";

interface Props {
  repos: Repo[];
  tasksByRepo: Record<string, Task[]>;
  defaultRepoId?: string;
  defaultTaskId?: string;
  onSubmit: (req: CreateSessionRequest) => void;
  onCancel: () => void;
}

export function NewSessionModal({
  repos,
  tasksByRepo,
  defaultRepoId,
  defaultTaskId,
  onSubmit,
  onCancel,
}: Props) {
  const [sessionType, setSessionType] = useState<SessionType>("claude");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [repoId, setRepoId] = useState(defaultRepoId ?? repos[0]?.id ?? "");
  const [taskId, setTaskId] = useState(defaultTaskId ?? "");
  const [useWorktree, setUseWorktree] = useState(false);
  const [skipPermissions, setSkipPermissions] = useState(false);

  const tasks = tasksByRepo[repoId] ?? [];
  const selectedRepo = repos.find((r) => r.id === repoId);
  const selectedTask = tasks.find((t) => t.id === taskId);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || !repoId || !taskId) return;
    const req: CreateSessionRequest = {
      name: name.trim().toLowerCase(),
      description: description.trim().toLowerCase(),
      repoId,
      taskId,
      sessionType,
      useWorktree,
      skipPermissions: sessionType === "claude" ? skipPermissions : false,
    };
    onSubmit(req);
  }

  const inputStyle: React.CSSProperties = {
    backgroundColor: "#0A0A0A",
    border: "1px solid #2a2a2a",
    color: "#FAFAFA",
    fontFamily: "'JetBrains Mono', monospace",
  };

  const tabs: { key: SessionType; label: string }[] = [
    { key: "claude", label: "claude session" },
    { key: "terminal", label: "terminal" },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ backgroundColor: "rgba(0,0,0,0.7)" }}>
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-md"
        style={{
          backgroundColor: "#0A0A0A",
          border: "1px solid #2a2a2a",
          fontFamily: "'JetBrains Mono', monospace",
        }}
      >
        {/* title */}
        <div className="px-8 pt-8">
          <h2 className="text-sm font-bold lowercase" style={{ color: "#FAFAFA" }}>
            <span style={{ color: "#10B981" }}>{">"}</span> new_session
          </h2>
        </div>

        {/* tabs */}
        <div
          className="flex px-8 mt-4"
          style={{ borderBottom: "1px solid #2a2a2a" }}
        >
          {tabs.map((tab) => (
            <button
              key={tab.key}
              type="button"
              onClick={() => setSessionType(tab.key)}
              className="px-4 py-2 text-[11px] lowercase transition-colors"
              style={{
                color: sessionType === tab.key ? "#10B981" : "#6B7280",
                fontWeight: sessionType === tab.key ? 500 : "normal",
                borderBottom: sessionType === tab.key ? "2px solid #10B981" : "2px solid transparent",
                fontFamily: "'JetBrains Mono', monospace",
                marginBottom: -1,
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* form body */}
        <div className="px-8 pt-4 pb-8 flex flex-col gap-4">
          {/* repo dropdown */}
          <div>
            <span className="text-[10px] lowercase block mb-1" style={{ color: "#6B7280" }}>repo</span>
            <CustomSelect
              value={repoId}
              onChange={(v) => { setRepoId(v); setTaskId(""); }}
              options={repos.map((r) => ({ value: r.id, label: `${r.name} (${r.path})` }))}
              placeholder="select a repo"
              displayValue={selectedRepo ? `${selectedRepo.name} (${selectedRepo.path})` : ""}
            />
          </div>

          {/* task dropdown */}
          <div>
            <span className="text-[10px] lowercase block mb-1" style={{ color: "#6B7280" }}>task</span>
            <CustomSelect
              value={taskId}
              onChange={setTaskId}
              options={tasks.map((t) => ({ value: t.id, label: `# ${t.tag}  ${t.name}` }))}
              placeholder="select a task"
              displayValue={selectedTask ? `# ${selectedTask.tag}  ${selectedTask.name}` : ""}
            />
          </div>

          {/* name */}
          <label className="block">
            <span className="text-[10px] lowercase" style={{ color: "#6B7280" }}>name</span>
            <input
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={sessionType === "claude" ? "implement fix" : "deploy setup"}
              className="mt-1 block w-full px-3 py-2 text-xs focus:outline-none"
              style={inputStyle}
              onFocus={(e) => (e.currentTarget.style.borderColor = "#10B981")}
              onBlur={(e) => (e.currentTarget.style.borderColor = "#2a2a2a")}
            />
          </label>

          {/* description */}
          <label className="block">
            <span className="text-[10px] lowercase" style={{ color: "#6B7280" }}>description</span>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="what is this session for?"
              className="mt-1 block w-full px-3 py-2 text-xs focus:outline-none"
              style={inputStyle}
              onFocus={(e) => (e.currentTarget.style.borderColor = "#10B981")}
              onBlur={(e) => (e.currentTarget.style.borderColor = "#2a2a2a")}
            />
          </label>

          {/* options checkboxes */}
          <div className="flex flex-col gap-2">
            <label
              className="flex items-center gap-2 cursor-pointer"
              onClick={() => setUseWorktree(!useWorktree)}
            >
              <div
                className="flex items-center justify-center"
                style={{
                  width: 14,
                  height: 14,
                  backgroundColor: "#0A0A0A",
                  border: `1px solid ${useWorktree ? "#10B981" : "#2a2a2a"}`,
                }}
              >
                {useWorktree && (
                  <span style={{ color: "#10B981", fontSize: 10, lineHeight: 1 }}>x</span>
                )}
              </div>
              <span className="text-[10px] lowercase" style={{ color: "#6B7280" }}>
                use worktree
              </span>
            </label>
            {sessionType === "claude" && (
              <label
                className="flex items-center gap-2 cursor-pointer"
                onClick={() => setSkipPermissions(!skipPermissions)}
              >
                <div
                  className="flex items-center justify-center"
                  style={{
                    width: 14,
                    height: 14,
                    backgroundColor: "#0A0A0A",
                    border: `1px solid ${skipPermissions ? "#10B981" : "#2a2a2a"}`,
                  }}
                >
                  {skipPermissions && (
                    <span style={{ color: "#10B981", fontSize: 10, lineHeight: 1 }}>x</span>
                  )}
                </div>
                <span className="text-[10px] lowercase" style={{ color: "#6B7280" }}>
                  skip permissions
                </span>
              </label>
            )}
          </div>

          <div className="flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 text-xs lowercase transition-colors"
              style={{ color: "#6B7280" }}
              onMouseEnter={(e) => (e.currentTarget.style.color = "#FAFAFA")}
              onMouseLeave={(e) => (e.currentTarget.style.color = "#6B7280")}
            >
              cancel
            </button>
            <button
              type="submit"
              disabled={!name.trim() || !repoId || !taskId}
              className="px-4 py-2 text-xs lowercase transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              style={{
                backgroundColor: "#10B981",
                color: "#0A0A0A",
                fontWeight: 500,
              }}
            >
              create
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}

// Custom dropdown matching the Pencil design
function CustomSelect({
  value,
  onChange,
  options,
  placeholder,
  displayValue,
}: {
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
  placeholder: string;
  displayValue: string;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-3 py-2 text-xs text-left"
        style={{
          backgroundColor: "#0A0A0A",
          border: `1px solid ${open ? "#10B981" : "#2a2a2a"}`,
          color: value ? "#FAFAFA" : "#4B5563",
          fontFamily: "'JetBrains Mono', monospace",
        }}
      >
        <span className="overflow-hidden whitespace-nowrap" style={{ textOverflow: "ellipsis" }}>
          {displayValue || placeholder}
        </span>
        <span style={{ color: "#6B7280", fontSize: 10 }}>v</span>
      </button>
      {open && (
        <div
          className="absolute z-10 w-full mt-1 max-h-40 overflow-y-auto"
          style={{
            backgroundColor: "#0A0A0A",
            border: "1px solid #2a2a2a",
          }}
        >
          {options.length === 0 && (
            <div
              className="px-3 py-2 text-xs"
              style={{ color: "#4B5563", fontFamily: "'JetBrains Mono', monospace" }}
            >
              // none available
            </div>
          )}
          {options.map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => { onChange(opt.value); setOpen(false); }}
              className="w-full text-left px-3 py-2 text-xs transition-colors"
              style={{
                color: opt.value === value ? "#10B981" : "#FAFAFA",
                backgroundColor: opt.value === value ? "#1F1F1F" : "transparent",
                fontFamily: "'JetBrains Mono', monospace",
              }}
              onMouseEnter={(e) => {
                if (opt.value !== value) e.currentTarget.style.backgroundColor = "#1F1F1F";
              }}
              onMouseLeave={(e) => {
                if (opt.value !== value) e.currentTarget.style.backgroundColor = "transparent";
              }}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
