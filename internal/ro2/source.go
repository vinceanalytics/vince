package ro2

import (
	"github.com/vinceanalytics/vince/internal/bsi"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
)

type KV []*roaring.Bitmap

func NewKV(tx *Tx, ts, shard uint64, field models.Field) *bsi.BSI {
	key := encoding.Bitmap(ts, shard, field, 0,
		make([]byte, encoding.BitmapKeySize))
	kv := make(KV, 0, 64)
	prefix := key[:len(key)-1]
	it := tx.Iter()
	for it.Seek(key); it.ValidForPrefix(prefix); it.Next() {
		value, _ := it.Item().ValueCopy(nil)
		b := roaring.FromBuffer(value)
		kv = append(kv, b)
	}
	return &bsi.BSI{Source: kv}
}

var _ bsi.Source = (*KV)(nil)

func (kv KV) GetOrCreate(i int) *roaring.Bitmap { return nil }

func (kv KV) Get(i int) *roaring.Bitmap {
	if i < len(kv) {
		return kv[i]
	}
	return nil
}
