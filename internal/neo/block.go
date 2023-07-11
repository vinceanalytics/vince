package neo

import (
	"bytes"
	"errors"
	"io"
	"path"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/pkg/entry"
	"google.golang.org/protobuf/proto"
)

const (
	MetaFile    = "METADATA"
	MetaPrefix  = "meta"
	BlockPrefix = "block"
)

type ActiveBlock struct {
	domain   string
	Min, Max time.Time
	bytes.Buffer
	// rows are already buffered with w. It is wise to send entries to w as they
	// arrive. tmp allows us to avoid creating a slice with one entry on every
	// WriteRow call
	tmp [1]*entry.Entry
	w   *parquet.SortingWriter[*entry.Entry]
}

func (a *ActiveBlock) Init(domain string) {
	a.domain = domain
	a.w = Writer[*entry.Entry](a)
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	if a.Min.IsZero() {
		a.Min = e.Timestamp
	}
	a.Max = e.Timestamp
	a.tmp[0] = e.Clone()
	a.w.Write(a.tmp[:])
}

func (a *ActiveBlock) Reset() {
	a.Buffer.Reset()
	a.Min, a.Max = time.Time{}, time.Time{}
	a.domain = ""
	a.w.Reset(a)
}

func (a *ActiveBlock) Save(db *badger.DB) error {
	err := a.w.Close()
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		meta := &Metadata{}
		metaPath := path.Join(MetaPrefix, a.domain, MetaFile)
		if x, err := txn.Get([]byte(metaPath)); err != nil {
			meta = &Metadata{}
			err := x.Value(func(val []byte) error {
				return proto.Unmarshal(val, meta)
			})
			if err != nil {
				return err
			}
		}
		id := ulid.Make().String()
		blockPath := path.Join(BlockPrefix, a.domain, id)

		meta.Blocks = append(meta.Blocks, &Block{
			Id:   id,
			Min:  a.Min.UnixMilli(),
			Max:  a.Max.UnixMilli(),
			Size: int64(a.Len()),
		})
		mb, err := proto.Marshal(meta)
		if err != nil {
			return err
		}
		return errors.Join(
			txn.Set([]byte(blockPath), a.Bytes()),
			txn.Set([]byte(metaPath), mb),
		)
	})

}

// Writer returns a parquet.SortingWriter for T that sorts timestamp field in
// ascending order.
func Writer[T any](w io.Writer) *parquet.SortingWriter[T] {
	return parquet.NewSortingWriter[T](w, 4<<10, parquet.SortingWriterConfig(
		parquet.SortingColumns(
			parquet.Ascending("timestamp"),
		),
	))
}
