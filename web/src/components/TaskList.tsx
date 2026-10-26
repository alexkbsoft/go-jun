import { useRef, useState, useCallback, useEffect } from "react";
import gsap from "gsap";
import type { List, Task } from "../types";
import { api } from "../api";
import TaskItem from "./TaskItem";

interface TaskListProps {
  list: List;
  onBack: () => void;
}

export default function TaskList({ list, onBack }: TaskListProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTitle, setNewTitle] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);

  const loadTasks = useCallback(async () => {
    try {
      const data = await api.getTasks(list.id);
      setTasks(data);
      setError("");
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, [list.id]);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  useEffect(() => {
    if (containerRef.current) {
      gsap.fromTo(
        containerRef.current,
        { opacity: 0, y: 30 },
        { opacity: 1, y: 0, duration: 0.4, ease: "power2.out" }
      );
    }
  }, []);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim()) return;
    try {
      const task = await api.createTask(list.id, newTitle.trim());
      setTasks((prev) => [...prev, task]);
      setNewTitle("");
      setError("");
    } catch (e) {
      setError(String(e));
    }
  };

  const handleToggle = async (id: string, done: boolean) => {
    setTasks((prev) =>
      prev.map((t) => (t.id === id ? { ...t, done } : t))
    );
    try {
      await api.updateTask(id, { done });
    } catch (e) {
      setError(String(e));
      loadTasks();
    }
  };

  const handleEdit = async (id: string, title: string) => {
    setTasks((prev) =>
      prev.map((t) => (t.id === id ? { ...t, title } : t))
    );
    try {
      await api.updateTask(id, { title });
    } catch (e) {
      setError(String(e));
      loadTasks();
    }
  };

  const handleDelete = async (id: string) => {
    setTasks((prev) => prev.filter((t) => t.id !== id));
    try {
      await api.deleteTask(id);
    } catch (e) {
      setError(String(e));
      loadTasks();
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin w-8 h-8 border-4 border-purple-500 border-t-transparent rounded-full" />
      </div>
    );
  }

  const doneCount = tasks.filter((t) => t.done).length;
  const totalCount = tasks.length;

  return (
    <div ref={containerRef} className="max-w-2xl mx-auto px-4 py-12">
      {/* Header */}
      <div className="flex items-center gap-4 mb-8">
        <button
          onClick={onBack}
          className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          title="Назад к спискам"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
        <div className="flex-1 text-left">
          <h1 className="text-3xl font-medium text-gray-900 dark:text-gray-100">
            {list.name}
          </h1>
          {totalCount > 0 && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {doneCount} из {totalCount} {doneCount === 1 ? "выполнено" : doneCount < 5 ? "выполнено" : "выполнено"}
            </p>
          )}
        </div>
      </div>

      {/* Progress bar */}
      {totalCount > 0 && (
        <div className="mb-6 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-purple-500 to-green-500 rounded-full transition-all duration-500 ease-out"
            style={{ width: `${(doneCount / totalCount) * 100}%` }}
          />
        </div>
      )}

      {/* Add task form */}
      <form onSubmit={handleAdd} className="flex gap-3 mb-8">
        <input
          type="text"
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          placeholder="Новая задача"
          className="flex-1 px-4 py-2.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
        />
        <button
          type="submit"
          disabled={!newTitle.trim()}
          className="px-6 py-2.5 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-medium rounded-lg transition-colors duration-200 shadow-sm hover:shadow-md"
        >
          Добавить
        </button>
      </form>

      {error && (
        <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-600 dark:text-red-400">
          Ошибка: {error}
        </div>
      )}

      {/* Task list */}
      {tasks.length === 0 ? (
        <div className="text-center py-16 text-gray-500 dark:text-gray-400">
          <p className="text-lg">Задач пока нет</p>
          <p className="text-sm mt-2">Добавьте первую задачу</p>
        </div>
      ) : (
        <ul className="space-y-3">
          {tasks.map((task, index) => (
            <TaskItem
              key={task.id}
              task={task}
              index={index}
              onToggle={handleToggle}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          ))}
        </ul>
      )}
    </div>
  );
}
