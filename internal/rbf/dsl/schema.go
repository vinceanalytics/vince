package dsl

import (
	"fmt"
	"math"
	"slices"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/bsi"
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	ID = "_id"
)

// Schema maps proto fields to rbf types.
type Schema[T proto.Message] struct {
	ids    []uint64
	rowIDs [][]uint64
	values [][]int64

	keys   [][]string
	trKeys [][]uint64
	sets   [][][]string
	trSets [][][]uint64

	blobs     [][][]byte
	trBlobs   [][]uint64
	blobSets  [][][][]byte
	trBlobSet [][][]uint64

	mapping map[string]int
	bsi     map[string]struct{}
}

func NewSchema[T proto.Message](bsi ...string) (*Schema[T], error) {
	var a T

	rs := &Schema[T]{
		mapping: make(map[string]int),
		bsi:     make(map[string]struct{}),
	}
	for i := range bsi {
		rs.bsi[bsi[i]] = struct{}{}
	}

	fields := a.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		name := string(f.Name())
		if f.IsList() {
			// only []string  and [][]byte is supported
			switch f.Kind() {
			case protoreflect.StringKind:
				pos := len(rs.sets)
				rs.sets = append(rs.sets, nil)
				rs.trSets = append(rs.trSets, nil)
				rs.mapping[name] = pos
			case protoreflect.BytesKind:
				pos := len(rs.blobSets)
				rs.blobSets = append(rs.blobSets, nil)
				rs.trBlobSet = append(rs.trBlobSet, nil)
				rs.mapping[name] = pos
			default:
				return nil, fmt.Errorf("%s list is not supported", f.Kind())
			}
			continue
		}
		switch f.Kind() {
		case protoreflect.BoolKind,
			protoreflect.EnumKind:
			pos := len(rs.rowIDs)
			rs.rowIDs = append(rs.rowIDs, nil)
			rs.mapping[name] = pos

		case protoreflect.Int64Kind,
			protoreflect.Uint64Kind,
			protoreflect.DoubleKind:
			pos := len(rs.values)
			rs.values = append(rs.values, nil)
			rs.mapping[name] = pos

		case protoreflect.StringKind:
			pos := len(rs.keys)
			rs.keys = append(rs.keys, nil)
			rs.trKeys = append(rs.trKeys, nil)
			rs.mapping[name] = pos
		case protoreflect.BytesKind:
			pos := len(rs.blobs)
			rs.blobs = append(rs.blobs, nil)
			rs.trBlobs = append(rs.trBlobs, nil)
			rs.mapping[name] = pos
		default:
			return nil, fmt.Errorf("%q %s is not supported", name, f.Kind())
		}
	}
	return rs, nil
}

func (s *Schema[T]) Reset() {
	s.ids = s.ids[:0]
	reset(s.rowIDs)
	reset(s.values)

	resetClear(s.keys)
	reset(s.trKeys)
	resetClear(s.sets)
	reset(s.trSets)

	resetClear(s.blobs)
	reset(s.trBlobs)
	resetClear(s.blobSets)
	reset(s.trBlobSet)
}

func reset[T any](ls [][]T) {
	for i := range ls {
		ls[i] = ls[i][:0]
	}
}

func resetClear[T any](ls [][]T) {
	for i := range ls {
		clear(ls[i])
		ls[i] = ls[i][:0]
	}
}

func (s *Schema[T]) Write(msg T) {
	// We generate ids later on when applying the schema
	s.ids = append(s.ids, 0)
	r := msg.ProtoReflect()
	fs := r.Descriptor().Fields()
	for i := 0; i < fs.Len(); i++ {
		fd := fs.Get(i)
		v := r.Get(fd)
		name := string(fd.Name())
		pos := s.mapping[name]
		switch fd.Kind() {
		case protoreflect.BoolKind:
			value := uint64(0)
			if v.Bool() {
				value = 1
			}
			s.rowIDs[pos] = append(s.rowIDs[pos], value)
		case protoreflect.EnumKind:
			s.rowIDs[pos] = append(s.rowIDs[pos], uint64(v.Enum()))
		case protoreflect.Int64Kind:
			s.values[pos] = append(s.values[pos], v.Int())
		case protoreflect.Uint64Kind:
			s.values[pos] = append(s.values[pos], int64(v.Uint()))
		case protoreflect.DoubleKind:
			s.values[pos] = append(s.values[pos], int64(math.Float64bits(v.Float())))
		case protoreflect.StringKind:
			if fd.IsList() {
				ls := v.List()
				vs := []string{}
				if ls.Len() != 0 {
					vs = make([]string, 0, ls.Len())
					for n := range ls.Len() {
						vs = append(vs, ls.Get(n).String())
					}
				}
				s.sets[pos] = append(s.sets[pos], vs)
			} else {
				s.keys[pos] = append(s.keys[pos], v.String())
			}
		case protoreflect.BytesKind:
			if fd.IsList() {
				ls := v.List()
				vs := [][]byte{}
				if ls.Len() != 0 {
					vs = make([][]byte, 0, ls.Len())
					for n := range ls.Len() {
						vs = append(vs, ls.Get(n).Bytes())
					}
				}
				s.blobSets[pos] = append(s.blobSets[pos], vs)
			} else {
				s.blobs[pos] = append(s.blobs[pos], v.Bytes())
			}
		}
	}
}

