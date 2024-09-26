package ro2

import (
	"strconv"

	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/query"
	wq "github.com/vinceanalytics/vince/internal/web/query"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (tx *Tx) compile(fs wq.Filters) query.Filter {
	a := make(query.And, 0, len(fs))
	for _, f := range fs {
		switch f.Key {
		case "city":
			switch f.Op {
			case "is":
				code, _ := strconv.Atoi(f.Value[0])
				if code == 0 {
					return query.Noop{}
				}
				return bsi.Filter(f.Key, bsi.EQ, int64(code), 0)

			}
		default:
			fd := fields.ByName(protoreflect.Name(f.Key))
			if fd == nil {
				return query.Noop{}
			}
			fx := uint64(fd.Number())

			switch f.Op {
			case "is":
				var values []uint64
				for _, v := range f.Value {
					id, ok := tx.ID(fx, v)
					if ok {
						values = append(values, id)
					}
				}
				if len(values) == 0 {
					return query.Noop{}
				}
				a = append(a, matchMutex(f.Key, values))
			case "matches":
				var values []uint64
				for _, source := range f.Value {
					prefix, exact := searchPrefix([]byte(source))
					if exact {
						id, ok := tx.ID(fx, source)
						if ok {
							values = append(values, id)
						}
					} else {
						tx.Search(fx, prefix, func(b []byte, u uint64) {
							values = append(values, u)
						})
					}
				}
				if len(values) == 0 {
					return query.Noop{}
				}
				a = append(a, matchMutex(f.Key, values))
			default:
				return query.Noop{}
			}
		}
	}
	return a
}

func matchMutex(field string, id []uint64) query.Filter {
	key := []byte(field)
	prefix := len(field)
	return query.FilterFn(func(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (r *rows.Row, err error) {
		bitmap := string(append(key[:prefix], view...))
		err = viewCu(rtx, bitmap, func(rCu *rbf.Cursor) error {
			if len(id) == 1 {
				r, err = cursor.Row(rCu, shard, id[0])
				return err
			}
			m := make([]*rows.Row, len(id))

			for i := range id {
				m[i], err = cursor.Row(rCu, shard, id[i])
				if err != nil {
					return err
				}
			}
			r = m[0].Union(m[1:]...)
			return nil
		})
		return

	})
}
