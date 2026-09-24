package sqlite

import "testing"

// TestMigration021_TagsSeedRecipeProvenance pins the provenance counts that
// docs/recipe-provenance.md relies on: every shared seed recipe carries a
// source, 110 of them are ICA.se-derived, and none are left unlabelled.
func TestMigration021_TagsSeedRecipeProvenance(t *testing.T) {
	db := newMigratedMemoryDB(t)

	counts := map[string]int{}
	rows, err := db.Query(`SELECT COALESCE(source, ''), COUNT(*) FROM recipes WHERE household_id IS NULL GROUP BY source`)
	if err != nil {
		t.Fatalf("query sources: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var src string
		var n int
		if err := rows.Scan(&src, &n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		counts[src] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	if counts["ica.se"] != 110 {
		t.Errorf("ica.se seed recipes = %d, want 110", counts["ica.se"])
	}
	if counts["maltiden-seed"] != 20 {
		t.Errorf("maltiden-seed recipes = %d, want 20", counts["maltiden-seed"])
	}
	if counts[""] != 0 {
		t.Errorf("%d seed recipes have no source", counts[""])
	}
}
