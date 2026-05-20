package sqlite

import (
	"context"
	"database/sql"
	"maltiden/internal/domain"
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

		_, err = tx.ExecContext(ctx,
			`INSERT INTO menu_days (id, menu_id, date, recipe_id, servings, skip)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			"md_"+uuid.New().String(), menu.ID, day.Date, recipeID, day.Servings, skip,
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

		_, err = tx.ExecContext(ctx,
			`INSERT INTO menu_days (id, menu_id, date, recipe_id, servings, skip)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			"md_"+uuid.New().String(), menu.ID, day.Date, recipeID, day.Servings, skip,
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
		`SELECT date, recipe_id, servings, skip FROM menu_days
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

		err := rows.Scan(&day.Date, &recipeID, &day.Servings, &skip)
		if err != nil {
			return nil, err
		}

		if recipeID.Valid {
			day.RecipeID = recipeID.String
		}
		day.Skip = skip == 1

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
		`SELECT date, recipe_id, servings, skip FROM menu_days
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

		err := rows.Scan(&day.Date, &recipeID, &day.Servings, &skip)
		if err != nil {
			return nil, err
		}

		if recipeID.Valid {
			day.RecipeID = recipeID.String
		}
		day.Skip = skip == 1

		menu.Days = append(menu.Days, day)
	}

	if menu.Days == nil {
		menu.Days = []domain.MenuDay{}
	}

	return &menu, rows.Err()
}

// UpdateDayRecipe sets recipe_id for a single menu_day row identified by menu+date.
// Returns domain.ErrNotFound if no matching row exists.
func (s *MenuStorage) UpdateDayRecipe(menuID, date, recipeID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := s.db.ExecContext(ctx,
		`UPDATE menu_days SET recipe_id = ?, skip = 0 WHERE menu_id = ? AND date = ?`,
		recipeID, menuID, date,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
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
