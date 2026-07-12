CREATE TABLE IF NOT EXISTS dashboard_layouts (
    budget_id   INTEGER PRIMARY KEY REFERENCES budgets(id) ON DELETE CASCADE,
    version     INTEGER NOT NULL CHECK (version > 0),
    layout_json TEXT NOT NULL CHECK (json_valid(layout_json)),
    updated_at  TEXT NOT NULL
);