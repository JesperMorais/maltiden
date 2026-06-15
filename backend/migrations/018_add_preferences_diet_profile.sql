ALTER TABLE menu_preferences ADD COLUMN diet_profile         TEXT NOT NULL DEFAULT '';
ALTER TABLE menu_preferences ADD COLUMN disliked_ingredients TEXT NOT NULL DEFAULT '[]';
