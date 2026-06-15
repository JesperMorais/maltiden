package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"time"
)

type MenuPreferencesStorage struct {
	db *sql.DB
}

func NewMenuPreferencesStorage(db *sql.DB) *MenuPreferencesStorage {
	return &MenuPreferencesStorage{db: db}
}

func (s *MenuPreferencesStorage) Get(householdID string) (*domain.MenuPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		prefs        domain.MenuPreferences
		excludedJSON string
		dislikedJSON string
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT household_id, excluded_tags, default_days, default_servings, vegetarian_days, updated_at,
		        diet_profile, disliked_ingredients
		 FROM menu_preferences WHERE household_id = ?`,
		householdID,
	).Scan(
		&prefs.HouseholdID,
		&excludedJSON,
		&prefs.DefaultDays,
		&prefs.DefaultServings,
		&prefs.VegetarianDays,
		&prefs.UpdatedAt,
		&prefs.DietProfile,
		&dislikedJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(excludedJSON), &prefs.ExcludedTags); err != nil {
		return nil, err
	}
	if prefs.ExcludedTags == nil {
		prefs.ExcludedTags = []string{}
	}

	if err := json.Unmarshal([]byte(dislikedJSON), &prefs.DislikedIngredients); err != nil {
		return nil, err
	}
	if prefs.DislikedIngredients == nil {
		prefs.DislikedIngredients = []string{}
	}

	return &prefs, nil
}

func (s *MenuPreferencesStorage) Upsert(prefs *domain.MenuPreferences) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tags := prefs.ExcludedTags
	if tags == nil {
		tags = []string{}
	}
	excludedJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	disliked := prefs.DislikedIngredients
	if disliked == nil {
		disliked = []string{}
	}
	dislikedJSON, err := json.Marshal(disliked)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO menu_preferences
		     (household_id, excluded_tags, default_days, default_servings, vegetarian_days, updated_at,
		      diet_profile, disliked_ingredients)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?)
		 ON CONFLICT(household_id) DO UPDATE SET
		     excluded_tags        = excluded.excluded_tags,
		     default_days         = excluded.default_days,
		     default_servings     = excluded.default_servings,
		     vegetarian_days      = excluded.vegetarian_days,
		     diet_profile         = excluded.diet_profile,
		     disliked_ingredients = excluded.disliked_ingredients,
		     updated_at           = CURRENT_TIMESTAMP`,
		prefs.HouseholdID, string(excludedJSON), prefs.DefaultDays, prefs.DefaultServings, prefs.VegetarianDays,
		prefs.DietProfile, string(dislikedJSON),
	)
	return err
}
