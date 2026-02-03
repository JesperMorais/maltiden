-- Invite codes for household joining
CREATE TABLE IF NOT EXISTS invite_codes (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    used_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE,
    FOREIGN KEY (used_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_invite_codes_code ON invite_codes(code);
CREATE INDEX IF NOT EXISTS idx_invite_codes_household_id ON invite_codes(household_id);

-- Member status columns for eating today / lunch box tracking
ALTER TABLE household_members ADD COLUMN is_eating_today INTEGER NOT NULL DEFAULT 1;
ALTER TABLE household_members ADD COLUMN wants_lunch_box INTEGER NOT NULL DEFAULT 0;
