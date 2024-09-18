package roaring

import (
	"bytes"
	"fmt"
	"unsafe"
)

type Container struct {
	container
}

func toInterval16(a []byte) []interval16 {
	return (*[2048]interval16)(unsafe.Pointer(&a[0]))[: len(a)/4 : len(a)/4]
}

func fromInterval16(a []interval16) []byte {
	return (*[8192]byte)(unsafe.Pointer(&a[0]))[: len(a)*4 : len(a)*4]
}

func (c *Container) Type() uint8 {
	return uint8(c.container.containerType())
}

func (c *Container) Bitmap() *Bitmap {
	return &Bitmap{
		highlowcontainer: roaringArray{
			keys:            []uint16{0},
			containers:      []container{c.clone()},
			needCopyOnWrite: []bool{false},
		},
	}
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
	// NOTE: we are not expecting r.RunOptimized being called. So only array and
	// bitmap containers are at play.
	//
	// RunOptimize should be run for older shards which no longer accept updates.
	// So we leave shrinking the database further to possibly an external tool
	// that will just iterate on old shards/containers and convert them to run
	// containers to optimize storage for historical data.
	//
	// Run containers are not ideal for our case because they are not fast like
	// array and bitmap and the storage benefits don't matter because data is
	// always compressed. Only use case is for historical data, which we leave to
	// an external tool.
	//
	// However, we still transform bitmap containers to run containers when they
	// are full. The above note is for containers which receive less updates
	var buf bytes.Buffer
	for i, c := range r.highlowcontainer.containers {
		k := r.highlowcontainer.keys[i]
		kind := c.containerType()
		buf.Reset()
		var skip bool
		err := ctx.Value(k, func(u uint8, data []byte) error {
			n := containerFromWireOwned(u, data)
			// Incoming containers are smaller. We can avoid generating unneeded writes
			// by checking if we need to update.
			diff := c.andNot(n)
			if diff.isEmpty() {
				skip = true
				return nil
			}
			diff.iterate(func(x uint16) bool {
				n = n.iaddReturnMinimized(x)
				return true
			})
			kind = n.containerType()
			// data may be invalidated by the underlying key value store. We serialize
			// while we still have lease to it.
			//
			// There is no empty containers. So we are guaranteeing buf.Len() > 0. This
			// is essential later to avoid serializing again a wrong container.
			if kind == run16Contype {
				// saves 2 bytes for bitmap runs
				buf.Write(fromInterval16(n.(*runContainer16).iv))
				return nil
			}
			_, err := n.writeTo(&buf)
			return err
		})
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		if buf.Len() == 0 {
			// This container is seen for the first time. No need to optimize for storage
			// because we are in a hot loop, eventually the container will get optimized
			// when we are doing updates.
			//
			// For the rare case we don't update the container again, compression
			// will fix the storage and for most cases this will always be an array
			// container.
			_, err = c.writeTo(&buf)
			if err != nil {
				return err
			}
		}

		err = ctx.Write(k, uint8(kind), bytes.Clone(buf.Bytes()))
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
	b = bytes.Clone(b)
	switch contype(typ) {
	case arrayContype:
		return &arrayContainer{
			byteSliceAsUint16Slice(b),
		}
	case run16Contype:
		return &runContainer16{
			iv: toInterval16(b),
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

func containerFromWireOwned(typ uint8, b []byte) container {
	switch contype(typ) {
	case arrayContype:
		return &arrayContainer{
			byteSliceAsUint16Slice(b),
		}
	case run16Contype:
		return &runContainer16{
			iv: toInterval16(b),
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
