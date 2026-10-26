import { useRef, useEffect, useState } from "react";
import gsap from "gsap";
import type { Task } from "../types";

interface TaskItemProps {
  task: Task;
  onToggle: (id: string, done: boolean) => void;
  onEdit: (id: string, title: string) => void;
  onDelete: (id: string) => void;
  index: number;
}

export default function TaskItem({ task, onToggle, onEdit, onDelete, index }: TaskItemProps) {
  const rowRef = useRef<HTMLLIElement>(null);
  const [editing, setEditing] = useState(false);
  const [editTitle, setEditTitle] = useState(task.title);

  useEffect(() => {
    if (rowRef.current) {
      gsap.fromTo(
        rowRef.current,
        { opacity: 0, x: -20 },
        { opacity: 1, x: 0, duration: 0.3, delay: index * 0.05, ease: "power2.out" }
      );
    }
  }, [index]);

  const handleSave = () => {
    if (editTitle.trim() && editTitle.trim() !== task.title) {
      onEdit(task.id, editTitle.trim());
    } else {
      setEditTitle(task.title);
    }
    setEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSave();
    if (e.key === "Escape") {
      setEditTitle(task.title);
      setEditing(false);
    }
  };

  return (
    <li
      ref={rowRef}
      className={`group flex items-center gap-3 p-4 rounded-lg transition-all duration-200 ${
        task.done
          ? "bg-gray-50 dark:bg-gray-800/50"
          : "bg-white dark:bg-gray-800 hover:shadow-md"
      } border border-gray-200 dark:border-gray-700`}
    >
      {/* Checkbox */}
      <button
        onClick={() => onToggle(task.id, !task.done)}
        className={`flex-shrink-0 w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all duration-200 ${
          task.done
            ? "bg-green-500 border-green-500 text-white"
            : "border-gray-300 dark:border-gray-600 hover:border-purple-500 dark:hover:border-purple-400"
        }`}
        title={task.done ? "Отметить как активную" : "Отметить как выполненную"}
      >
        {task.done && (
          <svg xmlns="http://www.w3.org/2000/svg" className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        )}
      </button>

      {/* Title / Edit */}
      <div className="flex-1 min-w-0">
        {editing ? (
          <input
            type="text"
            value={editTitle}
            onChange={(e) => setEditTitle(e.target.value)}
            onBlur={handleSave}
            onKeyDown={handleKeyDown}
            className="w-full px-3 py-1.5 rounded-md border border-purple-400 dark:border-purple-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-purple-500"
            autoFocus
          />
        ) : (
          <span
            onDoubleClick={() => {
              setEditing(true);
              setEditTitle(task.title);
            }}
            className={`block text-base cursor-default select-none transition-all duration-200 ${
              task.done
                ? "line-through text-gray-400 dark:text-gray-500"
                : "text-gray-900 dark:text-gray-100"
            }`}
            title="Дважды щёлкните для редактирования"
          >
            {task.title}
          </span>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        {!editing && (
          <>
            <button
              onClick={() => {
                setEditing(true);
                setEditTitle(task.title);
              }}
              className="p-1.5 text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
              title="Редактировать"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>
            <button
              onClick={() => onDelete(task.id)}
              className="p-1.5 text-gray-400 hover:text-red-500 dark:hover:text-red-400 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
              title="Удалить задачу"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </>
        )}
      </div>
    </li>
  );
}
