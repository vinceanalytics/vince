package timeseries

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"
)

type meta struct {
	db *badger.DB
}

func (m *meta) SaveBucket(id ulid.ULID, start, end time.Time) error {
	return m.db.Update(func(txn *badger.Txn) error {
		key := []byte(formatKeyTime(start))
		var sheet *Series_Sheet
		if i, err := txn.Get(key); err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		} else {
			err = i.Value(func(val []byte) error {
				sheet = &Series_Sheet{}
				return proto.Unmarshal(val, sheet)
			})
			if err != nil {
				return err
			}
		}
		begin, _ := ptypes.TimestampProto(start)
		end, _ := ptypes.TimestampProto(start)

		if sheet == nil {
			// This is the first  sheet for the day
			sheet = &Series_Sheet{
				Buckets: []*Series_Bucket{
					{
						Id: id[:],
						Range: &Series_Range{
							Start: begin,
							End:   end,
						},
					},
				},
				Range: &Series_Range{
					Start: begin,
					End:   end,
				},
			}
		} else {
			sheet.Buckets = append(sheet.Buckets, &Series_Bucket{
				Id: id[:],
				Range: &Series_Range{
					Start: begin,
					End:   end,
				},
			})
			sheet.Range.End = end
		}
		b, err := proto.Marshal(sheet)
		if err != nil {
			return err
		}
		return txn.Set(key, b)
	})
}

func (m *meta) Buckets(start, end time.Time) (ids []string, err error) {
	err = m.iterate(start, end, func(s *Series_Sheet) {
		for _, b := range s.Buckets {
			from, to := b.Range.Time()
			if from.After(start) && to.Before(end) {
				var x ulid.ULID
				copy(x[:], b.Id)
				ids = append(ids, x.String())
			}
		}
	})
	return
}

func (r *Series_Range) Time() (begin time.Time, end time.Time) {
	begin, _ = ptypes.Timestamp(r.Start)
	end, _ = ptypes.Timestamp(r.End)
	return
}
func (m *meta) iterate(start time.Time, end time.Time, f func(*Series_Sheet)) error {
	return m.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		startKey := formatKeyTime(start)
		endKey := formatKeyTime(start)
		s := &Series_Sheet{}

		for it.Seek([]byte(startKey)); it.Valid(); it.Next() {
			k := string(it.Item().Key())
			switch {
			case startKey == endKey:
				return it.Item().Value(func(val []byte) error {
					err := proto.Unmarshal(val, s)
					if err != nil {
						return err
					}
					f(s)
					return nil
				})
			case k <= endKey:
				err := it.Item().Value(func(val []byte) error {
					err := proto.Unmarshal(val, s)
					if err != nil {
						return err
					}
					f(s)
					return nil
				})
				if err != nil {
					return err
				}
			default:
				return nil
			}
		}
		return nil
	})
}
func (m *meta) Close() error {
	return m.db.Close()
}

func formatKeyTime(ts time.Time) string {
	return ts.Format("2006-01")
}
