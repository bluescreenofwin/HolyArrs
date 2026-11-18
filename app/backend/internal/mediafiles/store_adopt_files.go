package mediafiles

import (
	"context"
	"strings"

	"github.com/HolyArrs/holyarrs/app/backend/internal/core/database"
)

// AdoptedFileOwners identifies files independently of shared parent folders.
// An empty only returns all owned paths; otherwise it filters one path.
func AdoptedFileOwners(ctx context.Context, db database.Lister, libraryFolderID int64, fold bool, only string) (map[string]int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT f.relative_path, v.title_id FROM files f
		JOIN versions v ON v.id = f.version_id JOIN titles t ON t.id = v.title_id
		WHERE v.library_folder_id = ? AND t.state != 'removed' AND f.state IN ('present','missing')
		AND (? = '' OR lower(f.relative_path) = lower(?))`, libraryFolderID, only, only)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var rel string
		var id int64
		if err := rows.Scan(&rel, &id); err != nil {
			return nil, err
		}
		if fold {
			rel = strings.ToLower(rel)
		}
		out[rel] = id
	}
	return out, rows.Err()
}
