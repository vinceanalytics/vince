package timeseries

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gernest/vince/timex"
)

func TestKey(t *testing.T) {
	var id ID
	table := SYSTEM
	t.Run("sets table", func(t *testing.T) {
		id.SetTable(table)
		if id.GetTable() != SYSTEM {
			t.Fatal("failed to encode table")
		}
	})
	uid := uint64(rand.Int63())

	t.Run("sets user id", func(t *testing.T) {
		id.SetUserID(uid)
		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
	})
	sid := uint64(rand.Int63())

	t.Run("sets site id", func(t *testing.T) {
		id.SetSiteID(sid)
		if id.GetTable() != SYSTEM {
			t.Fatal("failed to encode table")
		}
		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
		if id.GetSiteID() != sid {
			t.Fatalf("expected %d got %d", sid, id.GetSiteID())
		}
	})

	now := time.Now()
	today := timex.Date(now)
	t.Run("sets date", func(t *testing.T) {
		id.SetDate(today)
		if id.GetTable() != SYSTEM {
			t.Fatal("failed to encode table")
		}
		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
		if id.GetSiteID() != sid {
			t.Fatalf("expected %d got %d", sid, id.GetSiteID())
		}
		if !id.GetTime().Equal(today) {
			t.Fatalf("expected %s got %s", today, id.GetTime())
		}
	})
	t.Run("sets entropy", func(t *testing.T) {
		id.SetEntropy()
		id.SetDate(today)
		if id.GetTable() != SYSTEM {
			t.Fatal("failed to encode table")
		}
		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
		if id.GetSiteID() != sid {
			t.Fatalf("expected %d got %d", sid, id.GetSiteID())
		}
		if !id.GetTime().Equal(today) {
			t.Fatalf("expected %s got %s", today, id.GetTime())
		}
	})
}
