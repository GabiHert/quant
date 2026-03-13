import { useEffect, useRef } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebglAddon } from "@xterm/addon-webgl";
import "@xterm/xterm/css/xterm.css";
import type { Session, Task } from "../types";
import { StatusDot } from "./StatusDot";
import * as api from "../api";

interface Props {
  session: Session;
  task: Task | null;
  onStart: (id: string, rows?: number, cols?: number) => void;
  onStop: (id: string) => void;
  onResume: (id: string, rows?: number, cols?: number) => void;
  onDelete: (id: string) => void;
  onClose: () => void;
}

export function SessionPanel({
  session,
  task,
  onStart,
  onStop,
  onResume,
  onDelete,
  onClose,
}: Props) {
  const termRef = useRef<HTMLDivElement>(null);
  const termInstance = useRef<Terminal | null>(null);
  const fitAddon = useRef<FitAddon | null>(null);

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
      cursorBlink: true,
      cursorStyle: "block",
      scrollback: 10000,
      convertEol: true,
    });

    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(termRef.current);

    // Use WebGL renderer for crisp rendering without ghosting artifacts.
    try {
      const webgl = new WebglAddon();
      term.loadAddon(webgl);
    } catch {
      // WebGL not available, fall back to canvas renderer.
    }

    fit.fit();

    termInstance.current = term;
    fitAddon.current = fit;

    // Send terminal input (keystrokes) to the PTY
    term.onData((data) => {
      api.sendMessage(session.id, data).catch(() => {
        // Process may not be running
      });
    });

    // Sync PTY size when terminal resizes
    term.onResize(({ rows, cols }) => {
      api.resizeTerminal(session.id, rows, cols).catch(() => {
        // Process may not be running
      });
    });

    // Handle window/container resize
    const resizeObserver = new ResizeObserver(() => {
      try { fit.fit(); } catch { /* ignore */ }
    });
    resizeObserver.observe(termRef.current);

    // Initial PTY size sync
    api.resizeTerminal(session.id, term.rows, term.cols).catch(() => {});

    // Replay saved output from disk
    api.getSessionOutput(session.id).then((output) => {
      if (output && term) {
        term.write(output);
      }
    }).catch(() => {});

    // Focus the terminal
    term.focus();

    return () => {
      resizeObserver.disconnect();
      term.dispose();
      termInstance.current = null;
      fitAddon.current = null;
    };
  }, [session.id]);

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
                color: "#10B981",
                border: "1px solid #10B981",
              }}
            >
              wt {session.branchName}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {session.status === "idle" && (
            <ActionBtn label="$ start" onClick={() => onStart(session.id, termInstance.current?.rows, termInstance.current?.cols)} color="#10B981" />
          )}
          {isRunning && (
            <ActionBtn label="$ stop" onClick={() => onStop(session.id)} color="#F59E0B" />
          )}
          {session.status === "paused" && (
            <ActionBtn label="$ resume" onClick={() => onResume(session.id, termInstance.current?.rows, termInstance.current?.cols)} color="#10B981" />
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

      {/* full terminal — handles all input/output, fills remaining space */}
      <div
        ref={termRef}
        className="flex-1 min-h-0 w-full"
        style={{ overflow: "hidden" }}
        onClick={() => termInstance.current?.focus()}
      />
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
