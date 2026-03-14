export function EmptyState() {
  return (
    <div
      className="flex items-center justify-center h-full"
      style={{ backgroundColor: "#0A0A0A" }}
    >
      <div
        className="text-center max-w-lg"
        style={{ fontFamily: "'JetBrains Mono', monospace" }}
      >
        <p className="text-3xl mb-3" style={{ color: "#10B981" }}>
          {">"}_ quant
        </p>
        <p className="text-sm mb-8" style={{ color: "#6B7280" }}>
          multiple agents. one dashboard. zero chaos.
        </p>

        <div
          className="text-xs text-left space-y-4"
          style={{ color: "#4B5563" }}
        >
          <div>
            <p style={{ color: "#6B7280" }}>// getting started</p>
            <p>
              <span style={{ color: "#10B981" }}>1.</span> click{" "}
              <span style={{ color: "#FAFAFA" }}>+ repo</span> to open a git
              repository
            </p>
            <p>
              <span style={{ color: "#10B981" }}>2.</span> create a{" "}
              <span style={{ color: "#FAFAFA" }}># task</span> to organize your
              work (e.g. PLT-123)
            </p>
            <p>
              <span style={{ color: "#10B981" }}>3.</span> add a{" "}
              <span style={{ color: "#FAFAFA" }}>session</span> under the task
              to start a claude code agent
            </p>
          </div>

          <div>
            <p style={{ color: "#6B7280" }}>// features</p>
            <p>
              <span style={{ color: "#10B981" }}>$</span> run multiple claude
              code sessions in parallel
            </p>
            <p>
              <span style={{ color: "#10B981" }}>$</span> sessions persist
              across app restarts
            </p>
            <p>
              <span style={{ color: "#10B981" }}>$</span> optional git worktrees
              for branch isolation
            </p>
            <p>
              <span style={{ color: "#10B981" }}>$</span> right-click for
              context menus on repos, tasks, sessions
            </p>
          </div>

          <div>
            <p style={{ color: "#6B7280" }}>// tips</p>
            <p>
              <span style={{ color: "#F59E0B" }}>$</span> check "skip
              permissions" to run with --dangerously-skip-permissions
            </p>
            <p>
              <span style={{ color: "#F59E0B" }}>$</span> check "use worktree"
              to isolate work on a new git branch
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
