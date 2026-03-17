import { useEffect, useRef, useState, useCallback } from "react";
import "@xterm/xterm/css/xterm.css";
import type { Session, Task, Config } from "../types";
import { StatusDot } from "./StatusDot";
import { TerminalPane } from "./TerminalPane";
import * as api from "../api";

type SplitLayout = "horizontal" | "vertical";

interface SplitState {
  open: boolean;
  terminalSession: Session | null;
  layout: SplitLayout;
  dividerPercent: number;
}

interface Props {
  session: Session;
  task: Task | null;
  onStart: (id: string, rows: number, cols: number) => void;
  onResume: (id: string, rows: number, cols: number) => void;
  onUnarchive?: (id: string) => void;
  displayStatus: import("./StatusBadge").DisplayStatus;
  onCreateEmbeddedTerminal: (parentSession: Session) => Promise<Session>;
  onDeleteEmbeddedTerminal: (terminalSessionId: string) => void;
}

export function SessionPanel({
  session,
  task,
  onStart,
  onResume,
  onUnarchive,
  displayStatus,
  onCreateEmbeddedTerminal,
  onDeleteEmbeddedTerminal,
}: Props) {
  const [autoScroll, setAutoScroll] = useState(true);
  const [termConfig, setTermConfig] = useState<Config | null>(null);
  const [splitState, setSplitState] = useState<SplitState>({
    open: false,
    terminalSession: null,
    layout: "horizontal",
    dividerPercent: 55,
  });
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const splitContainerRef = useRef<HTMLDivElement>(null!)
  const isDragging = useRef(false);

  const isArchived = displayStatus === "archived";
  const isPaused = displayStatus === "paused";

  // Load terminal config on mount
  useEffect(() => {
    api.getConfig().then((cfg) => {
      setTermConfig(cfg);
    }).catch(() => {});
  }, []);

  // Close embedded terminal when parent session changes
  useEffect(() => {
    if (splitState.open && splitState.terminalSession) {
      onDeleteEmbeddedTerminal(splitState.terminalSession.id);
      setSplitState(prev => ({ ...prev, open: false, terminalSession: null }));
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session.id]);

  // Close menu on click outside
  useEffect(() => {
    if (!menuOpen) return;
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [menuOpen]);

  // Split divider drag handling
  const handleDividerMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDragging.current = true;
    document.body.style.cursor = splitState.layout === "horizontal" ? "row-resize" : "col-resize";
    document.body.style.userSelect = "none";
  }, [splitState.layout]);

  useEffect(() => {
    function handleMouseMove(e: MouseEvent) {
      if (!isDragging.current || !splitContainerRef.current) return;
      const rect = splitContainerRef.current.getBoundingClientRect();
      let percent: number;
      if (splitState.layout === "horizontal") {
        percent = ((e.clientY - rect.top) / rect.height) * 100;
      } else {
        percent = ((e.clientX - rect.left) / rect.width) * 100;
      }
      setSplitState(prev => ({ ...prev, dividerPercent: Math.min(80, Math.max(20, percent)) }));
    }
    function handleMouseUp() {
      if (!isDragging.current) return;
      isDragging.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    }
    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [splitState.layout]);

  async function handleOpenTerminal() {
    if (splitState.open) return;
    try {
      const termSession = await onCreateEmbeddedTerminal(session);
      setSplitState(prev => ({ ...prev, open: true, terminalSession: termSession }));
    } catch {
      // Failed to create embedded terminal
    }
  }

  function handleCloseTerminal() {
    if (splitState.terminalSession) {
      onDeleteEmbeddedTerminal(splitState.terminalSession.id);
    }
    setSplitState(prev => ({ ...prev, open: false, terminalSession: null }));
  }

  function handleToggleLayout() {
    setSplitState(prev => ({
      ...prev,
      layout: prev.layout === "horizontal" ? "vertical" : "horizontal",
    }));
  }

  return (
    <div className="flex flex-col h-full" style={{ backgroundColor: "#0A0A0A" }}>
      {/* Action bar */}
      <div
        className="flex items-center justify-between px-5 shrink-0"
        style={{
          backgroundColor: "#0A0A0A",
          borderBottom: "1px solid #2a2a2a",
          fontFamily: "'JetBrains Mono', monospace",
          height: 32,
        }}
      >
        {/* Left: status + name + badges */}
        <div className="flex items-center gap-2 overflow-hidden">
          <StatusDot status={displayStatus} />
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
                backgroundColor: "#1F1F1F",
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
                border: "1px solid #2a2a2a",
                backgroundColor: "#1F1F1F",
              }}
            >
              wt {session.branchName}
            </span>
          )}
        </div>

        {/* Right: terminal btn + layout toggle + hamburger */}
        <div className="flex items-center gap-3 shrink-0">
          {/* Resume button (only when paused) */}
          {isPaused && !isArchived && (
            <ActionBtn label="$ resume" onClick={() => {
              // Resume is handled by TerminalPane auto-resume,
              // but provide manual button for paused sessions
              onResume(session.id, 24, 80);
            }} color="#10B981" />
          )}

          {/* Unarchive button */}
          {isArchived && onUnarchive && (
            <ActionBtn label="$ unarchive" onClick={() => onUnarchive(session.id)} color="#10B981" />
          )}

          {/* Terminal button */}
          {!isArchived && (
            <button
              onClick={splitState.open ? handleCloseTerminal : handleOpenTerminal}
              className="flex items-center gap-1 px-2 py-1 text-[11px]"
              style={{
                fontFamily: "'JetBrains Mono', monospace",
                color: splitState.open ? "#0A0A0A" : "#06B6D4",
                backgroundColor: splitState.open ? "#06B6D4" : "#1F1F1F",
                border: `1px solid ${splitState.open ? "#06B6D4" : "#2a2a2a"}`,
              }}
              onMouseEnter={(e) => {
                if (!splitState.open) e.currentTarget.style.backgroundColor = "#2a2a2a";
              }}
              onMouseLeave={(e) => {
                if (!splitState.open) e.currentTarget.style.backgroundColor = "#1F1F1F";
              }}
            >
              <span style={{ fontWeight: 700 }}>$</span>
              <span>terminal</span>
            </button>
          )}

          {/* Layout toggle (only when split is open) */}
          {splitState.open && (
            <div className="flex items-center gap-0.5">
              <LayoutIcon
                type="horizontal"
                active={splitState.layout === "horizontal"}
                onClick={() => setSplitState(prev => ({ ...prev, layout: "horizontal" }))}
              />
              <LayoutIcon
                type="vertical"
                active={splitState.layout === "vertical"}
                onClick={() => setSplitState(prev => ({ ...prev, layout: "vertical" }))}
              />
            </div>
          )}

          {/* Hamburger menu */}
          {!isArchived && (
            <div className="relative" ref={menuRef}>
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="flex items-center justify-center"
                style={{
                  width: 20,
                  height: 20,
                  color: menuOpen ? "#FAFAFA" : "#6B7280",
                }}
                onMouseEnter={(e) => { if (!menuOpen) e.currentTarget.style.color = "#FAFAFA"; }}
                onMouseLeave={(e) => { if (!menuOpen) e.currentTarget.style.color = "#6B7280"; }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="4" y1="6" x2="20" y2="6" />
                  <line x1="4" y1="12" x2="20" y2="12" />
                  <line x1="4" y1="18" x2="20" y2="18" />
                </svg>
              </button>

              {menuOpen && (
                <HamburgerMenu
                  autoScroll={autoScroll}
                  onAutoScrollToggle={() => { setAutoScroll(!autoScroll); }}
                />
              )}
            </div>
          )}
        </div>
      </div>

      {/* Split container */}
      <SplitContainer
        splitContainerRef={splitContainerRef}
        splitState={splitState}
        onDividerMouseDown={handleDividerMouseDown}
        primaryPane={
          <>
            {splitState.open && (
              <PaneHeader
                label={session.sessionType === "claude" ? "claude" : "terminal"}
                dotColor={session.sessionType === "claude" ? "#10B981" : "#06B6D4"}
              />
            )}
            <TerminalPane
              session={session}
              isArchived={isArchived}
              onStart={onStart}
              onResume={onResume}
              termConfig={termConfig}
              autoScroll={autoScroll}
              onAutoScrollChange={setAutoScroll}
            />
          </>
        }
        secondaryPane={
          splitState.open && splitState.terminalSession ? (
            <>
              <PaneHeader
                label="terminal"
                dotColor="#06B6D4"
                onClose={handleCloseTerminal}
              />
              <TerminalPane
                session={splitState.terminalSession}
                isArchived={false}
                onStart={onStart}
                onResume={onResume}
                termConfig={termConfig}
                autoScroll={true}
                onAutoScrollChange={() => {}}
              />
            </>
          ) : null
        }
      />
    </div>
  );
}

