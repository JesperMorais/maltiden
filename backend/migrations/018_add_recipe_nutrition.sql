-- Add per-serving nutrition fields to recipes (issue #248, Phase 3 prereq).
-- Data-layer foundation only: this lays the columns the future nutrition
-- feature (menu-level macro balancing) will read/write. No aggregation logic
-- lives here.
--
-- Additive & non-destructive: four new nullable REAL columns, no rewrite of
-- existing rows. Columns are NULL by default so existing recipes (seed + user)
-- with no nutrition data stay valid and unchanged. Values are per serving.
ALTER TABLE recipes ADD COLUMN calories REAL;
ALTER TABLE recipes ADD COLUMN protein_g REAL;
ALTER TABLE recipes ADD COLUMN carbs_g REAL;
ALTER TABLE recipes ADD COLUMN fat_g REAL;
