/*
 * Copyright 2020 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package btree

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

var (
	pageSize = os.Getpagesize()
	maxKeys  = (pageSize / 16) - 1
)

const (
	absoluteMax = uint64(math.MaxUint64 - 1)
	minSize     = 1 << 20
)

// Tree represents the structure for custom mmaped B+ tree.
// It supports keys in range [1, math.MaxUint64-1] and values [1, math.Uint64].
type Tree[T Source] struct {
	base     Source
	data     []byte
	nextPage uint64
	freePage uint64
	stats    TreeStats
}

func (t *Tree[Data]) initRootNode() {
	// This is the root node.
	t.newNode(0)
	// This acts as the rightmost pointer (all the keys are <= this key).
	t.Set(absoluteMax, 0)
}

type SliceTree = Tree[*Slice]

// NewTree returns an in-memory B+ tree.
func NewTree() *SliceTree {
	return NewWith(new(Slice))
}

func NewWith[T Source](data T) *Tree[T] {
	t := &Tree[T]{base: data}
	t.Init()
	return t
}

func (t *Tree[Data]) Init() {
	t.Reset()
}

func (t *Tree[Data]) reinit() {
	// Calculate t.nextPage by finding the first node whose pageID is not set.
	t.nextPage = 1
	for int(t.nextPage)*pageSize < len(t.data) {
		n := t.node(t.nextPage)
		if n.pageID() == 0 {
			break
		}
		t.nextPage++
	}
	maxPageID := t.nextPage - 1

	// Calculate t.freePage by finding the page to which no other page points.
	// This would be the head of the page linked list.
	// tailPages[i] is true if pageId i+1 is not the head of the list.
	tailPages := make([]bool, maxPageID)
	// Mark all pages containing nodes as tail pages.
	t.Iterate(func(n node) {
		i := n.pageID() - 1
		tailPages[i] = true
		// If this is a leaf node, increment the stats.
		if n.isLeaf() {
			t.stats.NumLeafKeys += n.numKeys()
		}
	})
	// pointedPages is a list of page IDs that the tail pages point to.
	pointedPages := make([]uint64, 0)
	for i, isTail := range tailPages {
		if !isTail {
			pageID := uint64(i) + 1
			// Skip if nextPageID = 0, as that is equivalent to null page.
			if nextPageID := t.node(pageID).uint64(0); nextPageID != 0 {
				pointedPages = append(pointedPages, nextPageID)
			}
			t.stats.NumPagesFree++
		}
	}

	// Mark all pages being pointed to as tail pages.
	for _, pageID := range pointedPages {
		i := pageID - 1
		tailPages[i] = true
	}
	// There should only be one head page left.
	for i, isTail := range tailPages {
		if !isTail {
			pageID := uint64(i) + 1
			t.freePage = pageID
			break
		}
	}
}

// Reset resets the tree and truncates it to maxSz.
func (t *Tree[Data]) Reset() {
	t.data = t.base.Reset()
	t.stats = TreeStats{}
	t.nextPage = 1
	t.freePage = 0
	t.initRootNode()
}

// Close releases the memory used by the tree.
func (t *Tree[Data]) Close() error {
	return t.base.Release()
}

type TreeStats struct {
	Allocated    int     // Derived.
	Bytes        int     // Derived.
	NumLeafKeys  int     // Calculated.
	NumPages     int     // Derived.
	NumPagesFree int     // Calculated.
	Occupancy    float64 // Derived.
	PageSize     int     // Derived.
}

// Stats returns stats about the tree.
func (t *Tree[Data]) Stats() TreeStats {
	numPages := int(t.nextPage - 1)
	out := TreeStats{
		Bytes:        numPages * pageSize,
		Allocated:    len(t.data),
		NumLeafKeys:  t.stats.NumLeafKeys,
		NumPages:     numPages,
		NumPagesFree: t.stats.NumPagesFree,
		PageSize:     pageSize,
	}
	out.Occupancy = 100.0 * float64(out.NumLeafKeys) / float64(maxKeys*numPages)
	return out
}

func (t *Tree[Data]) newNode(bit uint64) node {
	var pageID uint64
	if t.freePage > 0 {
		pageID = t.freePage
		t.stats.NumPagesFree--
	} else {
		pageID = t.nextPage
		t.nextPage++
		offset := int(pageID) * pageSize
		reqSize := offset + pageSize
		if reqSize > len(t.data) {
			t.data = t.base.AllocateOffset(reqSize - len(t.data))
		}
	}
	n := t.node(pageID)
	if t.freePage > 0 {
		t.freePage = n.uint64(0)
	}
	zeroOut(n)
	n.setBit(bit)
	n.setAt(keyOffset(maxKeys), pageID)
	return n
}

func getNode(data []byte) node {
	return node(bytesToUint64Slice(data))
}

func bytesToUint64Slice(b []byte) []uint64 {
	if len(b) == 0 {
		return nil
	}
	var u64s []uint64
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&u64s))
	hdr.Len = len(b) / 8
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return u64s
}

func zeroOut(data []uint64) {
	for i := 0; i < len(data); i++ {
		data[i] = 0
	}
}

func (t *Tree[Data]) node(pid uint64) node {
	// page does not exist
	if pid == 0 {
		return nil
	}
	start := pageSize * int(pid)
	return getNode(t.data[start : start+pageSize])
}

// Set sets the key-value pair in the tree.
func (t *Tree[Data]) Set(k, v uint64) {
	if k == math.MaxUint64 || k == 0 {
		panic("util/btree: error setting zero or MaxUint64")
	}
	root := t.set(1, k, v)
	if root.isFull() {
		right := t.split(1)
		left := t.newNode(root.bits())
		// Re-read the root as the underlying buffer for tree might have changed during split.
		root = t.node(1)
		copy(left[:keyOffset(maxKeys)], root)
		left.setNumKeys(root.numKeys())

		// reset the root node.
		zeroOut(root[:keyOffset(maxKeys)])
		root.setNumKeys(0)

		// set the pointers for left and right child in the root node.
		root.set(left.maxKey(), left.pageID())
		root.set(right.maxKey(), right.pageID())
	}
}

// For internal nodes, they contain <key, ptr>.
// where all entries <= key are stored in the corresponding ptr.
func (t *Tree[Data]) set(pid, k, v uint64) node {
	n := t.node(pid)
	if n.isLeaf() {
		t.stats.NumLeafKeys += n.set(k, v)
		return n
	}

	// This is an internal node.
	idx := n.search(k)
	if idx >= maxKeys {
		panic("util/btree: search returned index >= maxKeys")
	}
	// If no key at idx.
	if n.key(idx) == 0 {
		n.setAt(keyOffset(idx), k)
		n.setNumKeys(n.numKeys() + 1)
	}
	child := t.node(n.val(idx))
	if child == nil {
		child = t.newNode(bitLeaf)
		n = t.node(pid)
		n.setAt(valOffset(idx), child.pageID())
	}
	child = t.set(child.pageID(), k, v)
	// Re-read n as the underlying buffer for tree might have changed during set.
	n = t.node(pid)
	if child.isFull() {
		// Just consider the left sibling for simplicity.
		// if t.shareWithSibling(n, idx) {
		// 	return n
		// }

		nn := t.split(child.pageID())
		// Re-read n and child as the underlying buffer for tree might have changed during split.
		n = t.node(pid)
		child = t.node(n.uint64(valOffset(idx)))
		// Set child pointers in the node n.
		// Note that key for right node (nn) already exist in node n, but the
		// pointer is updated.
		n.set(child.maxKey(), child.pageID())
		n.set(nn.maxKey(), nn.pageID())
	}
	return n
}

func (t *Tree[Data]) Incr(k uint64) uint64 {
	if k == math.MaxUint64 || k == 0 {
		panic("util/btree: error incr zero or MaxUint64")
	}
	root, newVal := t.incr(1, k)
	if root.isFull() {
		right := t.split(1)
		left := t.newNode(root.bits())
		// Re-read the root as the underlying buffer for tree might have changed during split.
		root = t.node(1)
		copy(left[:keyOffset(maxKeys)], root)
		left.setNumKeys(root.numKeys())

		// reset the root node.
		zeroOut(root[:keyOffset(maxKeys)])
		root.setNumKeys(0)

		// set the pointers for left and right child in the root node.
		root.set(left.maxKey(), left.pageID())
		root.set(right.maxKey(), right.pageID())
	}
	return newVal
}

func (t *Tree[Data]) incr(pid, k uint64) (node, uint64) {
	n := t.node(pid)
	if n.isLeaf() {
		numleaf, newVal := n.incr(k)
		t.stats.NumLeafKeys += numleaf
		return n, newVal
	}

	// This is an internal node.
	idx := n.search(k)
	if idx >= maxKeys {
		panic("util/btree: search returned index >= maxKeys")
	}
	// If no key at idx.
	if n.key(idx) == 0 {
		n.setAt(keyOffset(idx), k)
		n.setNumKeys(n.numKeys() + 1)
	}
	child := t.node(n.val(idx))
	if child == nil {
		child = t.newNode(bitLeaf)
		n = t.node(pid)
		n.setAt(valOffset(idx), child.pageID())
	}
	child, newVal := t.incr(child.pageID(), k)
	// Re-read n as the underlying buffer for tree might have changed during set.
	n = t.node(pid)
	if child.isFull() {
		// Just consider the left sibling for simplicity.
		// if t.shareWithSibling(n, idx) {
		// 	return n
		// }

		nn := t.split(child.pageID())
		// Re-read n and child as the underlying buffer for tree might have changed during split.
		n = t.node(pid)
		child = t.node(n.uint64(valOffset(idx)))
		// Set child pointers in the node n.
		// Note that key for right node (nn) already exist in node n, but the
		// pointer is updated.
		n.set(child.maxKey(), child.pageID())
		n.set(nn.maxKey(), nn.pageID())
	}
	return n, newVal
}

// Get looks for key and returns the corresponding value.
// If key is not found, 0 is returned.
func (t *Tree[Data]) Get(k uint64) uint64 {
	if k == math.MaxUint64 || k == 0 {
		panic("util/btree: does not support getting MaxUint64/Zero")
	}
	root := t.node(1)
	return t.get(root, k)
}

func (t *Tree[Data]) get(n node, k uint64) uint64 {
	if n.isLeaf() {
		return n.get(k)
	}
	// This is internal node
	idx := n.search(k)
	if idx == n.numKeys() || n.key(idx) == 0 {
		return 0
	}
	child := t.node(n.uint64(valOffset(idx)))
	assert(child != nil)
	return t.get(child, k)
}

// DeleteBelow deletes all keys with value under ts.
func (t *Tree[Data]) DeleteBelow(ts uint64) {
	root := t.node(1)
	t.stats.NumLeafKeys = 0
	t.compact(root, ts)
	assert(root.numKeys() >= 1)
}

func (t *Tree[Data]) compact(n node, ts uint64) int {
	if n.isLeaf() {
		numKeys := n.compact(ts)
		t.stats.NumLeafKeys += n.numKeys()
		return numKeys
	}
	// Not leaf.
	N := n.numKeys()
	for i := 0; i < N; i++ {
		assert(n.key(i) > 0)
		childID := n.uint64(valOffset(i))
		child := t.node(childID)
		if rem := t.compact(child, ts); rem == 0 && i < N-1 {
			// If no valid key is remaining we can drop this child. However, don't do that if this
			// is the max key.
			t.stats.NumLeafKeys -= child.numKeys()
			child.setAt(0, t.freePage)
			t.freePage = childID
			n.setAt(valOffset(i), 0)
			t.stats.NumPagesFree++
		}
	}
	// We use ts=1 here because we want to delete all the keys whose value is 0, which means they no
	// longer have a valid page for that key.
	return n.compact(1)
}

func (t *Tree[Data]) iterate(n node, fn func(node)) {
	fn(n)
	if n.isLeaf() {
		return
	}
	// Explore children.
	for i := 0; i < maxKeys; i++ {
		if n.key(i) == 0 {
			return
		}
		childID := n.uint64(valOffset(i))
		assert(childID > 0)

		child := t.node(childID)
		t.iterate(child, fn)
	}
}

// Iterate iterates over the tree and executes the fn on each node.
func (t *Tree[Data]) Iterate(fn func(node)) {
	root := t.node(1)
	t.iterate(root, fn)
}

// IterateKV iterates through all keys and values in the tree.
// If newVal is non-zero, it will be set in the tree.
func (t *Tree[Data]) IterateKV(f func(key, val uint64) (newVal uint64)) {
	t.Iterate(func(n node) {
		// Only leaf nodes contain keys.
		if !n.isLeaf() {
			return
		}

		for i := 0; i < n.numKeys(); i++ {
			key := n.key(i)
			val := n.val(i)

			// A zero value here means that this is a bogus entry.
			if val == 0 {
				continue
			}

			newVal := f(key, val)
			if newVal != 0 {
				n.setAt(valOffset(i), newVal)
			}
		}
	})
}

func (t *Tree[Data]) print(n node, parentID uint64) {
	n.print(parentID)
	if n.isLeaf() {
		return
	}
	pid := n.pageID()
	for i := 0; i < maxKeys; i++ {
		if n.key(i) == 0 {
			return
		}
		childID := n.uint64(valOffset(i))
		child := t.node(childID)
		t.print(child, pid)
	}
}

// Print iterates over the tree and prints all valid KVs.
func (t *Tree[Data]) Print() {
	root := t.node(1)
	t.print(root, 0)
}

// Splits the node into two. It moves right half of the keys from the original node to a newly
// created right node. It returns the right node.
func (t *Tree[Data]) split(pid uint64) node {
	n := t.node(pid)
	if !n.isFull() {
		panic("util/btree: split should be called only when n is full")
	}

	// Create a new node nn, copy over half the keys from n, and set the parent to n's parent.
	nn := t.newNode(n.bits())
	// Re-read n as the underlying buffer for tree might have changed during newNode.
	n = t.node(pid)
	rightHalf := n[keyOffset(maxKeys/2):keyOffset(maxKeys)]
	copy(nn, rightHalf)
	nn.setNumKeys(maxKeys - maxKeys/2)

	// Remove entries from node n.
	zeroOut(rightHalf)
	n.setNumKeys(maxKeys / 2)
	return nn
}

// | pageID (8 bytes) | metaBits (1 byte) | 3 free bytes | numKeys (4 bytes) |.
type node []uint64

func (n node) uint64(start int) uint64 { return n[start] }

// func (n node) uint32(start int) uint32 { return *(*uint32)(unsafe.Pointer(&n[start])) }

func keyOffset(i int) int          { return 2 * i }
func valOffset(i int) int          { return 2*i + 1 }
func (n node) numKeys() int        { return int(n.uint64(valOffset(maxKeys)) & 0xFFFFFFFF) }
func (n node) pageID() uint64      { return n.uint64(keyOffset(maxKeys)) }
func (n node) key(i int) uint64    { return n.uint64(keyOffset(i)) }
func (n node) val(i int) uint64    { return n.uint64(valOffset(i)) }
func (n node) data(i int) []uint64 { return n[keyOffset(i):keyOffset(i+1)] }

func (n node) setAt(start int, k uint64) {
	n[start] = k
}

func (n node) incrAt(start int) uint64 {
	n[start]++
	return n[start]
}

func (n node) setNumKeys(num int) {
	idx := valOffset(maxKeys)
	val := n[idx]
	val &= 0xFFFFFFFF00000000
	val |= uint64(num)
	n[idx] = val
}

func (n node) moveRight(lo int) {
	hi := n.numKeys()
	assert(hi != maxKeys)
	// copy works despite of overlap in src and dst.
	// See https://golang.org/pkg/builtin/#copy
	copy(n[keyOffset(lo+1):keyOffset(hi+1)], n[keyOffset(lo):keyOffset(hi)])
}

const (
	bitLeaf = uint64(1 << 63)
)

func (n node) setBit(b uint64) {
	vo := valOffset(maxKeys)
	val := n[vo]
	val &= 0xFFFFFFFF
	val |= b
	n[vo] = val
}

func (n node) bits() uint64 {
	return n.val(maxKeys) & 0xFF00000000000000
}

func (n node) isLeaf() bool {
	return n.bits()&bitLeaf > 0
}

// isFull checks that the node is already full.
func (n node) isFull() bool {
	return n.numKeys() == maxKeys
}

// Search returns the index of a smallest key >= k in a node.
func (n node) search(k uint64) int {
	N := n.numKeys()
	if N < 4 {
		for i := 0; i < N; i++ {
			if ki := n.key(i); ki >= k {
				return i
			}
		}
		return N
	}
	lo, hi := 0, N
	// Reduce the search space using binary search and then do linear search.
	for hi-lo > 32 {
		mid := (hi + lo) / 2
		km := n.key(mid)
		if k == km {
			return mid
		}
		if k > km {
			// key is greater than the key at mid, so move right.
			lo = mid + 1
		} else {
			// else move left.
			hi = mid
		}
	}
	for i := lo; i <= hi; i++ {
		if ki := n.key(i); ki >= k {
			return i
		}
	}
	return N
}

func (n node) maxKey() uint64 {
	idx := n.numKeys()
	// idx points to the first key which is zero.
	if idx > 0 {
		idx--
	}
	return n.key(idx)
}

// compacts the node i.e., remove all the kvs with value < lo. It returns the remaining number of
// keys.
func (n node) compact(lo uint64) int {
	N := n.numKeys()
	mk := n.maxKey()
	var left, right int
	for right = 0; right < N; right++ {
		if n.val(right) < lo && n.key(right) < mk {
			// Skip over this key. Don't copy it.
			continue
		}
		// Valid data. Copy it from right to left. Advance left.
		if left != right {
			copy(n.data(left), n.data(right))
		}
		left++
	}
	// zero out rest of the kv pairs.
	zeroOut(n[keyOffset(left):keyOffset(right)])
	n.setNumKeys(left)

	// If the only key we have is the max key, and its value is less than lo, then we can indicate
	// to the caller by returning a zero that it's OK to drop the node.
	if left == 1 && n.key(0) == mk && n.val(0) < lo {
		return 0
	}
	return left
}

func (n node) get(k uint64) uint64 {
	idx := n.search(k)
	// key is not found
	if idx == n.numKeys() {
		return 0
	}
	if ki := n.key(idx); ki == k {
		return n.val(idx)
	}
	return 0
}

// set returns true if it added a new key.
func (n node) set(k, v uint64) (numAdded int) {
	idx := n.search(k)
	ki := n.key(idx)
	if n.numKeys() == maxKeys {
		// This happens during split of non-root node, when we are updating the child pointer of
		// right node. Hence, the key should already exist.
		assert(ki == k)
	}
	if ki > k {
		// Found the first entry which is greater than k. So, we need to fit k
		// just before it. For that, we should move the rest of the data in the
		// node to the right to make space for k.
		n.moveRight(idx)
	}
	// If the k does not exist already, increment the number of keys.
	if ki != k {
		n.setNumKeys(n.numKeys() + 1)
		numAdded = 1
	}
	if ki == 0 || ki >= k {
		n.setAt(keyOffset(idx), k)
		n.setAt(valOffset(idx), v)
		return
	}
	panic("util/btree: unreachable")
}

func (n node) incr(k uint64) (numAdded int, newVal uint64) {
	idx := n.search(k)
	ki := n.key(idx)
	if n.numKeys() == maxKeys {
		// This happens during split of non-root node, when we are updating the child pointer of
		// right node. Hence, the key should already exist.
		assert(ki == k)
	}
	if ki > k {
		// Found the first entry which is greater than k. So, we need to fit k
		// just before it. For that, we should move the rest of the data in the
		// node to the right to make space for k.
		n.moveRight(idx)
	}
	// If the k does not exist already, increment the number of keys.
	if ki != k {
		n.setNumKeys(n.numKeys() + 1)
		numAdded = 1
	}
	if ki == 0 || ki >= k {
		n.setAt(keyOffset(idx), k)
		newVal = n.incrAt(valOffset(idx))
		return
	}
	panic("util/btree: unreachable")
}

func (n node) iterate(fn func(node, int)) {
	for i := 0; i < maxKeys; i++ {
		if k := n.key(i); k > 0 {
			fn(n, i)
		} else {
			break
		}
	}
}

func (n node) print(parentID uint64) {
	var keys []string
	n.iterate(func(n node, i int) {
		keys = append(keys, strconv.FormatUint(n.key(i), 10))
	})
	if len(keys) > 8 {
		copy(keys[4:], keys[len(keys)-4:])
		keys[3] = "..."
		keys = keys[:8]
	}
	fmt.Printf("%d Child of: %d num keys: %d keys: %s\n",
		n.pageID(), parentID, n.numKeys(), strings.Join(keys, " "))
}

func assert(ok bool) {
	if !ok {
		panic("util/btree: failed assertion")
	}
}