func (s *Schema[T]) process(db *Store[T]) error {
	defer s.Reset()

	w, err := db.ops.write()
	if err != nil {
		return err
	}
	defer w.Release()

	// generate ids
	err = w.fill(s.ids)
	if err != nil {
		return err
	}
	var msg T
	fields := msg.ProtoReflect().Descriptor().Fields()
	for fi := 0; fi < fields.Len(); fi++ {
		f := fields.Get(fi)
		name := string(f.Name())
		pos := s.mapping[name]
		switch f.Kind() {
		case protoreflect.BoolKind, protoreflect.EnumKind:
			x := s.rowIDs[pos]
			for i := range s.ids {
				x[i] = (x[i] * shardwidth.ShardWidth) + (s.ids[i] % shardwidth.ShardWidth)
			}
		case protoreflect.Int64Kind, protoreflect.DoubleKind:
			// need special handling, delay this to the next per shard iteration

		case protoreflect.StringKind:
			st, err := w.tr.String(name)
			if err != nil {
				return err
			}
			if f.IsList() {
				s.trSets[pos] = adjustSet(s.trSets[pos], s.sets[pos])
				err = st.BulkSet(s.sets[pos], s.trSets[pos])
				if err != nil {
					return err
				}
				x := s.trSets[pos]
				for i := range s.ids {
					for j := range x[i] {
						x[i][j] = (x[i][j] * shardwidth.ShardWidth) + (s.ids[i] % shardwidth.ShardWidth)
					}
				}
				continue
			}
			s.trKeys[pos] = adjust(s.trKeys[pos], s.keys[pos])
			err = st.Bulk(s.keys[pos], s.trKeys[pos])
			if err != nil {
				return err
			}
			x := s.trKeys[pos]
			for i := range s.ids {
				x[i] = (x[i] * shardwidth.ShardWidth) + (s.ids[i] % shardwidth.ShardWidth)
			}
		case protoreflect.BytesKind:
			st, err := w.tr.Blobs(name)
			if err != nil {
				return err
			}
			if f.IsList() {
				s.trBlobSet[pos] = adjustSet(s.trBlobSet[pos], s.blobSets[pos])
				err = st.BulkSet(s.blobSets[pos], s.trBlobSet[pos])
				if err != nil {
					return err
				}
				x := s.trBlobSet[pos]

				for i := range s.ids {
					for j := range x[i] {
						x[i][j] = (x[i][j] * shardwidth.ShardWidth) + (s.ids[i] % shardwidth.ShardWidth)
					}
				}
				continue
			}
			s.trBlobs[pos] = adjust(s.trBlobs[pos], s.blobs[pos])
			err = st.Bulk(s.blobs[pos], s.trBlobs[pos])
			if err != nil {
				return err
			}
			if _, ok := s.bsi[name]; ok {
				continue
			}
			x := s.trBlobs[pos]
			for i := range s.ids {
				x[i] = (x[i] * shardwidth.ShardWidth) + (s.ids[i] % shardwidth.ShardWidth)
			}
		}
	}
	// sufficient to cover up to 2097152 records
	shards := make([]uint64, 0, 2)
	viewKey := tx.ViewKey
	err = db.update(func(tx *rbf.Tx) error {
		for start, end := 0,
			shardwidth.FindNextShard(0, s.ids); start < len(s.ids) && end <= len(s.ids); start, end = end,
			shardwidth.FindNextShard(end, s.ids) {
			shard := s.ids[start] / shardwidth.ShardWidth
			shards = append(shards, shard)
			for i := 0; i < fields.Len(); i++ {
				f := fields.Get(i)
				view := viewKey(string(f.Name()), shard)
				pos := s.mapping[string(f.Name())]
				switch f.Kind() {
				case protoreflect.Int64Kind,
					protoreflect.Uint64Kind,
					protoreflect.DoubleKind:
					b := roaring.NewBitmap()
					for n := start; n < end; n++ {
						bsi.Add(b, s.ids[n], s.values[pos][n])
					}
					_, err := tx.AddRoaring(view, b)
					if err != nil {
						return err
					}
				case protoreflect.BoolKind, protoreflect.EnumKind:
					_, err := tx.Add(view, s.rowIDs[pos][start:end]...)
					if err != nil {
						return err
					}
				case protoreflect.StringKind:
					if f.IsList() {
						x := s.trSets[pos][shard:end]
						for n := range x {
							_, err := tx.Add(view, x[n]...)
							if err != nil {
								return err
							}
						}
						continue
					}
					_, err := tx.Add(view, s.trKeys[pos][start:end]...)
					if err != nil {
						return err
					}

				case protoreflect.BytesKind:
					if f.IsList() {
						x := s.trBlobSet[pos][shard:end]
						for n := range x {
							_, err := tx.Add(view, x[n]...)
							if err != nil {
								return err
							}
						}
						continue
					}
					if _, ok := s.bsi[string(f.Name())]; ok {
						b := roaring.NewBitmap()
						for n := start; n < end; n++ {
							bsi.Add(b, s.ids[n], int64(s.trBlobs[pos][n]))
						}
						_, err := tx.AddRoaring(view, b)
						if err != nil {
							return err
						}
						continue
					}
					_, err := tx.Add(view, s.trBlobs[pos][start:end]...)
					if err != nil {
						return err
					}
				}
			}
			// save ids in the _id field
			b := roaring.NewBitmap()
			for n := start; n < end; n++ {
				mutex.Add(b, s.ids[n], s.ids[n])
			}
			_, err := tx.AddRoaring(viewKey(ID, shard), b)
			if err != nil {
				return err
			}

		}
		return nil
	})

	if err != nil {
		return err
	}
	err = w.Commit()
	if err != nil {
		return err
	}
	// update observed shards
	db.updateShards(shards)
	return nil
}

func adjust[T any](tr []uint64, in []T) []uint64 {
	return slices.Grow(tr, len(in))[:len(in)]
}

func adjustSet[T any](tr [][]uint64, in [][]T) [][]uint64 {
	tr = slices.Grow(tr, len(in))[:len(in)]
	for i := range tr {
		tr[i] = adjust(tr[i], in[i])
	}
	return tr
}
