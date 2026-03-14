import type { DisplayStatus } from "./StatusBadge";

const statusColors: Record<DisplayStatus, string> = {
  running: "#10B981",
  paused: "#F59E0B",
  idle: "#6B7280",
  done: "#06B6D4",
  error: "#EF4444",
  starting: "#10B981",
  stopping: "#F59E0B",
  resuming: "#10B981",
};

const isTransitional = (s: DisplayStatus) =>
  s === "starting" || s === "stopping" || s === "resuming";

interface Props {
  status: DisplayStatus;
  className?: string;
}

export function StatusDot({ status, className = "" }: Props) {
  return (
    <span
      className={`inline-block h-2 w-2 shrink-0 ${className}`}
      style={{
        backgroundColor: statusColors[status],
        animation: isTransitional(status) ? "pulse-dot 0.8s ease-in-out infinite" : undefined,
      }}
      title={status}
    />
  );
}
