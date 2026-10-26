import { useEffect, useState, type FormEvent } from "react";

type List = { id: string; name: string };
type ListsResponse = { lists: List[] | null };

export default function App() {
  const [lists, setLists] = useState<List[]>([]);
  const [name, setName] = useState("");
  const [error, setError] = useState("");

  async function load() {
    try {
      const res = await fetch("/v1/lists");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: ListsResponse = await res.json();
      // если списков нет, Go может отдать null вместо []
      setLists(data.lists ?? []);
      setError("");
    } catch (e) {
      setError(String(e));
    }
  }

  console.log("lists: ", lists);

  useEffect(() => {
    load();
  }, []);

  async function create(e: FormEvent) {
    e.preventDefault();
    const res = await fetch("/v1/lists", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    if (res.ok) {
      setName("");
      load();
    } else {
      setError(`HTTP ${res.status}`);
    }
  }

  return (
    <main
      style={{ maxWidth: 480, margin: "2rem auto", fontFamily: "sans-serif" }}
    >
      <h1>Списки</h1>

      <form onSubmit={create}>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Название списка"
        />
        <button type="submit" disabled={!name.trim()}>
          Создать
        </button>
      </form>

      {error && <p style={{ color: "crimson" }}>Ошибка: {error}</p>}

      <ul>
        {lists.map((l) => (
          <li key={l.id}>{l.name}</li>
        ))}
      </ul>
    </main>
  );
}
