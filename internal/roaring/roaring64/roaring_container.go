package roaring64

import (
	"github.com/vinceanalytics/vince/internal/roaring"
)

func NewFromMap(m *roaring.Bitmap) *Bitmap {
	if m.IsEmpty() {
		return &Bitmap{}
	}
	return &Bitmap{
		highlowcontainer: roaringArray64{
			keys:            []uint32{0},
			containers:      []*roaring.Bitmap{m},
			needCopyOnWrite: []bool{false},
		},
	}
}

type Context interface {
	Value(key uint32, cKey uint16, value func(uint8, []byte) error) error
	Write(key uint32, cKey uint16, typ uint8, value []byte) error
}

type wrapCtx struct {
	key uint32
	ctx Context
}

var _ roaring.Context = (*wrapCtx)(nil)

func (w *wrapCtx) Value(cKey uint16, value func(uint8, []byte) error) error {
	return w.ctx.Value(w.key, cKey, value)
}

func (w *wrapCtx) Write(cKey uint16, typ uint8, value []byte) error {
	return w.ctx.Write(w.key, cKey, typ, value)
}

func (r *Bitmap) Save(ctx Context) error {
	w := wrapCtx{ctx: ctx}
	for i, c := range r.highlowcontainer.containers {
		k := r.highlowcontainer.keys[i]
		w.key = k
		err := c.Save(&w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Bitmap) Each(f func(key uint32, cKey uint16, value *roaring.Container) error) error {
	for i, c := range r.highlowcontainer.containers {
		k := r.highlowcontainer.keys[i]
		err := c.Each(func(cKey uint16, v *roaring.Container) error {
			return f(k, cKey, v)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
