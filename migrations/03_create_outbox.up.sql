BEGIN;

CREATE TABLE outbox (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type  TEXT       NOT NULL,
    aggregate_id TEXT      NOT NULL,   -- ID агрегата (task ID)
    json_payload JSONB      NOT NULL,  -- сериализованное событие
    status      TEXT       NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE INDEX outbox_status_idx ON outbox (status, created_at) WHERE status = 'pending';

COMMIT;