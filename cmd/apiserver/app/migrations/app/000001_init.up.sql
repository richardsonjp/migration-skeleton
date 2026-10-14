-- Example migration for the "app" schema. Replace with your own.
-- search_path = app,public (set by the migrator per schema directory),
-- so unqualified names land in the app schema.

CREATE TABLE IF NOT EXISTS example (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
