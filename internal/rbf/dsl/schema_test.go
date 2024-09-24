package dsl

import (
	"math"
	"testing"

	"github.com/gernest/rbf/dsl/kase"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/roaring/shardwidth"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	db, err := New[*kase.Model](t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	db.Append([]*kase.Model{
		{},
		{
			Enum:    kase.Model_one,
			Bool:    true,
			String_: "hello",
			Blob:    []byte("hello"),
			Int64:   math.MaxInt64,
			Uint64:  shardwidth.ShardWidth,
			Double:  math.MaxFloat64,
			Set:     []string{"hello"},
			BlobSet: [][]byte{[]byte("hello")},
		},
	})
	require.NoError(t, db.Flush())
	want := []string{
		"~_id;0<",
		"~blob;0<", "~blob_set;0<",
		"~bool;0<", "~double;0<", "~enum;0<", "~int64;0<", "~set;0<", "~string;0<", "~uint64;0<"}
	var got []string
	r, err := db.Reader()
	require.NoError(t, err)
	defer r.Release()
	err = r.View(func(txn *tx.Tx) error {
		got = txn.Views()
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func BenchmarkWrite(b *testing.B) {
	db, err := New[*kase.Model](b.TempDir())
	require.NoError(b, err)
	defer db.Close()
	data := []*kase.Model{
		{},
		{
			Enum:    kase.Model_one,
			Bool:    true,
			String_: "hello",
			Blob:    []byte("hello"),
			Int64:   math.MaxInt64,
			Uint64:  shardwidth.ShardWidth,
			Double:  math.MaxFloat64,
			Set:     []string{"hello"},
			BlobSet: [][]byte{[]byte("hello")},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		db.Append(data)
		db.Flush()
	}
}
