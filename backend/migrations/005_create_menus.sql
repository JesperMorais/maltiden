CREATE TABLE IF NOT EXISTS menus (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (household_id) REFERENCES households(id)
);

CREATE TABLE IF NOT EXISTS menu_days (
    id TEXT PRIMARY KEY,
    menu_id TEXT NOT NULL,
    date TEXT NOT NULL,
    recipe_id TEXT,
    servings INTEGER NOT NULL DEFAULT 4,
    skip INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE INDEX IF NOT EXISTS idx_menus_household ON menus(household_id);
CREATE INDEX IF NOT EXISTS idx_menu_days_menu ON menu_days(menu_id);

-- Track checked items in shopping list
CREATE TABLE IF NOT EXISTS shopping_items (
    id TEXT PRIMARY KEY,
    menu_id TEXT NOT NULL,
    ingredient_name TEXT NOT NULL,
    checked INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_shopping_items_menu ON shopping_items(menu_id);
