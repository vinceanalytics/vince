package models

import (
	"path/filepath"
	"testing"
)

func TestQueryCacheSites(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "vince.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		CloseDB(db)
	})
	usr := &User{
		Name:  "john doe",
		Email: "test@example.com",
		Sites: []*Site{
			{
				Domain: "vince.test",
			},
		},
	}
	err = db.Save(usr).Error
	if err != nil {
		t.Fatal(err)
	}
	usr2 := &User{
		Name:  "john doe",
		Email: "test2@example.com",
	}
	err = db.Save(usr2).Error
	if err != nil {
		t.Fatal(err)
	}
	mem := &SiteMembership{
		UserID: usr2.ID,
		SiteID: usr.Sites[0].ID,
		Role:   "viewer",
	}
	err = db.Save(mem).Error
	if err != nil {
		t.Fatal(err)
	}
	var c []*CachedSite
	err = QuerySitesToCache(db, &c)
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 1 {
		t.Fatalf("expected 1 got %d", len(c))
	}
	if c[0].UserID != usr.ID {
		t.Errorf("expected %d got %d", usr.ID, c[0].UserID)
	}
}
