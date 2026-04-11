import { useCallback, useEffect, useRef, useState } from "react";
import { formatKeyCombo } from "../keybindings";

export interface PaletteCommand {
  id: string;
  label: string;
  category: string;
  shortcut?: string;
  onExecute: () => void;
}

interface Props {
  commands: PaletteCommand[];
  onClose: () => void;
}

const font = "'JetBrains Mono', monospace";

function fuzzyMatch(query: string, text: string): boolean {
  const q = query.toLowerCase();
  const t = text.toLowerCase();
  let qi = 0;
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] === q[qi]) qi++;
  }
  return qi === q.length;
}

export function CommandPalette({ commands, onClose }: Props) {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const filtered = query
    ? commands.filter((c) => fuzzyMatch(query, c.label) || fuzzyMatch(query, c.category))
    : commands;

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  // Scroll selected item into view
  useEffect(() => {
    const el = listRef.current;
    if (!el) return;
    const item = el.children[selectedIndex] as HTMLElement | undefined;
    if (item) item.scrollIntoView({ block: "nearest" });
  }, [selectedIndex]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
        return;
      }
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex((i) => (i + 1) % Math.max(filtered.length, 1));
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex((i) => (i - 1 + Math.max(filtered.length, 1)) % Math.max(filtered.length, 1));
        return;
      }
      if (e.key === "Enter") {
        e.preventDefault();
        if (filtered[selectedIndex]) {
          filtered[selectedIndex].onExecute();
          onClose();
        }
        return;
      }
    },
    [filtered, selectedIndex, onClose],
  );

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 99999,
        display: "flex",
        justifyContent: "center",
        paddingTop: 80,
      }}
      onClick={onClose}
    >
      {/* backdrop */}
      <div style={{ position: "absolute", inset: 0, backgroundColor: "var(--q-modal-backdrop)" }} />

      {/* palette */}
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          position: "relative",
          width: 520,
          maxHeight: "60vh",
          backgroundColor: "var(--q-bg-elevated)",
          border: "1px solid var(--q-border)",
          borderRadius: 8,
          boxShadow: "0 8px 32px rgba(0,0,0,0.5)",
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
          fontFamily: font,
        }}
      >
        {/* search input */}
        <div style={{ padding: "12px 16px", borderBottom: "1px solid var(--q-border)" }}>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="type a command..."
            style={{
              width: "100%",
              backgroundColor: "transparent",
              border: "none",
              outline: "none",
              color: "var(--q-fg)",
              fontFamily: font,
              fontSize: 14,
            }}
          />
        </div>

        {/* results */}
        <div
          ref={listRef}
          style={{
            flex: 1,
            overflowY: "auto",
            padding: "4px 0",
          }}
        >
          {filtered.length === 0 && (
            <div style={{ padding: "12px 16px", color: "var(--q-fg-muted)", fontSize: 12 }}>
              no matching commands
            </div>
          )}
          {filtered.map((cmd, i) => {
            const isSelected = i === selectedIndex;
            return (
              <div
                key={cmd.id}
                onClick={() => {
                  cmd.onExecute();
                  onClose();
                }}
                onMouseEnter={() => setSelectedIndex(i)}
                style={{
                  padding: "8px 16px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  cursor: "pointer",
                  backgroundColor: isSelected ? "var(--q-bg-hover)" : "transparent",
                  color: isSelected ? "var(--q-fg)" : "var(--q-fg-secondary)",
                  fontSize: 13,
                }}
              >
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <span style={{ color: "var(--q-fg-muted)", fontSize: 10, minWidth: 60 }}>
                    {cmd.category}
                  </span>
                  <span>{cmd.label}</span>
                </div>
                {cmd.shortcut && (
                  <span
                    style={{
                      fontSize: 11,
                      color: "var(--q-fg-muted)",
                      backgroundColor: "var(--q-bg-input)",
                      padding: "2px 6px",
                      borderRadius: 3,
                      border: "1px solid var(--q-border-light)",
                    }}
                  >
                    {formatKeyCombo(cmd.shortcut)}
                  </span>
                )}
              </div>
            );
          })}
        </div>

        {/* footer hint */}
        <div
          style={{
            padding: "8px 16px",
            borderTop: "1px solid var(--q-border)",
            display: "flex",
            gap: 16,
            fontSize: 10,
            color: "var(--q-fg-muted)",
          }}
        >
          <span>
            <kbd style={{ backgroundColor: "var(--q-bg-input)", padding: "1px 4px", borderRadius: 2, border: "1px solid var(--q-border-light)" }}>
              ↑↓
            </kbd>{" "}
            navigate
          </span>
          <span>
            <kbd style={{ backgroundColor: "var(--q-bg-input)", padding: "1px 4px", borderRadius: 2, border: "1px solid var(--q-border-light)" }}>
              ↵
            </kbd>{" "}
            execute
          </span>
          <span>
            <kbd style={{ backgroundColor: "var(--q-bg-input)", padding: "1px 4px", borderRadius: 2, border: "1px solid var(--q-border-light)" }}>
              esc
            </kbd>{" "}
            dismiss
          </span>
        </div>
      </div>
    </div>
  );
}
