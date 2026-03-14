import type { Task } from "../types";

interface Props {
  sessionId: string;
  currentTaskId: string;
  tasks: Task[];
  onSelect: (sessionId: string, targetTaskId: string) => void;
  onCancel: () => void;
}

export function MoveSessionModal({
  sessionId,
  currentTaskId,
  tasks,
  onSelect,
  onCancel,
}: Props) {
  const availableTasks = tasks.filter((t) => t.id !== currentTaskId);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{ backgroundColor: "rgba(0,0,0,0.7)" }}
    >
      <div
        className="w-full max-w-sm p-6"
        style={{
          backgroundColor: "#0A0A0A",
          border: "1px solid #2a2a2a",
          fontFamily: "'JetBrains Mono', monospace",
        }}
      >
        <h2
          className="text-sm font-bold lowercase mb-5"
          style={{ color: "#FAFAFA" }}
        >
          <span style={{ color: "#10B981" }}>{">"}</span> move_to_task
        </h2>

        <p
          className="text-[10px] mb-4"
          style={{ color: "#6B7280" }}
        >
          // select a target task
        </p>

        <div
          className="overflow-y-auto mb-5"
          style={{ maxHeight: 200 }}
        >
          {availableTasks.map((task) => (
            <button
              key={task.id}
              onClick={() => onSelect(sessionId, task.id)}
              className="w-full flex items-center gap-2 px-3 py-2 text-left text-xs transition-colors"
              style={{
                color: "#FAFAFA",
                fontFamily: "'JetBrains Mono', monospace",
              }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.backgroundColor = "#1F1F1F")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.backgroundColor = "transparent")
              }
            >
              <span style={{ color: "#10B981" }}>#</span>
              <span>
                {task.tag} {task.name}
              </span>
            </button>
          ))}
        </div>

        <div className="flex items-center justify-end">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-xs lowercase transition-colors"
            style={{ color: "#6B7280" }}
            onMouseEnter={(e) => (e.currentTarget.style.color = "#FAFAFA")}
            onMouseLeave={(e) => (e.currentTarget.style.color = "#6B7280")}
          >
            cancel
          </button>
        </div>
      </div>
    </div>
  );
}
