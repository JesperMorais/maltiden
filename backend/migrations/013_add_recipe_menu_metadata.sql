-- Add smart-menu metadata columns to recipes.
-- NULL/0 means metadata not yet backfilled (see follow-up backfill job).
ALTER TABLE recipes ADD COLUMN main_protein TEXT;
ALTER TABLE recipes ADD COLUMN diet_class TEXT;
ALTER TABLE recipes ADD COLUMN batchable INTEGER;
ALTER TABLE recipes ADD COLUMN cook_minutes INTEGER;

CREATE INDEX IF NOT EXISTS idx_recipes_diet_class ON recipes(diet_class);
