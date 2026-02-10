package sqlite

import (
	"context"
	"database/sql"
	"time"
)

type ShoppingStorage struct {
	db *sql.DB
}

func NewShoppingStorage(db *sql.DB) *ShoppingStorage {
	return &ShoppingStorage{db: db}
}

func (s *ShoppingStorage) IsChecked(menuID, ingredientName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var checked int
	err := s.db.QueryRowContext(ctx,
		`SELECT checked FROM shopping_items WHERE menu_id = ? AND ingredient_name = ?`,
		menuID, ingredientName,
	).Scan(&checked)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return checked == 1, nil
}

func (s *ShoppingStorage) SetChecked(menuID, itemID string, checked bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkedInt := 0
	if checked {
		checkedInt = 1
	}

	// Upsert - insert or update
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO shopping_items (id, menu_id, ingredient_name, checked)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET checked = ?
	`, itemID, menuID, itemID, checkedInt, checkedInt)

	return err
}

func (s *ShoppingStorage) GetCheckedItems(menuID string) (map[string]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx,
		`SELECT ingredient_name, checked FROM shopping_items WHERE menu_id = ?`,
		menuID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var name string
		var checked int
		if err := rows.Scan(&name, &checked); err != nil {
			return nil, err
		}
		result[name] = checked == 1
	}

	return result, rows.Err()
}
