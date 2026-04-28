-- SQLite cannot ALTER TABLE to add a FOREIGN KEY, so we rebuild the table:
-- create new table with the household_id FK, copy data, drop old, rename.
CREATE TABLE custom_shopping_items_new (
    id TEXT PRIMARY KEY,
    menu_id TEXT NOT NULL,
    household_id TEXT NOT NULL,
    name TEXT NOT NULL,
    unit TEXT NOT NULL DEFAULT 'st',
    amount REAL NOT NULL DEFAULT 1,
    checked INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
    FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE
);

INSERT INTO custom_shopping_items_new (id, menu_id, household_id, name, unit, amount, checked, created_at)
SELECT id, menu_id, household_id, name, unit, amount, checked, created_at FROM custom_shopping_items;

DROP TABLE custom_shopping_items;
ALTER TABLE custom_shopping_items_new RENAME TO custom_shopping_items;

CREATE INDEX IF NOT EXISTS idx_custom_shopping_menu ON custom_shopping_items(menu_id);
CREATE INDEX IF NOT EXISTS idx_custom_shopping_household ON custom_shopping_items(household_id);