/**
 * SplitContainer uses absolute positioning so that pane resizes are instant
 * and xterm.js does not do a slow reflow animation. The parent is `position: relative`
 * and each pane is `position: absolute` with explicit top/left/width/height in pixels,
 * calculated from the container's own dimensions via a ResizeObserver.
 */
function SplitContainer({
  splitContainerRef,
  splitState,
  onDividerMouseDown,
  primaryPane,
  secondaryPane,
}: {
  splitContainerRef: React.RefObject<HTMLDivElement>;
  splitState: SplitState;
  onDividerMouseDown: (e: React.MouseEvent) => void;
  primaryPane: React.ReactNode;
  secondaryPane: React.ReactNode;
}) {
  const [size, setSize] = useState({ w: 0, h: 0 });

  useEffect(() => {
    const el = splitContainerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      const { width, height } = entries[0].contentRect;
      setSize({ w: width, h: height });
    });
    ro.observe(el);
    // Set initial size
    setSize({ w: el.clientWidth, h: el.clientHeight });
    return () => ro.disconnect();
  }, [splitContainerRef]);

  const DIVIDER = 6;
  const isH = splitState.layout === "horizontal";
  const isOpen = splitState.open && secondaryPane != null;

  // Calculate pixel sizes
  let primaryStyle: React.CSSProperties;
  let dividerStyle: React.CSSProperties | null = null;
  let secondaryStyle: React.CSSProperties | null = null;

  if (!isOpen) {
    primaryStyle = { position: "absolute", top: 0, left: 0, width: size.w, height: size.h };
  } else {
    const total = isH ? size.h : size.w;
    const primaryPx = Math.round((total - DIVIDER) * splitState.dividerPercent / 100);
    const secondaryPx = total - DIVIDER - primaryPx;

    if (isH) {
      primaryStyle = { position: "absolute", top: 0, left: 0, width: size.w, height: primaryPx };
      dividerStyle = { position: "absolute", top: primaryPx, left: 0, width: size.w, height: DIVIDER };
      secondaryStyle = { position: "absolute", top: primaryPx + DIVIDER, left: 0, width: size.w, height: secondaryPx };
    } else {
      primaryStyle = { position: "absolute", top: 0, left: 0, width: primaryPx, height: size.h };
      dividerStyle = { position: "absolute", top: 0, left: primaryPx, width: DIVIDER, height: size.h };
      secondaryStyle = { position: "absolute", top: 0, left: primaryPx + DIVIDER, width: secondaryPx, height: size.h };
    }
  }

  return (
    <div
      ref={splitContainerRef}
      className="flex-1 min-h-0"
      style={{ position: "relative", overflow: "hidden" }}
    >
      {/* Primary pane */}
      <div style={{ ...primaryStyle, display: "flex", flexDirection: "column", overflow: "hidden" }}>
        {primaryPane}
      </div>

      {/* Divider */}
      {isOpen && dividerStyle && (
        <div
          onMouseDown={onDividerMouseDown}
          style={{
            ...dividerStyle,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            cursor: isH ? "row-resize" : "col-resize",
            borderTop: isH ? "1px solid #2a2a2a" : undefined,
            borderBottom: isH ? "1px solid #2a2a2a" : undefined,
            borderLeft: !isH ? "1px solid #2a2a2a" : undefined,
            borderRight: !isH ? "1px solid #2a2a2a" : undefined,
            zIndex: 1,
          }}
          onMouseEnter={(e) => {
            const grip = e.currentTarget.querySelector("[data-grip]") as HTMLElement;
            if (grip) grip.style.backgroundColor = "#10B981";
          }}
          onMouseLeave={(e) => {
            const grip = e.currentTarget.querySelector("[data-grip]") as HTMLElement;
            if (grip) grip.style.backgroundColor = "#4B5563";
          }}
        >
          <div
            data-grip
            style={{
              width: isH ? 32 : 2,
              height: isH ? 2 : 32,
              backgroundColor: "#4B5563",
              borderRadius: 1,
            }}
          />
        </div>
      )}

      {/* Secondary pane */}
      {isOpen && secondaryStyle && (
        <div style={{ ...secondaryStyle, display: "flex", flexDirection: "column", overflow: "hidden" }}>
          {secondaryPane}
        </div>
      )}
    </div>
  );
}

