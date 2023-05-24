package timeseries

import (
	"context"
	"encoding/binary"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
)

func FindCity(ctx context.Context, geoname uint32) (s string) {
	if geoname == 0 {
		return
	}
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, geoname)
	err := GetGeo(ctx).View(func(txn *badger.Txn) error {
		x, err := txn.Get(b)
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to get city by geoname id")
			return nil
		}
		return x.Value(func(val []byte) error {
			s = string(val)
			return nil
		})
	})
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to get city by geoname id")
	}
	return
}
