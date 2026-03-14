import { useEffect, useState } from "react";
import type { Session } from "../types";

export type DisplayStatus = Session["status"] | "starting" | "stopping" | "resuming";

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

function AnimatedDots() {
  const [dots, setDots] = useState(1);

  useEffect(() => {
    const interval = setInterval(() => {
      setDots((d) => (d % 3) + 1);
    }, 400);
    return () => clearInterval(interval);
  }, []);

  return <>{".".repeat(dots) + " ".repeat(3 - dots)}</>;
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
      [{isTransitional(status) ? <>{status}<AnimatedDots /></> : status}]
    </span>
  );
}
