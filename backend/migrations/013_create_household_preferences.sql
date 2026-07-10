CREATE TABLE IF NOT EXISTS household_preferences (
    household_id TEXT PRIMARY KEY,
    diet_profile TEXT,
    vegetarian_days_per_week INTEGER,
    disliked_ingredients TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (household_id) REFERENCES households(id)
);
