package ro2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func TestStore_sequence(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	// zero sequence at the beginning
	require.Equal(t, uint64(0), db.Seq())
	err = db.One(&v1.Model{
		Timestamp: 1,
		Country:   "TZ",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), db.Seq())
	db.Close()

	db, err = Open(dir)
	require.NoError(t, err)
	defer db.Close()
	var country string
	var tr string
	var id uint64
	db.View(func(tx *Tx) error {
		tx.Search(uint64(alicia.COUNTRY), nil, func(key []byte, val uint64) {
			country = string(key)
			id = val
		})
		tr = tx.Find(uint64(alicia.COUNTRY), 1)
		return nil
	})
	require.Equal(t, uint64(1), db.Seq())
	require.Equal(t, "TZ", country)
	require.Equal(t, "TZ", tr)
	require.Equal(t, uint64(1), id)
}

func TestStore_Bounce(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	sample := []int32{1, 1, -1, 1, -1, -1}
	for _, k := range sample {
		db.One(&v1.Model{
			Bounce: k,
		})
	}
	var yes, no *roaring64.Bitmap
	all := roaring64.NewDefaultBSI()
	db.View(func(tx *Tx) error {
		yes = tx.Row(0, uint64(alicia.BOUNCE), 0)
		no = tx.Row(0, uint64(alicia.BOUNCE), 1)
		tx.ExtractBounce(0, uint64(alicia.BOUNCE), nil, all.SetValue)
		return nil
	})
	require.Equal(t, []uint64{1, 2, 4}, yes.ToArray())
	require.Equal(t, []uint64{3, 5, 6}, no.ToArray())
	var got []int32
	for _, v := range all.GetExistenceBitmap().ToArray() {
		n, _ := all.GetValue(v)
		got = append(got, int32(n))
	}
	require.Equal(t, sample, got)
}

func TestStore_quantum(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	ts, _ := time.Parse(time.RFC822, time.RFC822)
	ts = ts.UTC()
	err = db.One(&v1.Model{
		Timestamp: ts.UnixMilli(),
	})
	require.NoError(t, err)
	got := map[alicia.Field]map[int64][]uint64{
		alicia.MINUTE: {},
		alicia.HOUR:   {},
		alicia.DAY:    {},
		alicia.WEEK:   {},
		alicia.MONTH:  {},
	}

	extract := func(tx *Tx, field alicia.Field) {
		tx.ExtractBSI(0, uint64(field), nil, func(row uint64, c int64) {
			got[field][c] = append(got[field][c], row)
		})
	}
	db.View(func(tx *Tx) error {
		extract(tx, alicia.MINUTE)
		extract(tx, alicia.HOUR)
		extract(tx, alicia.DAY)
		extract(tx, alicia.WEEK)
		extract(tx, alicia.MONTH)
		return nil
	})
	want := map[alicia.Field]map[int64][]uint64{
		alicia.MINUTE: {
			minute(ts).UnixMilli(): []uint64{1},
		},
		alicia.HOUR: {
			hour(ts).UnixMilli(): []uint64{1},
		},
		alicia.DAY: {
			day(ts).UnixMilli(): []uint64{1},
		},
		alicia.WEEK: {
			week(ts).UnixMilli(): []uint64{1},
		},
		alicia.MONTH: {
			month(ts).UnixMilli(): []uint64{1},
		},
	}
	require.Equal(t, want, got)
}

func BenchmarkAddOne(t *testing.B) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	// zero sequence at the beginning
	m := &v1.Model{
		Timestamp: time.Now().UnixMilli(),
		Country:   "TZ",
	}
	t.ResetTimer()
	t.ReportAllocs()
	for range t.N {
		db.One(m)
	}
}
