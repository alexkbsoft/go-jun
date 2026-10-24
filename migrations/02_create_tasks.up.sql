BEGIN;

CREATE TABLE tasks (
    id         TEXT        NOT NULL PRIMARY KEY,
    list_id    TEXT        NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    title      TEXT        NOT NULL,
    done       BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX tasks_list_id_idx ON tasks (list_id);

COMMIT;