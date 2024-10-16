package ro2

import (
	"github.com/vinceanalytics/vince/internal/bsi"
	"github.com/vinceanalytics/vince/internal/roaring"
)

type KV []*roaring.Bitmap

func (tx *Tx) newKv(key []byte) *bsi.BSI {
	pos := len(tx.bitmaps)
	prefix := key[:len(key)-1]
	it := tx.Iter()
	for it.Seek(key); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		item.Value(func(val []byte) error {
			dst := tx.enc.Allocate(len(val))
			copy(dst, val)
			b := roaring.FromBuffer(dst)
			tx.bitmaps = append(tx.bitmaps, b)
			return nil
		})
	}
	kv := tx.bitmaps[pos:len(tx.bitmaps)]
	if tx.pos < len(tx.kv) {
		b := &tx.kv[tx.pos]
		tx.pos++
		b.Source = KV(kv)
		return b
	}
	return &bsi.BSI{Source: KV(kv)}
}

var _ bsi.Source = (*KV)(nil)

func (kv KV) GetOrCreate(i int) *roaring.Bitmap { return nil }

func (kv KV) Get(i int) *roaring.Bitmap {
	if i < len(kv) {
		return kv[i]
	}
	return nil
}
