package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"maltiden/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MenuStorage struct {
	db *sql.DB
}

func NewMenuStorage(db *sql.DB) *MenuStorage {
	return &MenuStorage{db: db}
}

func (s *MenuStorage) Create(menu *domain.Menu) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert menu
	_, err = tx.ExecContext(ctx,
		`INSERT INTO menus (id, household_id, created_at) VALUES (?, ?, ?)`,
		menu.ID, menu.HouseholdID, menu.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Insert menu days
	for _, day := range menu.Days {
		skip := 0
		if day.Skip {
			skip = 1
		}

		var recipeID *string
		if day.RecipeID != "" {
			recipeID = &day.RecipeID
		}

		leftover := 0
		if day.Leftover {
			leftover = 1
		}
		var cookDate *string
		if day.CookDate != "" {
			cookDate = &day.CookDate
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO menu_days (id, menu_id, date, recipe_id, servings, skip, leftover, cook_date)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"md_"+uuid.New().String(), menu.ID, day.Date, recipeID, day.Servings, skip, leftover, cookDate,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *MenuStorage) Update(menu *domain.Menu) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing menu days
	_, err = tx.ExecContext(ctx, `DELETE FROM menu_days WHERE menu_id = ?`, menu.ID)
	if err != nil {
		return err
	}

	// Insert new menu days
	for _, day := range menu.Days {
		skip := 0
		if day.Skip {
			skip = 1
		}

		var recipeID *string
		if day.RecipeID != "" {
			recipeID = &day.RecipeID
		}

		leftover := 0
		if day.Leftover {
			leftover = 1
		}
		var cookDate *string
		if day.CookDate != "" {
			cookDate = &day.CookDate
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO menu_days (id, menu_id, date, recipe_id, servings, skip, leftover, cook_date)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"md_"+uuid.New().String(), menu.ID, day.Date, recipeID, day.Servings, skip, leftover, cookDate,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *MenuStorage) GetCurrentByHousehold(householdID string) (*domain.Menu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the most recent menu for the household
	var menu domain.Menu
	err := s.db.QueryRowContext(ctx,
		`SELECT id, household_id, created_at FROM menus
		 WHERE household_id = ? ORDER BY created_at DESC LIMIT 1`,
		householdID,
	).Scan(&menu.ID, &menu.HouseholdID, &menu.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get menu days
	rows, err := s.db.QueryContext(ctx,
		`SELECT date, recipe_id, servings, skip, leftover, cook_date FROM menu_days
		 WHERE menu_id = ? ORDER BY date`,
		menu.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day domain.MenuDay
		var recipeID sql.NullString
		var skip int
		var leftover int
		var cookDate sql.NullString

		err := rows.Scan(&day.Date, &recipeID, &day.Servings, &skip, &leftover, &cookDate)
		if err != nil {
			return nil, err
		}

		if recipeID.Valid {
			day.RecipeID = recipeID.String
		}
		day.Skip = skip == 1
		day.Leftover = leftover == 1
		if cookDate.Valid {
			day.CookDate = cookDate.String
		}

		menu.Days = append(menu.Days, day)
	}

	if menu.Days == nil {
		menu.Days = []domain.MenuDay{}
	}

	return &menu, rows.Err()
}

func (s *MenuStorage) GetByID(id string) (*domain.Menu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var menu domain.Menu
	err := s.db.QueryRowContext(ctx,
		`SELECT id, household_id, created_at FROM menus WHERE id = ?`,
		id,
	).Scan(&menu.ID, &menu.HouseholdID, &menu.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get menu days
	rows, err := s.db.QueryContext(ctx,
		`SELECT date, recipe_id, servings, skip, leftover, cook_date FROM menu_days
		 WHERE menu_id = ? ORDER BY date`,
		menu.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day domain.MenuDay
		var recipeID sql.NullString
		var skip int
		var leftover int
		var cookDate sql.NullString

		err := rows.Scan(&day.Date, &recipeID, &day.Servings, &skip, &leftover, &cookDate)
		if err != nil {
			return nil, err
		}

		if recipeID.Valid {
			day.RecipeID = recipeID.String
		}
		day.Skip = skip == 1
		day.Leftover = leftover == 1
		if cookDate.Valid {
			day.CookDate = cookDate.String
		}

		menu.Days = append(menu.Days, day)
	}

	if menu.Days == nil {
		menu.Days = []domain.MenuDay{}
	}

	return &menu, rows.Err()
}

// GetRecentRecipeIDs returns the set of recipe ids used across a household's
// most recent windowMenus menus. It is used by the menu generator's recency
// penalty to deprioritize recently-cooked dishes. Both queries hit existing
// indexes (idx_menus_household, idx_menu_days_menu); no migration required.
func (s *MenuStorage) GetRecentRecipeIDs(householdID string, windowMenus int) (map[string]bool, error) {
	ids := make(map[string]bool)
	if windowMenus <= 0 {
		return ids, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Most recent menus for the household. The secondary "id DESC" key makes the
	// ordering deterministic when several menus share the same created_at second
	// (CURRENT_TIMESTAMP has 1s resolution), so the recency window is stable.
	rows, err := s.db.QueryContext(ctx,
		`SELECT id FROM menus WHERE household_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`,
		householdID, windowMenus,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		menuIDs = append(menuIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(menuIDs) == 0 {
		return ids, nil
	}

	// Recipe ids across those menus' days.
	placeholders := make([]string, len(menuIDs))
	args := make([]interface{}, len(menuIDs))
	for i, id := range menuIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf(
		`SELECT recipe_id FROM menu_days WHERE menu_id IN (%s) AND recipe_id IS NOT NULL`,
		strings.Join(placeholders, ","),
	)

	dayRows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer dayRows.Close()

	for dayRows.Next() {
		var recipeID string
		if err := dayRows.Scan(&recipeID); err != nil {
			return nil, err
		}
		ids[recipeID] = true
	}

	return ids, dayRows.Err()
}

// GetHouseholdIDByMenuID returns the household_id for a given menu_id.
// Returns empty string if menu not found.
func (s *MenuStorage) GetHouseholdIDByMenuID(menuID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var householdID string
	err := s.db.QueryRowContext(ctx,
		`SELECT household_id FROM menus WHERE id = ?`,
		menuID,
	).Scan(&householdID)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return householdID, nil
}
