package kase

import (
	"github.com/gernest/rbf"
	"github.com/gernest/roaring"
	"github.com/gernest/rows"
	"github.com/stretchr/testify/suite"
)

//go:generate protoc -I=. --go_out=. --go_opt=paths=source_relative msg.proto

type Add[T any] func(r *roaring.Bitmap, id uint64, value T)

type Extract[T any] func(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value T) error) error

type Kase[T any] struct {
	suite.Suite
	Add     Add[T]
	Extract Extract[T]
	Source  []T
	db      *rbf.DB
}

func (k *Kase[T]) SetupSuite() {
	k.Require().Greater(len(k.Source), 4) // we need at least for samples

	db := rbf.NewDB(k.T().TempDir(), nil)
	k.Require().NoError(db.Open())
	k.db = db

	tx, err := db.Begin(true)
	k.Require().NoError(err)
	defer tx.Rollback()
	r := roaring.NewBitmap()
	for i := range k.Source {
		k.Add(r, uint64(i), k.Source[i])
	}
	_, err = tx.AddRoaring("kase", r)
	k.Require().NoError(err)
	k.Require().NoError(tx.Commit())
}

func (k *Kase[T]) TearDownSuite() {
	k.Require().NoError(k.db.Close())
}

func (k *Kase[T]) TestSelect() {
	want := map[uint64]T{
		1: k.Source[1],
		3: k.Source[3],
	}
	got := map[uint64]T{}
	k.view(func(c *rbf.Cursor) {
		err := k.Extract(c, 0, rows.NewRow(1, 3), func(column uint64, value T) error {
			got[column] = value
			return nil
		})
		k.Require().NoError(err)
	})
	k.Require().Equal(want, got)
}

func (k *Kase[T]) TestSelectAll() {
	keys := make([]uint64, 0, len(k.Source))
	want := map[uint64]T{}
	for i := range k.Source {
		want[uint64(i)] = k.Source[i]
		keys = append(keys, uint64(i))
	}
	got := map[uint64]T{}
	k.view(func(c *rbf.Cursor) {
		err := k.Extract(c, 0, rows.NewRow(keys...), func(column uint64, value T) error {
			got[column] = value
			return nil
		})
		k.Require().NoError(err)
	})
	k.Require().Equal(want, got)
}

func (k *Kase[T]) view(f func(c *rbf.Cursor)) {
	tx, err := k.db.Begin(false)
	k.Require().NoError(err)
	defer tx.Rollback()

	c, err := tx.Cursor("kase")
	k.Require().NoError(err)
	defer c.Close()
	f(c)
}
