import { StatusDot } from "./StatusDot";
import type { DisplayStatus } from "./StatusBadge";

interface Tab {
  id: string;
  name: string;
  displayStatus: DisplayStatus;
}

interface TabBarProps {
  tabs: Tab[];
  activeTabId: string | null;
  onSelectTab: (id: string) => void;
  onCloseTab: (id: string) => void;
}

export function TabBar({ tabs, activeTabId, onSelectTab, onCloseTab }: TabBarProps) {
  if (tabs.length === 0) return null;

  return (
    <div
      className="flex items-center shrink-0 overflow-x-auto"
      style={{
        backgroundColor: "#0A0A0A",
        borderBottom: "1px solid #2a2a2a",
        fontFamily: "'JetBrains Mono', monospace",
      }}
    >
      {tabs.map((tab) => {
        const isActive = tab.id === activeTabId;
        return (
          <div
            key={tab.id}
            className="flex items-center gap-1.5 px-3 py-2 shrink-0 cursor-pointer"
            style={{
              backgroundColor: isActive ? "#1F1F1F" : "transparent",
              borderRight: "1px solid #2a2a2a",
              maxWidth: 200,
            }}
            onClick={() => onSelectTab(tab.id)}
            onMouseEnter={(e) => {
              if (!isActive) e.currentTarget.style.backgroundColor = "#1F1F1F";
            }}
            onMouseLeave={(e) => {
              if (!isActive) e.currentTarget.style.backgroundColor = "transparent";
            }}
          >
            <StatusDot status={tab.displayStatus} />
            <span
              className="text-xs overflow-hidden whitespace-nowrap flex-1"
              style={{
                color: isActive ? "#FAFAFA" : "#6B7280",
                textOverflow: "ellipsis",
              }}
            >
              {tab.name}
            </span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onCloseTab(tab.id);
              }}
              className="shrink-0 ml-1 text-[10px] transition-colors"
              style={{ color: "#4B5563" }}
              onMouseEnter={(e) => (e.currentTarget.style.color = "#FAFAFA")}
              onMouseLeave={(e) => (e.currentTarget.style.color = "#4B5563")}
              title="close tab"
            >
              [x]
            </button>
          </div>
        );
      })}
    </div>
  );
}
