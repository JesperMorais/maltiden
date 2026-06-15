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

	query := `SELECT id, name, servings, emoji, tags, ingredients, instructions, household_id, calories, protein_g, carbs_g, fat_g, created_at
			  FROM recipes WHERE id = ?`

	var r domain.Recipe
	var emoji, householdID sql.NullString
	var tagsJSON, ingredientsJSON, instructionsJSON string
	var cal, prot, carb, fat sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.Name, &r.Servings, &emoji,
		&tagsJSON, &ingredientsJSON, &instructionsJSON, &householdID,
		&cal, &prot, &carb, &fat, &r.CreatedAt,
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
	r.Nutrition = scanNutrition(cal, prot, carb, fat)

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
		SELECT id, name, servings, emoji, tags, ingredients, instructions, calories, protein_g, carbs_g, fat_g, created_at
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
		var cal, prot, carb, fat sql.NullFloat64

		err := rows.Scan(
			&r.ID, &r.Name, &r.Servings, &emoji,
			&tagsJSON, &ingredientsJSON, &instructionsJSON,
			&cal, &prot, &carb, &fat, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if emoji.Valid {
			r.Emoji = emoji.String
		}
		r.Nutrition = scanNutrition(cal, prot, carb, fat)

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

	cal, prot, carb, fat := nutritionArgs(recipe.Nutrition)

	query := `
		UPDATE recipes SET name = ?, servings = ?, emoji = ?, tags = ?, ingredients = ?, instructions = ?,
			calories = ?, protein_g = ?, carbs_g = ?, fat_g = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		recipe.Name, recipe.Servings, recipe.Emoji,
		string(tagsJSON), string(ingredientsJSON), string(instructionsJSON),
		cal, prot, carb, fat,
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

	cal, prot, carb, fat := nutritionArgs(recipe.Nutrition)

	query := `
		INSERT INTO recipes (id, name, servings, emoji, tags, ingredients, instructions, household_id, calories, protein_g, carbs_g, fat_g, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		recipe.ID, recipe.Name, recipe.Servings, recipe.Emoji,
		string(tagsJSON), string(ingredientsJSON), string(instructionsJSON),
		householdID, cal, prot, carb, fat, recipe.CreatedAt,
	)

	return err
}

// nutritionArgs maps an optional Nutrition struct to four nullable SQL args.
// A nil Nutrition (or nil field) becomes a NULL column, keeping the per-serving
// nutrition data fully optional.
func nutritionArgs(n *domain.Nutrition) (cal, prot, carb, fat interface{}) {
	if n == nil {
		return nil, nil, nil, nil
	}
	toArg := func(v *float64) interface{} {
		if v == nil {
			return nil
		}
		return *v
	}
	return toArg(n.Calories), toArg(n.ProteinG), toArg(n.CarbsG), toArg(n.FatG)
}

// scanNutrition assembles a Nutrition struct from four nullable columns,
// returning nil when every field is NULL (recipe has no nutrition data).
func scanNutrition(cal, prot, carb, fat sql.NullFloat64) *domain.Nutrition {
	if !cal.Valid && !prot.Valid && !carb.Valid && !fat.Valid {
		return nil
	}
	n := &domain.Nutrition{}
	if cal.Valid {
		v := cal.Float64
		n.Calories = &v
	}
	if prot.Valid {
		v := prot.Float64
		n.ProteinG = &v
	}
	if carb.Valid {
		v := carb.Float64
		n.CarbsG = &v
	}
	if fat.Valid {
		v := fat.Float64
		n.FatG = &v
	}
	return n
}
