CREATE TABLE IF NOT EXISTS custom_shopping_items (
    id TEXT PRIMARY KEY,
    menu_id TEXT NOT NULL,
    household_id TEXT NOT NULL,
    name TEXT NOT NULL,
    unit TEXT NOT NULL DEFAULT 'st',
    amount REAL NOT NULL DEFAULT 1,
    checked INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_custom_shopping_menu ON custom_shopping_items(menu_id);
