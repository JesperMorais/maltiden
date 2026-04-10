package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"maltiden/internal/domain"
	"strings"
	"time"
)

type RecipeStorage struct {
	db *sql.DB
}

func NewRecipeStorage(db *sql.DB) *RecipeStorage {
	return &RecipeStorage{db: db}
}

func (s *RecipeStorage) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, servings, emoji, tags FROM recipes WHERE 1=1`
	args := []interface{}{}

	// Scope by household: own recipes + seed recipes (NULL household_id)
	if householdID != "" {
		query += ` AND (household_id = ? OR household_id IS NULL)`
		args = append(args, householdID)
	} else {
		query += ` AND household_id IS NULL`
	}

	// Add name filter (case-insensitive partial match)
	if filter != nil && filter.Name != "" {
		query += ` AND LOWER(name) LIKE LOWER(?)`
		args = append(args, "%"+filter.Name+"%")
	}

	// Add tag filter (JSON search)
	if filter != nil && filter.Tag != "" {
		query += ` AND id IN (SELECT r2.id FROM recipes r2, json_each(r2.tags) WHERE json_each.value = ?)`
		args = append(args, filter.Tag)
	}

	query += ` ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []domain.RecipeSummary
	for rows.Next() {
		var r domain.RecipeSummary
		var emoji sql.NullString
		var tagsJSON string

		err := rows.Scan(&r.ID, &r.Name, &r.Servings, &emoji, &tagsJSON)
		if err != nil {
			return nil, err
		}

		if emoji.Valid {
			r.Emoji = emoji.String
		}

		if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
			r.Tags = []string{}
		}

		recipes = append(recipes, r)
	}

	if recipes == nil {
		recipes = []domain.RecipeSummary{}
	}

	return recipes, rows.Err()
}

func (s *RecipeStorage) GetByID(id string) (*domain.Recipe, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, servings, emoji, tags, ingredients, instructions, household_id, created_at
			  FROM recipes WHERE id = ?`

	var r domain.Recipe
	var emoji, householdID sql.NullString
	var tagsJSON, ingredientsJSON, instructionsJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.Name, &r.Servings, &emoji,
		&tagsJSON, &ingredientsJSON, &instructionsJSON, &householdID, &r.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if emoji.Valid {
		r.Emoji = emoji.String
	}
	if householdID.Valid {
		r.HouseholdID = householdID.String
	}

	if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
		r.Tags = []string{}
	}
	if err := json.Unmarshal([]byte(ingredientsJSON), &r.Ingredients); err != nil {
		r.Ingredients = []domain.Ingredient{}
	}
	if err := json.Unmarshal([]byte(instructionsJSON), &r.Instructions); err != nil {
		r.Instructions = []string{}
	}

	return &r, nil
}

func (s *RecipeStorage) GetByIDs(ids []string) (map[string]*domain.Recipe, error) {
	if len(ids) == 0 {
		return make(map[string]*domain.Recipe), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Build query with correct number of placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, servings, emoji, tags, ingredients, instructions, created_at
		FROM recipes WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*domain.Recipe)
	for rows.Next() {
		var r domain.Recipe
		var emoji sql.NullString
		var tagsJSON, ingredientsJSON, instructionsJSON string

		err := rows.Scan(
			&r.ID, &r.Name, &r.Servings, &emoji,
			&tagsJSON, &ingredientsJSON, &instructionsJSON, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if emoji.Valid {
			r.Emoji = emoji.String
		}

		if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
			r.Tags = []string{}
		}
		if err := json.Unmarshal([]byte(ingredientsJSON), &r.Ingredients); err != nil {
			r.Ingredients = []domain.Ingredient{}
		}
		if err := json.Unmarshal([]byte(instructionsJSON), &r.Instructions); err != nil {
			r.Instructions = []string{}
		}

		result[r.ID] = &r
	}

	return result, rows.Err()
}

func (s *RecipeStorage) GetAllPaginated(filter *domain.RecipeFilter, householdID string, limit, offset int) ([]domain.RecipeSummary, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Default and max limits
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// Scope by household: own recipes + seed recipes (NULL household_id)
	if householdID != "" {
		whereClause += " AND (household_id = ? OR household_id IS NULL)"
		args = append(args, householdID)
	} else {
		whereClause += " AND household_id IS NULL"
	}

	// Add name filter (case-insensitive partial match)
	if filter != nil && filter.Name != "" {
		whereClause += " AND LOWER(name) LIKE LOWER(?)"
		args = append(args, "%"+filter.Name+"%")
	}

	// Add tag filter (JSON search)
	if filter != nil && filter.Tag != "" {
		whereClause += " AND id IN (SELECT r2.id FROM recipes r2, json_each(r2.tags) WHERE json_each.value = ?)"
		args = append(args, filter.Tag)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM recipes %s", whereClause)
	var totalCount int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, name, servings, emoji, tags
		FROM recipes %s
		ORDER BY name
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var recipes []domain.RecipeSummary
	for rows.Next() {
		var r domain.RecipeSummary
		var emoji sql.NullString
		var tagsJSON string

		err := rows.Scan(&r.ID, &r.Name, &r.Servings, &emoji, &tagsJSON)
		if err != nil {
			return nil, 0, err
		}

		if emoji.Valid {
			r.Emoji = emoji.String
		}

		if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
			r.Tags = []string{}
		}

		recipes = append(recipes, r)
	}

	if recipes == nil {
		recipes = []domain.RecipeSummary{}
	}

	return recipes, totalCount, rows.Err()
}

func (s *RecipeStorage) Update(recipe *domain.Recipe) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tagsJSON, err := json.Marshal(recipe.Tags)
	if err != nil {
		return err
	}

	ingredientsJSON, err := json.Marshal(recipe.Ingredients)
	if err != nil {
		return err
	}

	instructionsJSON, err := json.Marshal(recipe.Instructions)
	if err != nil {
		return err
	}

	query := `
		UPDATE recipes SET name = ?, servings = ?, emoji = ?, tags = ?, ingredients = ?, instructions = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		recipe.Name, recipe.Servings, recipe.Emoji,
		string(tagsJSON), string(ingredientsJSON), string(instructionsJSON),
		recipe.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *RecipeStorage) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `DELETE FROM recipes WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *RecipeStorage) Create(recipe *domain.Recipe) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tagsJSON, err := json.Marshal(recipe.Tags)
	if err != nil {
		return err
	}

	ingredientsJSON, err := json.Marshal(recipe.Ingredients)
	if err != nil {
		return err
	}

	instructionsJSON, err := json.Marshal(recipe.Instructions)
	if err != nil {
		return err
	}

	// Use NULL for empty householdID (seed recipes)
	var householdID interface{}
	if recipe.HouseholdID != "" {
		householdID = recipe.HouseholdID
	}

	query := `
		INSERT INTO recipes (id, name, servings, emoji, tags, ingredients, instructions, household_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		recipe.ID, recipe.Name, recipe.Servings, recipe.Emoji,
		string(tagsJSON), string(ingredientsJSON), string(instructionsJSON),
		householdID, recipe.CreatedAt,
	)

	return err
}
