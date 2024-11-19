package ro2

import (
	"unsafe"

	"github.com/gernest/roaring"
	"github.com/vinceanalytics/vince/internal/util/assert"
)

const (
	// ArrayMaxSize represents the maximum size of array containers.
	// This is sligtly less than roaring to accommodate the page header.
	ArrayMaxSize = 4079

	// RLEMaxSize represents the maximum size of run length encoded containers.
	RLEMaxSize = 2039
)

type ContainerType int

// Container types.
const (
	ContainerTypeArray ContainerType = 1 + iota
	ContainerTypeRLE
	ContainerTypeBitmap
)

func DecodeContainer(value []byte) *roaring.Container {
	data, typ := separate(value)
	switch typ {
	case ContainerTypeArray:
		d := toArray16(data)
		return roaring.NewContainerArray(d)
	case ContainerTypeBitmap:
		d := toArray64(data)
		return roaring.NewContainerBitmap(-1, d)
	case ContainerTypeRLE:
		d := toInterval16(data)
		return roaring.NewContainerRun(d)
	default:
		assert.Abort("invalid container type", "type", typ)
		return nil
	}
}

func EncodeContainer(c *roaring.Container) []byte {
	switch roaring.ContainerType(c) {
	case 1:
		a := roaring.AsArray(c)
		if len(a) > ArrayMaxSize {
			roaring.ConvertArrayToBitmap(c)
			return append(fromArray64(roaring.AsBitmap(c)), byte(ContainerTypeBitmap))
		}
		return append(fromArray16(a), byte(ContainerTypeArray))
	case 2:
		return append(fromArray64(roaring.AsBitmap(c)), byte(ContainerTypeBitmap))
	case 3:
		r := roaring.AsRuns(c)
		if len(r) > RLEMaxSize {
			roaring.ConvertArrayToBitmap(c)
			return append(fromArray64(roaring.AsBitmap(c)), byte(ContainerTypeBitmap))
		}
		return append(fromInterval16(r), byte(ContainerTypeRLE))
	default:
		assert.Abort("invalid container type", "type", roaring.ContainerType(c))
		return nil
	}
}

func LastValue(value []byte) uint16 {
	data, typ := separate(value)
	switch typ {
	case ContainerTypeArray:
		a := toArray16(data)
		return a[len(a)-1]
	case ContainerTypeRLE:
		r := toInterval16(data)
		return r[len(r)-1].Last
	case ContainerTypeBitmap:
		a := toArray64(data)
		return lastValueFromBitmap(a)
	default:
		return 0
	}
}

func separate(data []byte) (co []byte, typ ContainerType) {
	return data[:len(data)-1], ContainerType(data[len(data)-1])
}

func lastValueFromBitmap(a []uint64) uint16 {
	for i := len(a) - 1; i >= 0; i-- {
		for j := 63; j >= 0; j-- {
			if a[i]&(1<<j) != 0 {
				return (uint16(i) * 64) + uint16(j)
			}
		}
	}
	return 0
}

// toArray16 converts a byte slice into a slice of uint16 values using unsafe.
func toArray16(a []byte) []uint16 {
	return (*[4096]uint16)(unsafe.Pointer(&a[0]))[: len(a)/2 : len(a)/2]
}

// fromArray16 converts a slice of uint16 values into a byte slice using unsafe.
func fromArray16(a []uint16) []byte {
	return (*[8192]byte)(unsafe.Pointer(&a[0]))[: len(a)*2 : len(a)*2]
}

// toArray64 converts a byte slice into a slice of uint64 values using unsafe.
func toArray64(a []byte) []uint64 {
	return (*[1024]uint64)(unsafe.Pointer(&a[0]))[:1024:1024]
}

// fromArray64 converts a slice of uint64 values into a byte slice using unsafe.
func fromArray64(a []uint64) []byte {
	return (*[8192]byte)(unsafe.Pointer(&a[0]))[:8192:8192]
}

// toArray16 converts a byte slice into a slice of uint16 values using unsafe.
func toInterval16(a []byte) []roaring.Interval16 {
	return (*[2048]roaring.Interval16)(unsafe.Pointer(&a[0]))[: len(a)/4 : len(a)/4]
}

// fromArray16 converts a slice of uint16 values into a byte slice using unsafe.
func fromInterval16(a []roaring.Interval16) []byte {
	return (*[8192]byte)(unsafe.Pointer(&a[0]))[: len(a)*4 : len(a)*4]
}