function PaneHeader({
  label,
  dotColor,
  onClose,
}: {
  label: string;
  dotColor: string;
  onClose?: () => void;
}) {
  return (
    <div
      className="flex items-center justify-between px-4 shrink-0"
      style={{
        height: 24,
        backgroundColor: "#0F0F0F",
        borderBottom: "1px solid #2a2a2a",
        fontFamily: "'JetBrains Mono', monospace",
      }}
    >
      <div className="flex items-center gap-1.5">
        <div
          style={{
            width: 6,
            height: 6,
            borderRadius: "50%",
            backgroundColor: dotColor,
          }}
        />
        <span style={{ fontSize: 10, color: "#6B7280" }}>{label}</span>
      </div>
      {onClose && (
        <button
          onClick={onClose}
          className="text-[9px] transition-colors"
          style={{ color: "#4B5563", fontFamily: "'JetBrains Mono', monospace" }}
          onMouseEnter={(e) => (e.currentTarget.style.color = "#FAFAFA")}
          onMouseLeave={(e) => (e.currentTarget.style.color = "#4B5563")}
        >
          [x]
        </button>
      )}
    </div>
  );
}

function LayoutIcon({
  type,
  active,
  onClick,
}: {
  type: "horizontal" | "vertical";
  active: boolean;
  onClick: () => void;
}) {
  const borderColor = active ? "#10B981" : "#2a2a2a";
  const fillColor = active ? "#10B981" : "#6B7280";
  const isHorizontal = type === "horizontal";

  return (
    <button
      onClick={onClick}
      style={{
        width: 20,
        height: 16,
        backgroundColor: "#1F1F1F",
        border: `1px solid ${borderColor}`,
        padding: 2,
        display: "flex",
        flexDirection: isHorizontal ? "column" : "row",
        gap: 1,
        cursor: "pointer",
      }}
      title={`${type} split`}
    >
      <div
        style={{
          backgroundColor: fillColor,
          ...(isHorizontal
            ? { width: "100%", height: 5 }
            : { height: "100%", width: 6 }),
        }}
      />
      <div
        style={{
          backgroundColor: fillColor,
          opacity: 0.4,
          ...(isHorizontal
            ? { width: "100%", height: 5 }
            : { height: "100%", width: 6 }),
        }}
      />
    </button>
  );
}

