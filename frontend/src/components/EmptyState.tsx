export function EmptyState() {
  return (
    <div
      className="flex items-center justify-center h-full"
      style={{ backgroundColor: "#0A0A0A" }}
    >
      <div
        className="text-center"
        style={{ fontFamily: "'JetBrains Mono', monospace" }}
      >
        <p className="text-3xl mb-3" style={{ color: "#10B981" }}>
          {">"}_
        </p>
        <p className="text-xs lowercase" style={{ color: "#4B5563" }}>
          select a session or create a new one
        </p>
      </div>
    </div>
  );
}
