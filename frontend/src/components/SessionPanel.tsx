import { useState, useEffect, useRef } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import type { Session, Task } from "../types";
import { StatusDot } from "./StatusDot";

interface Props {
  session: Session;
  task: Task | null;
  onStart: (id: string) => void;
  onStop: (id: string) => void;
  onResume: (id: string) => void;
  onDelete: (id: string) => void;
  onSendMessage: (id: string, message: string) => void;
  onClose: () => void;
}

export function SessionPanel({
  session,
  task,
  onStart,
  onStop,
  onResume,
  onDelete,
  onSendMessage,
  onClose,
}: Props) {
  const termRef = useRef<HTMLDivElement>(null);
  const termInstance = useRef<Terminal | null>(null);
  const fitAddon = useRef<FitAddon | null>(null);
  const [inputMessage, setInputMessage] = useState("");

  const isRunning = session.status === "running";

  // Initialize xterm.js terminal
  useEffect(() => {
    if (!termRef.current) return;

    const term = new Terminal({
      theme: {
        background: "#0A0A0A",
        foreground: "#FAFAFA",
        cursor: "#10B981",
        selectionBackground: "#1F1F1F",
        black: "#0A0A0A",
        red: "#EF4444",
        green: "#10B981",
        yellow: "#F59E0B",
        blue: "#3B82F6",
        magenta: "#8B5CF6",
        cyan: "#06B6D4",
        white: "#FAFAFA",
        brightBlack: "#4B5563",
        brightRed: "#EF4444",
        brightGreen: "#10B981",
        brightYellow: "#F59E0B",
        brightBlue: "#3B82F6",
        brightMagenta: "#8B5CF6",
        brightCyan: "#06B6D4",
        brightWhite: "#FFFFFF",
      },
      fontFamily: "'JetBrains Mono', monospace",
      fontSize: 13,
      lineHeight: 1.3,
      cursorBlink: false,
      cursorStyle: "block",
      scrollback: 10000,
      convertEol: true,
    });

    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(termRef.current);
    fit.fit();

    termInstance.current = term;
    fitAddon.current = fit;

    // Handle window resize
    const resizeObserver = new ResizeObserver(() => {
      try { fit.fit(); } catch { /* ignore */ }
    });
    resizeObserver.observe(termRef.current);

    return () => {
      resizeObserver.disconnect();
      term.dispose();
      termInstance.current = null;
      fitAddon.current = null;
    };
  }, [session.id]); // re-create terminal when session changes

  // Listen for PTY output events
  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const w = window as any;
    if (!w?.runtime?.EventsOn) return;

    const cancel = w.runtime.EventsOn("session:output", (data: { sessionId: string; data: string }) => {
      if (data.sessionId === session.id && termInstance.current) {
        termInstance.current.write(data.data);
      }
    });

    return () => { if (cancel) cancel(); };
  }, [session.id]);

  // Listen for process exit
  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const w = window as any;
    if (!w?.runtime?.EventsOn) return;

    const cancel = w.runtime.EventsOn("session:exited", (sessionId: string) => {
      if (sessionId === session.id && termInstance.current) {
        termInstance.current.write("\r\n\x1b[90m// process exited\x1b[0m\r\n");
      }
    });

    return () => { if (cancel) cancel(); };
  }, [session.id]);

  function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!inputMessage.trim()) return;
    onSendMessage(session.id, inputMessage.trim());
    setInputMessage("");
  }

  return (
    <div className="flex flex-col h-full" style={{ backgroundColor: "#0A0A0A" }}>
      {/* tab bar */}
      <div
        className="flex items-center justify-between px-4 py-2 shrink-0"
        style={{
          backgroundColor: "#0A0A0A",
          borderBottom: "1px solid #2a2a2a",
          fontFamily: "'JetBrains Mono', monospace",
        }}
      >
        <div className="flex items-center gap-2 overflow-hidden">
          <StatusDot status={session.status} />
          <span
            className="text-xs font-bold overflow-hidden whitespace-nowrap"
            style={{ color: "#FAFAFA", textOverflow: "ellipsis" }}
          >
            {session.name}
          </span>
          {task && (
            <span
              className="shrink-0 text-[9px] px-1.5 py-0.5"
              style={{
                color: "#10B981",
                border: "1px solid #2a2a2a",
                backgroundColor: "#0A0A0A",
              }}
            >
              # {task.tag}
            </span>
          )}
          {session.worktreePath && (
            <span
              className="shrink-0 text-[9px] px-1.5 py-0.5"
              style={{
                color: "#4B5563",
                border: "1px solid #2a2a2a",
              }}
            >
              wt {session.branchName}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {session.status === "idle" && (
            <ActionBtn label="$ start" onClick={() => onStart(session.id)} color="#10B981" />
          )}
          {isRunning && (
            <ActionBtn label="$ stop" onClick={() => onStop(session.id)} color="#F59E0B" />
          )}
          {session.status === "paused" && (
            <ActionBtn label="$ resume" onClick={() => onResume(session.id)} color="#10B981" />
          )}
          <ActionBtn label="$ delete" onClick={() => onDelete(session.id)} color="#EF4444" />
          <button
            onClick={onClose}
            className="ml-1 text-xs transition-colors"
            style={{ color: "#6B7280" }}
            onMouseEnter={(e) => (e.currentTarget.style.color = "#FAFAFA")}
            onMouseLeave={(e) => (e.currentTarget.style.color = "#6B7280")}
            title="close tab"
          >
            [x]
          </button>
        </div>
      </div>

      {/* info bar */}
      <div
        className="px-4 py-2 shrink-0"
        style={{ borderBottom: "1px solid #2a2a2a" }}
      >
        {session.description && (
          <p
            className="text-xs mb-1"
            style={{
              color: "#6B7280",
              fontFamily: "'IBM Plex Mono', monospace",
            }}
          >
            // {session.description}
          </p>
        )}
        <p
          className="text-[10px]"
          style={{
            color: "#4B5563",
            fontFamily: "'JetBrains Mono', monospace",
          }}
        >
          dir: {session.directory}
          {session.claudeConvId && (
            <span className="ml-3">
              conv: {session.claudeConvId.slice(0, 8)}...
            </span>
          )}
          {session.pid > 0 && (
            <span className="ml-3">pid: {session.pid}</span>
          )}
        </p>
      </div>

      {/* xterm.js terminal */}
      <div
        ref={termRef}
        className="flex-1 min-h-0"
        style={{ padding: "4px 0 0 4px" }}
      />

      {/* chat bar */}
      <form
        onSubmit={handleSend}
        className="flex items-center gap-2 px-4 py-3 shrink-0"
        style={{
          backgroundColor: "#0A0A0A",
          borderTop: "1px solid #2a2a2a",
        }}
      >
        <span
          className="shrink-0 text-sm"
          style={{ color: "#10B981", fontFamily: "'JetBrains Mono', monospace" }}
        >
          {">"}
        </span>
        <input
          value={inputMessage}
          onChange={(e) => setInputMessage(e.target.value)}
          placeholder={isRunning ? "send a message..." : "start the session to send messages"}
          disabled={!isRunning}
          className="flex-1 px-3 py-2 text-xs focus:outline-none disabled:opacity-40 disabled:cursor-not-allowed"
          style={{
            backgroundColor: "#0A0A0A",
            border: "1px solid #2a2a2a",
            color: "#FAFAFA",
            fontFamily: "'JetBrains Mono', monospace",
          }}
          onFocus={(e) => (e.currentTarget.style.borderColor = "#10B981")}
          onBlur={(e) => (e.currentTarget.style.borderColor = "#2a2a2a")}
        />
        <button
          type="submit"
          disabled={!isRunning || !inputMessage.trim()}
          className="px-4 py-2 text-xs transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          style={{
            backgroundColor: "#10B981",
            color: "#0A0A0A",
            fontFamily: "'JetBrains Mono', monospace",
            fontWeight: 500,
          }}
        >
          send
        </button>
      </form>
    </div>
  );
}

function ActionBtn({
  label,
  onClick,
  color,
}: {
  label: string;
  onClick: () => void;
  color: string;
}) {
  return (
    <button
      onClick={onClick}
      className="px-2 py-1 text-[10px] lowercase transition-colors"
      style={{
        color,
        fontFamily: "'JetBrains Mono', monospace",
      }}
      onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = "#1F1F1F")}
      onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = "transparent")}
    >
      {label}
    </button>
  );
}
