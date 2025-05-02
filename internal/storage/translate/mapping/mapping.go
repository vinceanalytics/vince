package mapping

import (
	"hash/maphash"
	"path/filepath"
	"reflect"
	"sync"
	"unsafe"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/util/btree"
	"github.com/vinceanalytics/vince/internal/util/seq"
)

const (
	seed     = 1
	sequence = 2
)

type Mapping struct {
	fields [v1.Field_subdivision2_code]*field
	seq    *seq.Seq
}

func New(path string) (*Mapping, error) {
	se, err := seq.New(filepath.Join(path, "SEQ"))
	if err != nil {
		return nil, err
	}
	ma := new(Mapping)
	for f := range v1.Field_subdivision2_code {
		f++
		fx, err := newField(filepath.Join(path, f.String()))
		if err != nil {
			se.Close()
			return nil, err
		}
		f--
		ma.fields[f] = fx
	}
	ma.seq = se

	return ma, nil
}

func (m *Mapping) Next() uint64 {
	return m.seq.Next()
}

func (m *Mapping) Load() uint64 {
	return m.seq.Load()
}

func (m *Mapping) Close() {
	for i := range m.fields {
		m.fields[i].tree.Close()
	}
	m.seq.Close()
}

func (m *Mapping) GetOrCreate(field v1.Field, value []byte) (uint64, bool) {
	if field == 0 || field > v1.Field_subdivision2_code {
		panic("mapping: invalid translation field " + field.String())
	}
	field--
	return m.fields[field].GetOrCreate(value)
}

func (m *Mapping) Get(field v1.Field, value []byte) uint64 {
	if field == 0 || field > v1.Field_subdivision2_code {
		panic("mapping: invalid translation field " + field.String())
	}
	field--
	return m.fields[field].Get(value)
}

func newField(path string) (*field, error) {
	t, err := btree.NewFileTree(path)
	if err != nil {
		return nil, err
	}
	sx := maphash.MakeSeed()
	bs := t.Get(seed)
	if bs == 0 {
		t.Set(seed, extractSeed(sx))
	} else {
		sx = makeSeed(bs)
	}
	return &field{tree: t, seed: sx}, nil

}

type field struct {
	mu   sync.RWMutex
	tree *btree.FileTree
	seed maphash.Seed
}

func (f *field) GetOrCreate(value []byte) (uint64, bool) {
	ha := maphash.Bytes(f.seed, value)
	f.mu.RLock()
	id := f.tree.Get(ha)
	f.mu.RUnlock()
	if id != 0 {
		return id, true
	}
	f.mu.Lock()
	id = f.tree.Incr(sequence)
	f.tree.Set(ha, id)
	f.mu.Unlock()
	return id, false
}

func (f *field) Get(value []byte) uint64 {
	ha := maphash.Bytes(f.seed, value)
	f.mu.RLock()
	id := f.tree.Get(ha)
	f.mu.RUnlock()
	return id
}

func makeSeed(seed uint64) maphash.Seed {
	m := new(maphash.Seed)
	fieldValue := reflect.ValueOf(m).Elem().Field(0)
	if !fieldValue.CanSet() {
		fieldValue = reflect.NewAt(fieldValue.Type(), unsafe.Pointer(fieldValue.UnsafeAddr())).Elem()
	}
	fieldValue.SetUint(seed)
	return *m
}

func extractSeed(m maphash.Seed) uint64 {
	return reflect.ValueOf(m).Field(0).Uint()
}
