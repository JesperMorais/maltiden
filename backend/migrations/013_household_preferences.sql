-- Dietary preferences for households (diet profile, veg days, dislikes)
ALTER TABLE households ADD COLUMN preferences TEXT NOT NULL DEFAULT '{}';
