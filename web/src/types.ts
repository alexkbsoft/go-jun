export interface List {
  id: string;
  name: string;
}

export interface Task {
  id: string;
  list_id: string;
  title: string;
  done: boolean;
  created_at: string;
}

export interface ListsResponse {
  lists: List[] | null;
}

export interface TasksResponse {
  tasks: Task[] | null;
}
