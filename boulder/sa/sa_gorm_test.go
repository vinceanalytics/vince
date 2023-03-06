package sa

import (
	"path/filepath"
	"testing"
)

func TestMigration(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "boulder.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		x, _ := db.DB()
		x.Close()
	})
}
