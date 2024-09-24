package dsl

import (
	"cmp"
	"errors"
	"path/filepath"
	"slices"

	"github.com/gernest/rbf/dsl/tr"
	"github.com/gernest/roaring"
	"go.etcd.io/bbolt"
)

var (
	viewsBucket = []byte("views")
	seqBucket   = []byte("seq")
)

type Ops struct {
	db *bbolt.DB
	tr *tr.File
}

func (o *Ops) Close() error {
	return errors.Join(o.db.Close(), o.tr.Close())
}

func newOps(path string) (*Ops, error) {
	full := filepath.Join(path, "OPS")
	db, err := bbolt.Open(full, 0600, nil)
	if err != nil {
		return nil, err
	}
	fullTr := filepath.Join(path, "TRANSLATE")
	tr := tr.New(fullTr)
	err = tr.Open()
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(viewsBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(seqBucket)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Ops{db: db, tr: tr}, nil
}

func (o *Ops) read() (*readOps, error) {
	tx, err := o.db.Begin(false)
	if err != nil {
		return nil, err
	}
	r, err := o.tr.Read()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return &readOps{
		tx:    tx,
		tr:    r,
		views: tx.Bucket(viewsBucket),
	}, nil
}

type readOps struct {
	tx    *bbolt.Tx
	tr    *tr.Read
	views *bbolt.Bucket
}

func (r *readOps) Release() error {
	return errors.Join(r.tx.Rollback(), r.tr.Release())
}

type Shard struct {
	Shard uint64
	Views []string
}

// Shards returns shard -> view mapping for given views. Useful when querying
// quantum fields. This ensures we only open the shard once and process all
// views for the shard together.
func (r *readOps) Shards(views ...string) []Shard {
	m := map[uint64][]string{}
	for _, view := range views {
		data := r.views.Get([]byte(view))
		if data != nil {
			r := roaring.NewBitmap()
			r.UnmarshalBinary(data)
			it := r.Iterator()
			for nxt, eof := it.Next(); !eof; nxt, eof = it.Next() {
				m[nxt] = append(m[nxt], view)
			}
		}
	}
	o := make([]Shard, 0, len(m))
	for s, v := range m {
		slices.Sort(v)
		o = append(o, Shard{
			Shard: s, Views: slices.Compact(v),
		})
	}
	slices.SortFunc(o, func(a, b Shard) int {
		return cmp.Compare(a.Shard, b.Shard)
	})
	return o
}

func (r *readOps) All() []Shard {
	m := map[uint64][]string{}
	r.views.ForEach(func(k, v []byte) error {
		if v != nil {
			r := roaring.NewBitmap()
			r.UnmarshalBinary(v)
			it := r.Iterator()
			view := string(k)
			for nxt, eof := it.Next(); !eof; nxt, eof = it.Next() {
				m[nxt] = append(m[nxt], view)
			}
		}
		return nil
	})
	o := make([]Shard, 0, len(m))
	for s, v := range m {
		slices.Sort(v)
		o = append(o, Shard{
			Shard: s, Views: slices.Compact(v),
		})
	}
	slices.SortFunc(o, func(a, b Shard) int {
		return cmp.Compare(a.Shard, b.Shard)
	})
	return o
}

func (o *Ops) write() (*writeOps, error) {
	tx, err := o.db.Begin(true)
	if err != nil {
		return nil, err
	}
	w, err := o.tr.Write()
	if err != nil {
		return nil, err
	}
	return &writeOps{
		tx:    tx,
		tr:    w,
		views: tx.Bucket(viewsBucket),
		seq:   tx.Bucket(seqBucket),
	}, nil
}

type writeOps struct {
	tx    *bbolt.Tx
	tr    *tr.Write
	views *bbolt.Bucket
	seq   *bbolt.Bucket
}

func (o *writeOps) Release() error {
	return errors.Join(o.tx.Rollback(), o.tr.Release())
}

func (o *writeOps) fill(ids []uint64) (err error) {
	for i := range ids {
		ids[i], err = o.seq.NextSequence()
		if err != nil {
			return
		}
	}
	return
}

func (o *writeOps) Commit() error {
	return errors.Join(o.tx.Commit(), o.tr.Commit())
}
