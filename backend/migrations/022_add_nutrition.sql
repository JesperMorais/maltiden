ALTER TABLE menu_preferences ADD COLUMN nutrition_profile      TEXT NOT NULL DEFAULT '';
ALTER TABLE menu_preferences ADD COLUMN protein_target_per_day REAL NOT NULL DEFAULT 0;
ALTER TABLE recipes ADD COLUMN kcal_per_serving    REAL NOT NULL DEFAULT 0;
ALTER TABLE recipes ADD COLUMN protein_per_serving REAL NOT NULL DEFAULT 0;
ALTER TABLE recipes ADD COLUMN carbs_per_serving   REAL NOT NULL DEFAULT 0;
ALTER TABLE recipes ADD COLUMN fat_per_serving     REAL NOT NULL DEFAULT 0;
