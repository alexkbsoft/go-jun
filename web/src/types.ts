export interface List {
  id: string;
  name: string;
}

export interface Task {
  id: string;
  listId: string;
  title: string;
  done: boolean;
  createdAt: string;
}

export interface ListsResponse {
  lists: List[] | null;
}

export interface TasksResponse {
  tasks: Task[] | null;
}
