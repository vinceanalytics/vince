package ro2

import (
	"errors"
	"fmt"

	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/mutex"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/query"
)

func domainTx(rtx *rbf.Tx, view []byte, shard, domainId uint64, filter query.Filter) (*rows.Row, error) {
	rCu, err := rtx.Cursor(string(view))
	if err != nil {
		if errors.Is(err, rbf.ErrBitmapNotFound) {
			return rows.NewRow(), nil
		}
		return nil, err
	}
	defer rCu.Close()

	dRow, err := cursor.Row(rCu, shard, domainId)
	if err != nil {
		return nil, err
	}
	if dRow.IsEmpty() {
		return dRow, nil
	}
	return filter.Apply(rtx, shard, view[len(domainField):], dRow)
}

func propertyTx(rtx *rbf.Tx, view string, r *rows.Row, f func(row uint64, columns *rows.Row) error) error {
	rCu, err := rtx.Cursor(view)
	if err != nil {
		if errors.Is(err, rbf.ErrBitmapNotFound) {
			return nil
		}
		return err
	}
	defer rCu.Close()
	fmt.Println(view)
	return mutex.Distinct(rCu, r, f)
}
