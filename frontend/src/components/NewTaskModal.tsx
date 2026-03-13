import { useState } from "react";
import type { CreateTaskRequest } from "../types";

interface Props {
  repoId: string;
  repoName?: string;
  onSubmit: (req: CreateTaskRequest) => void;
  onCancel: () => void;
}

export function NewTaskModal({ repoId, repoName, onSubmit, onCancel }: Props) {
  const [tag, setTag] = useState("");
  const [name, setName] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!tag.trim()) return;
    onSubmit({
      repoId,
      tag: tag.trim().toLowerCase(),
      name: name.trim().toLowerCase(),
    });
  }

  const inputStyle: React.CSSProperties = {
    backgroundColor: "#0A0A0A",
    border: "1px solid #2a2a2a",
    color: "#FAFAFA",
    fontFamily: "'JetBrains Mono', monospace",
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ backgroundColor: "rgba(0,0,0,0.7)" }}>
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-md p-8"
        style={{
          backgroundColor: "#0A0A0A",
          border: "1px solid #2a2a2a",
          fontFamily: "'JetBrains Mono', monospace",
        }}
      >
        <h2 className="text-sm font-bold lowercase mb-5" style={{ color: "#FAFAFA" }}>
          <span style={{ color: "#10B981" }}>{">"}</span> new_task
        </h2>

        <label className="block mb-4">
          <span className="text-[10px] lowercase" style={{ color: "#6B7280" }}>tag</span>
          <input
            autoFocus
            value={tag}
            onChange={(e) => setTag(e.target.value)}
            placeholder="PLT-123"
            className="mt-1 block w-full px-3 py-2 text-xs focus:outline-none"
            style={inputStyle}
            onFocus={(e) => (e.currentTarget.style.borderColor = "#10B981")}
            onBlur={(e) => (e.currentTarget.style.borderColor = "#2a2a2a")}
          />
        </label>

        <label className="block mb-4">
          <span className="text-[10px] lowercase" style={{ color: "#6B7280" }}>name</span>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="fix auth flow"
            className="mt-1 block w-full px-3 py-2 text-xs focus:outline-none"
            style={inputStyle}
            onFocus={(e) => (e.currentTarget.style.borderColor = "#10B981")}
            onBlur={(e) => (e.currentTarget.style.borderColor = "#2a2a2a")}
          />
        </label>

        {repoName && (
          <p
            className="mb-5 text-[10px]"
            style={{ color: "#4B5563", fontFamily: "'IBM Plex Mono', monospace" }}
          >
            // repo: {repoName}
          </p>
        )}

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
            disabled={!tag.trim()}
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
      </form>
    </div>
  );
}
