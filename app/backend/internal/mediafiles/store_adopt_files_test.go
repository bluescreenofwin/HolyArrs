package mediafiles

import (
	"context"
	"strings"
	"testing"

	"github.com/HolyArrs/holyarrs/app/backend/internal/core/database"
)

func TestAdoptedFileLookupUsesPathIndex(t *testing.T) {
	s, err := database.Open(t.TempDir(), database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	plan := func(query string, args ...any) string {
		rows, err := s.Read().QueryContext(context.Background(), "EXPLAIN QUERY PLAN "+query, args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var details []string
		for rows.Next() {
			var id, parent, unused int
			var detail string
			if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
				t.Fatal(err)
			}
			details = append(details, detail)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return strings.Join(details, "; ")
	}
	before := plan(adoptedOwnersQuery+" AND (? = '' OR lower(f.relative_path) = lower(?))", 1, "Film.mkv", "Film.mkv")
	after := plan(adoptedOwnersQuery+" AND lower(f.relative_path) = lower(?)", 1, "Film.mkv")
	t.Logf("optional filter: %s", before)
	t.Logf("targeted filter: %s", after)
	if !strings.Contains(after, "SEARCH f USING INDEX files_active_path") {
		t.Fatalf("single-file adoption scans the catalog: %s", after)
	}
}
