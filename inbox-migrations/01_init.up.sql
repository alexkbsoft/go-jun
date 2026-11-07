CREATE TABLE tasks (
    id         TEXT        NOT NULL PRIMARY KEY,
    list_id    TEXT        NOT NULL,
    title      TEXT        NOT NULL,
    done       BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX tasks_list_id_idx ON tasks (list_id);
CREATE INDEX tasks_done_idx ON tasks (done);

CREATE TABLE inbox (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   TEXT        NOT NULL,
    aggregate_id TEXT        NOT NULL,
    json_payload JSONB       NOT NULL,
    occurred_at  TIMESTAMPTZ NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);
CREATE INDEX inbox_pending_idx ON inbox (created_at) WHERE status = 'pending';
CREATE UNIQUE INDEX inbox_event_unique_idx ON inbox (event_type, aggregate_id, occurred_at);