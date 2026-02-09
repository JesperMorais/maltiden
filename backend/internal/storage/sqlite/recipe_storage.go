package sqlite

import (
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
)

type RecipeStorage struct {
	db *sql.DB
}

func NewRecipeStorage(db *sql.DB) *RecipeStorage {
	return &RecipeStorage{db: db}
}

func (s *RecipeStorage) GetAll(filter *domain.RecipeFilter) ([]domain.RecipeSummary, error) {
	query := `SELECT id, name, servings, emoji, tags FROM recipes WHERE 1=1`
	args := []interface{}{}

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

	rows, err := s.db.Query(query, args...)
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
	query := `SELECT id, name, servings, emoji, tags, ingredients, instructions, created_at
			  FROM recipes WHERE id = ?`

	var r domain.Recipe
	var emoji sql.NullString
	var tagsJSON, ingredientsJSON, instructionsJSON string

	err := s.db.QueryRow(query, id).Scan(
		&r.ID, &r.Name, &r.Servings, &emoji,
		&tagsJSON, &ingredientsJSON, &instructionsJSON, &r.CreatedAt,
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

func (s *RecipeStorage) Create(recipe *domain.Recipe) error {
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
		INSERT INTO recipes (id, name, servings, emoji, tags, ingredients, instructions, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		recipe.ID, recipe.Name, recipe.Servings, recipe.Emoji,
		string(tagsJSON), string(ingredientsJSON), string(instructionsJSON),
		recipe.CreatedAt,
	)

	return err
}
