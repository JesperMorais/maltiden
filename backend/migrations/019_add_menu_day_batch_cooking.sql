-- Add batch-cooking markers to menu days (issue #248 Phase 2).
-- Additive: two new TEXT columns on menu_days, both defaulting to '' so every
-- existing menu day keeps its current (non-batch) behavior.
--   prep_mode   = '' for an ordinary day, 'batch' for a cook-once-eat-twice
--                 cook-day (double servings, reused later in the week).
--   leftover_of = the cook-day's date a leftovers day reuses; '' otherwise.
ALTER TABLE menu_days
    ADD COLUMN prep_mode TEXT NOT NULL DEFAULT '';
ALTER TABLE menu_days
    ADD COLUMN leftover_of TEXT NOT NULL DEFAULT '';
