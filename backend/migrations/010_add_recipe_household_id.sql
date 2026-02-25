-- Add household_id to recipes for ownership scoping.
-- NULL means seed/shared recipe (from migrations 004 and 008).
-- Non-NULL means the recipe belongs to a specific household.
ALTER TABLE recipes ADD COLUMN household_id TEXT;

CREATE INDEX IF NOT EXISTS idx_recipes_household ON recipes(household_id);
