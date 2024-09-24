package tx

import (
	"fmt"

	"github.com/vinceanalytics/vince/internal/rbf"
)

type Tx struct {
	tx    *rbf.Tx
	shard uint64
}

func New(tx *rbf.Tx, shard uint64) *Tx {
	return &Tx{tx: tx, shard: shard}
}

func (tx *Tx) Views() []string {
	return tx.tx.FieldViews()
}

func (tx *Tx) Shard() uint64 {
	return tx.shard
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
