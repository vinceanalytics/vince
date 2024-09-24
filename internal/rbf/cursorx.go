// Copyright 2022 Molecula Corp. (DBA FeatureBase).
// SPDX-License-Identifier: Apache-2.0
package rbf

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/gernest/rbf/vprint"
	"github.com/gernest/roaring"
	"github.com/pkg/errors"
)

func (c *Cursor) Iterator() roaring.ContainerIterator {
	return &containerIterator{cursor: c}
}

func (c *Cursor) ApplyRewriter(key uint64, rewriter roaring.BitmapRewriter) (err error) {
	f := c.getContainerFilter(nil, rewriter)
	defer f.Close()
	return f.ApplyRewriter()
}

func (c *Cursor) ApplyFilter(key uint64, filter roaring.BitmapFilter) (err error) {
	_, err = c.Seek(key)
	if err != nil {
		return err
	}
	f := c.getContainerFilter(filter, nil)
	defer f.Close()
	return f.ApplyFilter()
}

func (c *Cursor) Count() (uint64, error) {
	if err := c.First(); err == io.EOF {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	var n uint64
	for {
		if err := c.Next(); err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		elem := &c.stack.elems[c.stack.top]
		leafPage, _, err := c.tx.readPage(elem.pgno)
		if err != nil {
			return 0, err
		}
		cell := readLeafCell(leafPage, elem.index)

		n += uint64(cell.BitN)
	}
	return n, nil
}

func (c *Cursor) Max() (uint64, error) {
	if err := c.Last(); err == io.EOF {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	elem := &c.stack.elems[c.stack.top]
	leafPage, _, err := c.tx.readPage(elem.pgno)
	if err != nil {
		return 0, err
	}
	cell := readLeafCell(leafPage, elem.index)
	return uint64((cell.Key << 16) | uint64(cell.lastValue(c.tx))), nil
}

func (c *Cursor) Min() (uint64, bool, error) {
	if err := c.First(); err == io.EOF {
		return 0, false, nil
	} else if err != nil {
		return 0, false, err
	}

	elem := &c.stack.elems[c.stack.top]
	leafPage, _, err := c.tx.readPage(elem.pgno)
	if err != nil {
		return 0, false, err
	}
	cell := readLeafCell(leafPage, elem.index)
	return uint64((cell.Key << 16) | uint64(cell.firstValue(c.tx))), true, nil
}

func (c *Cursor) CountRange(start, end uint64) (uint64, error) {

	if start >= end {
		return 0, nil
	}

	skey := highbits(start)
	ekey := highbits(end)
	ebits := int32(lowbits(end))

	csr := c
	exact, err := csr.Seek(skey)
	_ = exact
	if err == io.EOF {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	var n uint64
	for {
		if err := csr.Next(); err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		elem := &csr.stack.elems[csr.stack.top]
		leafPage, _, err := csr.tx.readPage(elem.pgno)
		if err != nil {
			return 0, err
		}
		c := readLeafCell(leafPage, elem.index)

		k := c.Key
		if k > ekey {
			break
		}

		// If range is entirely in one container then just count that range.
		if skey == ekey {
			return uint64(c.countRange(csr.tx, int32(lowbits(start)), ebits)), nil
		}
		// INVAR: skey < ekey

		// k > ekey handles the case when start > end and where start and end
		// are in different containers. Same container case is already handled above.
		if k > ekey {
			break
		}
		if k == skey {
			n += uint64(c.countRange(csr.tx, int32(lowbits(start)), roaring.MaxContainerVal+1))
			continue
		}
		if k < ekey {
			n += uint64(c.BitN)
			continue
		}
		if k == ekey && ebits > 0 {
			n += uint64(c.countRange(csr.tx, 0, ebits))
			break
		}
	}
	return n, nil
}

func (c *Cursor) OffsetRange(offset, start, endx uint64) (*roaring.Bitmap, error) {
	if lowbits(offset) != 0 {
		vprint.PanicOn("offset must not contain low bits")
	} else if lowbits(start) != 0 {
		vprint.PanicOn("range start must not contain low bits")
	} else if lowbits(endx) != 0 {
		vprint.PanicOn("range endx must not contain low bits")
	}

	other := roaring.NewSliceBitmap()
	off := highbits(offset)
	hi0, hi1 := highbits(start), highbits(endx)

	if _, err := c.Seek(hi0); err == io.EOF {
		return other, nil
	} else if err != nil {
		return nil, err
	}

	for {
		if err := c.Next(); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		elem := &c.stack.elems[c.stack.top]
		leafPage, _, err := c.tx.readPage(elem.pgno)
		if err != nil {
			return nil, err
		}
		cell := readLeafCell(leafPage, elem.index)
		ckey := cell.Key

		// >= hi1 is correct b/c endx cannot have any lowbits set.
		if ckey >= hi1 {
			break
		}
		other.Containers.Put(off+(ckey-hi0), toContainer(cell, c.tx))
	}
	return other, nil
}

// probably should just implement the container interface
// but for now i'll do it
func (c *Cursor) Rows() ([]uint64, error) {
	shardVsContainerExponent := uint(4) // needs constant exported from roaring package
	rows := make([]uint64, 0)
	if err := c.First(); err != nil {
		if err == io.EOF { // root leaf with no elements
			return rows, nil
		}
		return nil, errors.Wrap(err, "rows")
	}
	var err error
	var lastRow uint64 = math.MaxUint64
	for {
		err := c.Next()
		if err != nil {
			break
		}

		elem := &c.stack.elems[c.stack.top]
		leafPage, _, err := c.tx.readPage(elem.pgno)
		if err != nil {
			return nil, err
		}
		cell := readLeafCell(leafPage, elem.index)

		vRow := cell.Key >> shardVsContainerExponent
		if vRow == lastRow {
			continue
		}
		rows = append(rows, vRow)
		lastRow = vRow
	}
	return rows, err
}

func (tx *Tx) FieldViews() []string {
	records, _ := tx.RootRecords()
	a := make([]string, 0, records.Len())
	for itr := records.Iterator(); !itr.Done(); {
		name, _, _ := itr.Next()
		a = append(a, name)
	}
	return a
}

func (c *Cursor) DumpKeys() {
	if err := c.First(); err != nil {
		// ignoring errors for this debug function
		return
	}
	for {
		err := c.Next()
		if err == io.EOF {
			break
		}
		elem := &c.stack.elems[c.stack.top]
		leafPage, _, err := c.tx.readPage(elem.pgno)
		if err != nil {
			return
		}
		cell := readLeafCell(leafPage, elem.index)
		fmt.Println("key", cell.Key)
	}
}

func (c *Cursor) DumpStack() {
	fmt.Println("STACK")
	for i := c.stack.top; i >= 0; i-- {
		fmt.Printf("%+v\n", c.stack.elems[i])
	}
	fmt.Println()
}

func (c *Cursor) Dump(name string) {
	writer, _ := os.Create(name)
	defer writer.Close()
	bufStdout := bufio.NewWriter(writer)
	fmt.Fprintf(bufStdout, "digraph RBF{\n")
	fmt.Fprintf(bufStdout, "rankdir=\"LR\"\n")

	fmt.Fprintf(bufStdout, "node [shape=record height=.1]\n")
	Dumpdot(c.tx, 0, " ", bufStdout)
	fmt.Fprintf(bufStdout, "\n}")
	bufStdout.Flush()
}

func (c *Cursor) Row(shard, rowID uint64) (*roaring.Bitmap, error) {
	base := rowID * ShardWidth

	offset := uint64(shard * ShardWidth)
	off := highbits(offset)
	hi0, hi1 := highbits(base), highbits((rowID+1)*ShardWidth)
	c.stack.top = 0
	ok, err := c.Seek(hi0)
	if err != nil {
		return nil, err
	}
	if !ok {
		elem := &c.stack.elems[c.stack.top]
		leafPage, _, err := c.tx.readPage(elem.pgno)
		if err != nil {
			return nil, err
		}
		n := readCellN(leafPage)
		if elem.index >= n {
			if err := c.goNextPage(); err != nil {
				return nil, errors.Wrap(err, "row")
			}
		}
	}
	other := roaring.NewSliceBitmap()
	for {
		err := c.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		elem := &c.stack.elems[c.stack.top]
		leafPage, _, err := c.tx.readPage(elem.pgno)
		if err != nil {
			return nil, err
		}
		cell := readLeafCell(leafPage, elem.index)
		if cell.Key >= hi1 {
			break
		}
		other.Containers.Put(off+(cell.Key-hi0), toContainer(cell, c.tx))
	}
	return other, nil
}

// CurrentPageType returns the type of the container currently pointed to by cursor used in testing
// sometimes the cursor needs to be positions prior to this call with First/Last etc.
func (c *Cursor) CurrentPageType() ContainerType {
	elem := &c.stack.elems[c.stack.top]
	leafPage, _, _ := c.tx.readPage(elem.pgno)
	cell := readLeafCell(leafPage, elem.index)
	return cell.Type
}

func intoContainer(l leafCell, tx *Tx, replacing *roaring.Container, target []byte) (c *roaring.Container) {
	if len(l.Data) == 0 {
		return nil
	}
	orig := l.Data
	var cpMaybe []byte
	var mapped bool
	if tx.db.cfg.DoAllocZero {
		// make a copy so no one will see corrupted data
		// or mmapped data that may disappear.
		cpMaybe = target[:len(orig)]
		copy(cpMaybe, orig)
		mapped = false
	} else {
		// not a copy
		cpMaybe = orig
		mapped = true
	}
	switch l.Type {
	case ContainerTypeArray:
		c = roaring.RemakeContainerArray(replacing, toArray16(cpMaybe))
	case ContainerTypeBitmapPtr:
		_, bm, _ := tx.leafCellBitmap(toPgno(cpMaybe))
		cloneMaybe := bm
		c = roaring.RemakeContainerBitmapN(replacing, cloneMaybe, int32(l.BitN))
	case ContainerTypeBitmap:
		c = roaring.RemakeContainerBitmapN(replacing, toArray64(cpMaybe), int32(l.BitN))
	case ContainerTypeRLE:
		c = roaring.RemakeContainerRunN(replacing, toInterval16(cpMaybe), int32(l.BitN))
	}
	// Note: If the "roaringparanoia" build tag isn't set, this
	// should be optimized away entirely. Otherwise it's moderately
	// expensive.
	c.CheckN()
	c.SetMapped(mapped)
	return c
}

// intoWritableContainer always uses the provided target for a copy of
// the container's contents, so the container can be modified safely.
func intoWritableContainer(l leafCell, tx *Tx, replacing *roaring.Container, target []byte) (c *roaring.Container, err error) {
	if len(l.Data) == 0 {
		return nil, nil
	}
	orig := l.Data
	target = target[:len(orig)]
	copy(target, orig)
	switch l.Type {
	case ContainerTypeArray:
		c = roaring.RemakeContainerArray(replacing, toArray16(target))
	case ContainerTypeBitmapPtr:
		pgno := toPgno(target)
		target = target[:PageSize] // reslice back to full size
		_, bm, err := tx.leafCellBitmapInto(pgno, target)
		if err != nil {
			return nil, fmt.Errorf("intoContainer: %s", err)
		}
		c = roaring.RemakeContainerBitmapN(replacing, bm, int32(l.BitN))
	case ContainerTypeBitmap:
		c = roaring.RemakeContainerBitmapN(replacing, toArray64(target), int32(l.BitN))
	case ContainerTypeRLE:
		c = roaring.RemakeContainerRunN(replacing, toInterval16(target), int32(l.BitN))
	}
	// Note: If the "roaringparanoia" build tag isn't set, this
	// should be optimized away entirely. Otherwise it's moderately
	// expensive.
	c.CheckN()
	c.SetMapped(false)
	return c, nil
}

func toContainer(l leafCell, tx *Tx) (c *roaring.Container) {
	if len(l.Data) == 0 {
		return nil
	}
	orig := l.Data
	var cpMaybe []byte
	var mapped bool
	if tx.db.cfg.DoAllocZero {
		// make a copy, otherwise someone could see corrupted data
		// or mmapped data that may disappear.
		cpMaybe = make([]byte, len(orig))
		copy(cpMaybe, orig)
		mapped = false
	} else {
		// not a copy
		cpMaybe = orig
		mapped = true
	}
	switch l.Type {
	case ContainerTypeArray:
		c = roaring.NewContainerArray(toArray16(cpMaybe))
	case ContainerTypeBitmapPtr:
		_, bm, _ := tx.leafCellBitmap(toPgno(cpMaybe))
		cloneMaybe := bm
		c = roaring.NewContainerBitmap(l.BitN, cloneMaybe)
	case ContainerTypeBitmap:
		c = roaring.NewContainerBitmap(l.BitN, toArray64(cpMaybe))
	case ContainerTypeRLE:
		c = roaring.NewContainerRunN(toInterval16(cpMaybe), int32(l.BitN))
	}
	// Note: If the "roaringparanoia" build tag isn't set, this
	// should be optimized away entirely. Otherwise it's moderately
	// expensive.
	c.CheckN()
	c.SetMapped(mapped)
	return c
}

type Nodetype int

const (
	Branch Nodetype = iota
	Leaf
	Bitmap
)

type Walker interface {
	Visitor(pgno uint32, records []*RootRecord)
	VisitRoot(pgno uint32, name string)
	Visit(pgno uint32, n Nodetype)
}

func WalkPage(tx *Tx, pgno uint32, walker Walker) {
	page, _, err := tx.readPage(pgno)
	if err != nil {
		panic(err)
	}

	if IsMetaPage(page) {
		Walk(tx, readMetaRootRecordPageNo(page), walker.Visitor)
		return
	}

	// Handle
	switch typ := readFlags(page); typ {
	case PageTypeBranch:
		walker.Visit(pgno, Branch)
		for i, n := 0, readCellN(page); i < n; i++ {
			cell := readBranchCell(page, i)
			WalkPage(tx, cell.ChildPgno, walker)
		}
	case PageTypeLeaf:
		walker.Visit(pgno, Leaf)
	default:
		panic(err)
	}
}
