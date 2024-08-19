package roaring

import (
	"bytes"
	"fmt"
)

type Container struct {
	container
}

func (c *Container) Type() uint8 {
	return uint8(c.container.containerType())
}

func (c *Container) Max() uint16 {
	return c.container.maximum()
}

func (c *Container) String() string {
	return c.container.String()
}

func (c *Container) IsEmpty() bool {
	return c.container.isEmpty()
}

func (c *Container) From(typ uint8, data []byte) error {
	c.container = containerFromWire(typ, data)
	return nil
}

func (c *Container) Intersects(o *Container) bool {
	if c == nil || o == nil {
		return false
	}
	return c.container.intersects(o.container)
}

func (c *Container) Each(f func(uint16) bool) {
	if c == nil {
		return
	}
	c.container.iterate(f)
}

type Context interface {
	Value(cKey uint16, value func(uint8, []byte) error) error
	Write(cKey uint16, typ uint8, value []byte) error
}

func (r *Bitmap) Each(f func(cKey uint16, v *Container) error) error {
	for i, c := range r.highlowcontainer.containers {
		err := f(r.highlowcontainer.keys[i], &Container{c})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Bitmap) Save(ctx Context) error {
	var buf bytes.Buffer
	for i, c := range r.highlowcontainer.containers {
		k := r.highlowcontainer.keys[i]
		err := ctx.Value(k, func(u uint8, b []byte) error {
			c = containerFromWire(u, b).or(c)
			return nil
		})
		if err != nil {
			return err
		}
		buf.Reset()
		c = c.toEfficientContainer()
		_, err = c.writeTo(&buf)
		if err != nil {
			return err
		}
		err = ctx.Write(k, uint8(c.containerType()), bytes.Clone(buf.Bytes()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Bitmap) FromWire(cKey uint16, typ uint8, data []byte) {
	w := containerFromWire(typ, data)
	ra := &r.highlowcontainer
	i := ra.getIndex(cKey)
	if i >= 0 {
		c := ra.getWritableContainerAtIndex(i).or(w)
		r.highlowcontainer.setContainerAtIndex(i, c)
	} else {
		r.highlowcontainer.insertNewKeyValueAt(-i-1, cKey, w)
	}
}

func containerFromWire(typ uint8, b []byte) container {
	switch contype(typ) {
	case arrayContype:
		return &arrayContainer{
			byteSliceAsUint16Slice(b),
		}
	case run16Contype:
		return &runContainer16{
			iv: byteSliceAsInterval16Slice(b),
		}
	case bitmapContype:
		o := &bitmapContainer{
			bitmap: byteSliceAsUint64Slice(b),
		}
		o.computeCardinality()
		return o
	default:
		panic(fmt.Sprintf("unknown container type %d", typ))
	}
}