function HamburgerMenu({
  autoScroll,
  onAutoScrollToggle,
}: {
  autoScroll: boolean;
  onAutoScrollToggle: () => void;
}) {
  return (
    <div
      style={{
        position: "absolute",
        top: "100%",
        right: 0,
        marginTop: 4,
        width: 160,
        backgroundColor: "#141414",
        border: "1px solid #2a2a2a",
        padding: "4px 0",
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 11,
        zIndex: 50,
        boxShadow: "0 4px 12px rgba(0,0,0,0.5)",
      }}
    >
      <MenuItemRow onClick={onAutoScrollToggle}>
        <span
          className="flex items-center justify-center"
          style={{
            width: 14,
            height: 14,
            border: `1px solid ${autoScroll ? "#10B981" : "#2a2a2a"}`,
            backgroundColor: "#0A0A0A",
            fontSize: 8,
            fontWeight: 700,
            color: "#10B981",
            lineHeight: 1,
          }}
        >
          {autoScroll ? "x" : ""}
        </span>
        <span style={{ color: "#FAFAFA" }}>auto-scroll</span>
      </MenuItemRow>
    </div>
  );
}

function MenuItemRow({
  onClick,
  children,
}: {
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-2 w-full px-3 transition-colors"
      style={{
        height: 32,
        background: "none",
        border: "none",
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 11,
        cursor: "pointer",
        textAlign: "left",
      }}
      onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = "#1F1F1F")}
      onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = "transparent")}
    >
      {children}
    </button>
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
