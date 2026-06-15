ALTER TABLE recipes ADD COLUMN main_protein TEXT NOT NULL DEFAULT '';
ALTER TABLE recipes ADD COLUMN diet_class   TEXT NOT NULL DEFAULT '';
ALTER TABLE recipes ADD COLUMN batchable    INTEGER NOT NULL DEFAULT 0;
ALTER TABLE recipes ADD COLUMN cook_minutes INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_recipes_diet_class   ON recipes(diet_class);
CREATE INDEX IF NOT EXISTS idx_recipes_main_protein ON recipes(main_protein);
