import type { Session } from "../types";

const statusColors: Record<Session["status"], string> = {
  running: "#10B981",
  paused: "#F59E0B",
  idle: "#6B7280",
  done: "#06B6D4",
  error: "#EF4444",
};

interface Props {
  status: Session["status"];
  className?: string;
}

export function StatusBadge({ status, className = "" }: Props) {
  return (
    <span
      className={`shrink-0 ${className}`}
      style={{
        color: statusColors[status],
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: "9px",
      }}
    >
      [{status}]
    </span>
  );
}
