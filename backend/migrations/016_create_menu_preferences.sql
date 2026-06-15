-- Per-household menu generation preferences.
-- One row per household (PRIMARY KEY enforces the 1:1 relationship).
-- excluded_tags is a JSON array of recipe tags to hard-filter out of generation
-- (e.g. ["fisk", "fläsk"]). Scalar columns hold defaulting preferences applied
-- when a generate request omits them.
CREATE TABLE IF NOT EXISTS menu_preferences (
    household_id     TEXT PRIMARY KEY,
    excluded_tags    TEXT NOT NULL DEFAULT '[]',
    default_days     INTEGER NOT NULL DEFAULT 7,
    default_servings INTEGER NOT NULL DEFAULT 4,
    vegetarian_days  INTEGER NOT NULL DEFAULT 0,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE
);
