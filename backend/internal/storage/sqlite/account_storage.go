package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AccountStorage implements the GDPR self-service operations: erasing a user
// (Art. 17) and exporting what the app stores about them (Art. 15/20).
type AccountStorage struct {
	db *sql.DB
}

func NewAccountStorage(db *sql.DB) *AccountStorage {
	return &AccountStorage{db: db}
}

// DeleteUser permanently removes a user and everything that is only theirs.
//
// Household data (menus, household recipes, shopping list, preferences) is
// shared by all members, so it is deleted only when the user is the last
// member. Otherwise the user just leaves, and if they owned the household the
// longest-standing remaining member (members before guests) becomes owner.
// Runs in one transaction; on error nothing is deleted.
func (s *AccountStorage) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	var householdID, role string
	err = tx.QueryRowContext(ctx,
		`SELECT household_id, role FROM household_members WHERE user_id = ?`, userID,
	).Scan(&householdID, &role)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("load membership: %w", err)
	}

	if householdID != "" {
		var others int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM household_members WHERE household_id = ? AND user_id <> ?`,
			householdID, userID,
		).Scan(&others); err != nil {
			return fmt.Errorf("count members: %w", err)
		}

		if others == 0 {
			if err := deleteHouseholdTx(ctx, tx, householdID); err != nil {
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx,
				`DELETE FROM household_members WHERE user_id = ?`, userID,
			); err != nil {
				return fmt.Errorf("leave household: %w", err)
			}
			if role == "owner" {
				if _, err := tx.ExecContext(ctx, `
					UPDATE household_members SET role = 'owner'
					WHERE id = (
						SELECT id FROM household_members
						WHERE household_id = ?
						ORDER BY CASE role WHEN 'member' THEN 0 ELSE 1 END, joined_at, id
						LIMIT 1
					)`, householdID,
				); err != nil {
					return fmt.Errorf("transfer ownership: %w", err)
				}
			}
		}
	}

	// feedback.user_id has no ON DELETE action, so remove it explicitly.
	// password_reset_tokens and household_members cascade; invite_codes.used_by
	// is set to NULL by its foreign key.
	if _, err := tx.ExecContext(ctx, `DELETE FROM feedback WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// deleteHouseholdTx removes a household and all data scoped to it. Order
// matters: menus.household_id and menu_days.recipe_id have no ON DELETE
// action, so menus and recipe references go before the household itself.
func deleteHouseholdTx(ctx context.Context, tx *sql.Tx, householdID string) error {
	steps := []struct {
		what  string
		query string
	}{
		// Another household's menu could reference one of these recipes;
		// detach rather than fail the whole deletion.
		{"detach recipes", `UPDATE menu_days SET recipe_id = NULL
			WHERE recipe_id IN (SELECT id FROM recipes WHERE household_id = ?)`},
		// Cascades to menu_days, shopping_items and custom_shopping_items.
		{"delete menus", `DELETE FROM menus WHERE household_id = ?`},
		{"delete recipes", `DELETE FROM recipes WHERE household_id = ?`},
		// Cascades to household_members, invite_codes, menu_preferences and
		// any remaining custom_shopping_items.
		{"delete household", `DELETE FROM households WHERE id = ?`},
	}
	for _, st := range steps {
		if _, err := tx.ExecContext(ctx, st.query, householdID); err != nil {
			return fmt.Errorf("%s: %w", st.what, err)
		}
	}
	return nil
}

// ExportUser returns everything stored about the user, plus the data of the
// household they belong to, as JSON-ready maps. The password hash, reset
// tokens and token version are deliberately left out: they are credentials,
// not personal data the user can use. Other members are listed by name and
// role only, which is what the app already shows every member.
func (s *AccountStorage) ExportUser(userID string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := queryMaps(ctx, s.db,
		`SELECT id, email, name, created_at FROM users WHERE id = ?`, userID)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, sql.ErrNoRows
	}

	out := map[string]any{
		"exportedAt": time.Now().UTC(),
		"user":       users[0],
	}

	feedback, err := queryMaps(ctx, s.db, `
		SELECT mood, categories, comment, page, viewport_width, user_agent, created_at
		FROM feedback WHERE user_id = ? ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	out["feedback"] = feedback

	membership, err := queryMaps(ctx, s.db, `
		SELECT h.id, h.name, h.created_at, hm.role, hm.joined_at,
		       hm.is_eating_today, hm.wants_lunch_box
		FROM household_members hm JOIN households h ON h.id = hm.household_id
		WHERE hm.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	if len(membership) == 0 {
		out["household"] = nil
		return out, nil
	}
	household := membership[0]
	householdID, _ := household["id"].(string)

	sections := []struct {
		key   string
		query string
	}{
		{"members", `SELECT u.name, hm.role FROM household_members hm
			JOIN users u ON u.id = hm.user_id
			WHERE hm.household_id = ? ORDER BY hm.joined_at`},
		{"menuPreferences", `SELECT excluded_tags, disliked_ingredients, default_days,
			default_servings, vegetarian_days, updated_at
			FROM menu_preferences WHERE household_id = ?`},
		{"recipes", `SELECT id, name, servings, emoji, tags, ingredients, instructions,
			calories, protein_g, carbs_g, fat_g, created_at
			FROM recipes WHERE household_id = ? ORDER BY created_at`},
		{"menus", `SELECT id, created_at, rationale FROM menus
			WHERE household_id = ? ORDER BY created_at`},
		{"menuDays", `SELECT d.menu_id, d.date, d.recipe_id, d.servings, d.skip,
			d.prep_mode, d.leftover_of
			FROM menu_days d JOIN menus m ON m.id = d.menu_id
			WHERE m.household_id = ? ORDER BY d.menu_id, d.date`},
		{"customShoppingItems", `SELECT menu_id, name, unit, amount, checked, created_at
			FROM custom_shopping_items WHERE household_id = ? ORDER BY created_at`},
	}
	for _, sec := range sections {
		rows, err := queryMaps(ctx, s.db, sec.query, householdID)
		if err != nil {
			return nil, fmt.Errorf("export %s: %w", sec.key, err)
		}
		household[sec.key] = rows
	}
	out["household"] = household
	return out, nil
}

// queryMaps runs a query and returns each row as a column→value map. TEXT
// columns holding JSON arrays/objects (tags, ingredients, ...) are embedded as
// JSON rather than as escaped strings so the export stays readable.
func queryMaps(ctx context.Context, db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := []map[string]any{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			row[c] = exportValue(vals[i])
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func exportValue(v any) any {
	var s string
	switch t := v.(type) {
	case []byte:
		s = string(t)
	case string:
		s = t
	default:
		return v
	}
	trimmed := strings.TrimSpace(s)
	if (strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{")) && json.Valid([]byte(trimmed)) {
		return json.RawMessage(trimmed)
	}
	return s
}
