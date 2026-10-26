import { useState, useRef, useEffect } from "react";
import gsap from "gsap";
import type { List } from "./types";
import ListsScreen from "./components/ListsScreen";
import TaskList from "./components/TaskList";

type View = "lists" | "tasks";

export default function App() {
  const [view, setView] = useState<View>("lists");
  const [selectedList, setSelectedList] = useState<List | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (panelRef.current) {
      gsap.fromTo(
        panelRef.current,
        { opacity: 0, x: view === "tasks" ? 30 : -30 },
        { opacity: 1, x: 0, duration: 0.35, ease: "power2.out" }
      );
    }
  }, [view, selectedList]);

  const handleOpenList = (list: List) => {
    setSelectedList(list);
    setView("tasks");
  };

  const handleBack = () => {
    setSelectedList(null);
    setView("lists");
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-200">
      <div ref={panelRef} className="transition-all duration-200">
        {view === "lists" ? (
          <ListsScreen onOpenList={handleOpenList} />
        ) : selectedList ? (
          <TaskList list={selectedList} onBack={handleBack} />
        ) : null}
      </div>
    </div>
  );
}
