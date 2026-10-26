import type { List, Task, ListsResponse, TasksResponse } from "./types";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export const api = {
  async getLists(): Promise<List[]> {
    const data: ListsResponse = await request("/v1/lists");
    return data.lists ?? [];
  },

  async createList(name: string): Promise<List> {
    return request<List>("/v1/lists", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
  },

  async getTasks(listId: string): Promise<Task[]> {
    const data: TasksResponse = await request(`/v1/lists/${listId}/tasks`);
    return data.tasks ?? [];
  },

  async createTask(listId: string, title: string): Promise<Task> {
    return request<Task>(`/v1/lists/${listId}/tasks`, {
      method: "POST",
      body: JSON.stringify({ title }),
    });
  },

  async updateTask(taskId: string, patch: { title?: string; done?: boolean }): Promise<Task> {
    return request<Task>(`/v1/tasks/${taskId}`, {
      method: "PATCH",
      body: JSON.stringify(patch),
    });
  },

  async deleteTask(taskId: string): Promise<void> {
    await request(`/v1/tasks/${taskId}`, { method: "DELETE" });
  },
};
