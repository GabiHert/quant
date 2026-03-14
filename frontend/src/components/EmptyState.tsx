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
        <div className="mt-6 text-xs text-left max-w-md mx-auto space-y-2" style={{ color: "#6B7280" }}>
          <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
          <p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.</p>
          <p>Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.</p>
          <p>Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p>
          <p>The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. How vexingly quick daft zebras jump.</p>
          <p>Buffalo buffalo Buffalo buffalo buffalo buffalo Buffalo buffalo. James while John had had had had had had had had had had had a better effect on the teacher.</p>
          <p>Sphinx of black quartz, judge my vow. Two driven jocks help fax my big quiz. The five boxing wizards jump quickly.</p>
        </div>
      </div>
    </div>
  );
}
