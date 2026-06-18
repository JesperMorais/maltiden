package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"maltiden/internal/domain"
	"strings"
	"time"
)

type LivsmedelStorage struct {
	db *sql.DB
}

func NewLivsmedelStorage(db *sql.DB) *LivsmedelStorage {
	return &LivsmedelStorage{db: db}
}

func scanLivsmedel(rows *sql.Rows) (domain.Livsmedel, error) {
	var lv domain.Livsmedel
	err := rows.Scan(&lv.Livsmedelsnummer, &lv.Namn, &lv.KcalPer100g, &lv.ProteinPer100g, &lv.CarbsPer100g, &lv.FatPer100g)
	return lv, err
}

// GetByNumbers returns the livsmedel rows for the given numbers, keyed by
// livsmedelsnummer. An empty input yields an empty (non-nil) map.
func (s *LivsmedelStorage) GetByNumbers(numbers []int) (map[int]domain.Livsmedel, error) {
	if len(numbers) == 0 {
		return make(map[int]domain.Livsmedel), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	placeholders := make([]string, len(numbers))
	args := make([]interface{}, len(numbers))
	for i, n := range numbers {
		placeholders[i] = "?"
		args[i] = n
	}

	query := fmt.Sprintf(`
		SELECT livsmedelsnummer, namn, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g
		FROM livsmedel WHERE livsmedelsnummer IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]domain.Livsmedel, len(numbers))
	for rows.Next() {
		lv, err := scanLivsmedel(rows)
		if err != nil {
			return nil, err
		}
		result[lv.Livsmedelsnummer] = lv
	}

	return result, rows.Err()
}

// GetAll returns every livsmedel row, ordered by livsmedelsnummer.
func (s *LivsmedelStorage) GetAll() ([]domain.Livsmedel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
		SELECT livsmedelsnummer, namn, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g
		FROM livsmedel ORDER BY livsmedelsnummer
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Livsmedel
	for rows.Next() {
		lv, err := scanLivsmedel(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, lv)
	}

	return result, rows.Err()
}

// Search returns livsmedel rows whose namn contains query (case-insensitive),
// capped at limit (defaulting to a small bound when limit <= 0).
func (s *LivsmedelStorage) Search(query string, limit int) ([]domain.Livsmedel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT livsmedelsnummer, namn, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g
		FROM livsmedel
		WHERE LOWER(namn) LIKE '%' || LOWER(?) || '%'
		ORDER BY LENGTH(namn), namn
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Livsmedel
	for rows.Next() {
		lv, err := scanLivsmedel(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, lv)
	}

	return result, rows.Err()
}
