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

export function StatusDot({ status, className = "" }: Props) {
  return (
    <span
      className={`inline-block h-2 w-2 shrink-0 ${className}`}
      style={{ backgroundColor: statusColors[status] }}
      title={status}
    />
  );
}
