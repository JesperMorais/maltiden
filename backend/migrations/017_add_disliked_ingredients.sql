-- Add per-household disliked ingredients to menu preferences (issue #248).
-- Additive: a new JSON-array column, mirroring how excluded_tags is stored.
-- Any recipe containing a disliked ingredient (name-normalized match) is
-- hard-filtered out of menu generation. Defaults to an empty array so existing
-- rows keep their current generation behavior.
ALTER TABLE menu_preferences
    ADD COLUMN disliked_ingredients TEXT NOT NULL DEFAULT '[]';
