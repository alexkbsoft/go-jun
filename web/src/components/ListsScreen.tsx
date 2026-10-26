import { useRef, useState, useEffect, useCallback } from "react";
import gsap from "gsap";
import type { List } from "../types";
import { api } from "../api";
import ListView from "./ListView";

interface ListsScreenProps {
  onOpenList: (list: List) => void;
}

export default function ListsScreen({ onOpenList }: ListsScreenProps) {
  const [lists, setLists] = useState<List[]>([]);
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);

  const load = useCallback(async () => {
    try {
      const data = await api.getLists();
      setLists(data);
      setError("");
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (containerRef.current) {
      gsap.fromTo(
        containerRef.current.children,
        { opacity: 0, y: 20 },
        { opacity: 1, y: 0, duration: 0.4, stagger: 0.05, ease: "power2.out" }
      );
    }
  }, [lists]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      const list = await api.createList(name.trim());
      setLists((prev) => [...prev, list]);
      setName("");
      setError("");
    } catch (e) {
      setError(String(e));
    }
  };

  const handleDelete = async (id: string) => {
    setLists((prev) => prev.filter((l) => l.id !== id));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin w-8 h-8 border-4 border-purple-500 border-t-transparent rounded-full" />
      </div>
    );
  }

  return (
    <div ref={containerRef} className="max-w-2xl mx-auto px-4 py-12">
      <h1 className="text-4xl font-medium text-gray-900 dark:text-gray-100 mb-8">
        Мои списки
      </h1>

      <form onSubmit={handleCreate} className="flex gap-3 mb-8">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Название нового списка"
          className="flex-1 px-4 py-2.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
        />
        <button
          type="submit"
          disabled={!name.trim()}
          className="px-6 py-2.5 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-medium rounded-lg transition-colors duration-200 shadow-sm hover:shadow-md"
        >
          Создать
        </button>
      </form>

      {error && (
        <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-600 dark:text-red-400">
          Ошибка: {error}
        </div>
      )}

      {lists.length === 0 ? (
        <div className="text-center py-16 text-gray-500 dark:text-gray-400">
          <p className="text-lg">Списков пока нет</p>
          <p className="text-sm mt-2">Создайте первый список дел</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {lists.map((list) => (
            <ListView
              key={list.id}
              list={list}
              onClick={() => onOpenList(list)}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}
