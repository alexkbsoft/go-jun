import { useRef, useEffect } from "react";
import gsap from "gsap";
import type { List } from "../types";

interface ListViewProps {
  list: List;
  onClick: () => void;
  onDelete: (id: string) => void;
}

export default function ListView({ list, onClick, onDelete }: ListViewProps) {
  const cardRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (cardRef.current) {
      gsap.fromTo(
        cardRef.current,
        { opacity: 0, scale: 0.95 },
        { opacity: 1, scale: 1, duration: 0.3, ease: "back.out(1.4)" }
      );
    }
  }, []);

  return (
    <div
      ref={cardRef}
      className="group relative bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl p-5 cursor-pointer hover:shadow-lg hover:border-purple-300 dark:hover:border-purple-600 transition-all duration-200 hover:-translate-y-0.5"
      onClick={onClick}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 text-left">
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 group-hover:text-purple-600 dark:group-hover:text-purple-400 transition-colors">
            {list.name}
          </h3>
        </div>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(list.id);
          }}
          className="ml-3 p-1.5 text-gray-400 hover:text-red-500 dark:hover:text-red-400 opacity-0 group-hover:opacity-100 transition-all duration-200 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20"
          title="Удалить список"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="w-4 h-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
            />
          </svg>
        </button>
      </div>
      <div className="mt-3 flex items-center text-sm text-gray-500 dark:text-gray-400">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-4 h-4 mr-1.5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
          />
        </svg>
        <span>Нажмите для просмотра задач</span>
      </div>
    </div>
  );
}
