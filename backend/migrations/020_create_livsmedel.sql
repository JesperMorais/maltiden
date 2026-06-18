CREATE TABLE livsmedel (
  livsmedelsnummer INTEGER PRIMARY KEY,
  namn             TEXT NOT NULL,
  kcal_per_100g    REAL NOT NULL DEFAULT 0,
  protein_per_100g REAL NOT NULL DEFAULT 0,
  carbs_per_100g   REAL NOT NULL DEFAULT 0,
  fat_per_100g     REAL NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_livsmedel_namn ON livsmedel(namn);
