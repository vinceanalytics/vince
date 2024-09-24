package tx

import (
	"fmt"

	"github.com/blevesearch/vellum"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/tr"
)

type Tx struct {
	tx    *rbf.Tx
	shard uint64
	tr    *tr.Read
}

func New(tx *rbf.Tx, shard uint64, tr *tr.Read) *Tx {
	return &Tx{tx: tx, shard: shard, tr: tr}
}

func (tx *Tx) Views() []string {
	return tx.tx.FieldViews()
}

func (tx *Tx) Shard() uint64 {
	return tx.shard
}

func (tx *Tx) Find(field string, key []byte) (uint64, bool) {
	return tx.tr.Find(field, key)
}

func (tx *Tx) Blob(field string, id uint64) []byte {
	return tx.tr.Blob(field, id)
}

func (tx *Tx) Key(field string, id uint64) []byte {
	return tx.tr.Key(field, id)
}

func (tx *Tx) Keys(field string, id []uint64, f func(value []byte)) {
	tx.tr.Keys(field, id, f)
}

func (tx *Tx) Search(field string, a vellum.Automaton, start []byte, end []byte, match func(key []byte, value uint64) error) error {
	return tx.tr.Search(field, a, start, end, match)
}

func (tx *Tx) SearchRe(field, like string, start []byte, end []byte, match func(key []byte, value uint64) error) error {
	return tx.tr.SearchRe(field, like, start, end, match)
}

func (tx *Tx) Get(field string) (*rbf.Cursor, error) {
	return tx.tx.Cursor(ViewKey(field, tx.shard))
}

func (tx *Tx) Cursor(field string, f func(c *rbf.Cursor, tx *Tx) error) error {
	c, err := tx.Get(field)
	if err != nil {
		return err
	}
	defer c.Close()
	return f(c, tx)
}

func ViewKey(field string, shard uint64) string {
	return fmt.Sprintf("~%s;%d<", field, shard)
}

func ViewKeyPrefix(field string) string {
	return fmt.Sprintf("~%s;", field)
}
