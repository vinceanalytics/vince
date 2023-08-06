package neo

import (
	"bytes"
	"context"
	"errors"
	"sync"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/memory"
	"github.com/apache/arrow/go/v13/parquet"
	"github.com/apache/arrow/go/v13/parquet/compress"
	"github.com/apache/arrow/go/v13/parquet/file"
	"github.com/apache/arrow/go/v13/parquet/pqarrow"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

const (
	blockPrefix    = "block/"
	metadataPrefix = "meta/"
	indexPrefix    = "index/"
)

type ActiveBlock struct {
	mu    sync.Mutex
	db    *badger.DB
	hosts map[string]*entry.MultiEntry
}

func NewBlock(dir string, db *badger.DB) *ActiveBlock {
	return &ActiveBlock{
		db:    db,
		hosts: make(map[string]*entry.MultiEntry),
	}
}

func (a *ActiveBlock) Save(ctx context.Context) {
	ts := core.Now(ctx).UnixMilli()
	a.mu.Lock()
	if len(a.hosts) == 0 {
		a.mu.Unlock()
		return
	}
	for k, v := range a.hosts {
		go a.save(ctx, k, ts, v)
		delete(a.hosts, k)
	}
	a.mu.Unlock()
}

func (a *ActiveBlock) Close() error {
	a.Save(context.Background())
	return nil
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	a.mu.Lock()
	h, ok := a.hosts[e.Domain]
	if !ok {
		h = entry.NewMulti()
		a.hosts[e.Domain] = h
	}
	a.mu.Unlock()
	h.Append(e)
	e.Release()
}

func (a *ActiveBlock) save(ctx context.Context, domain string, ts int64, m *entry.MultiEntry) {

	// Wrap default allocator to pass when doing computation
	ctx = entry.Context(ctx)

	txn := a.db.NewTransaction(true)
	meta := must.
		Must(ReadMetadata(txn, domain))("failed to read metadata for domain", domain)
	r := m.Record(ts)

	var block *v1.Block
	if len(meta.Blocks) > 0 {
		last := meta.Blocks[len(meta.Blocks)-1]
		if last.Size < (1 << 20) {
			block = last
		}
	}
	if block == nil {
		id := ulid.Make()
		block = &v1.Block{
			Id:  id.Bytes(),
			Min: ts,
		}
		meta.Blocks = append(meta.Blocks, block)
	}
	block.Max = ts
	buf := get()
	defer func() {
		put(buf)
		r.Release()
		m.Release()
	}()
	must.One(WriteBlock(ctx, txn, buf,
		[]byte(blockPrefix+domain+string(block.Id)), r))("failed to write block")
	block.Size = int64(buf.Len())
	index := Index(ctx, buf.Bytes())
	must.One(errors.Join(
		txn.Set([]byte(metadataPrefix+domain), must.Must(proto.Marshal(meta))()),
		txn.Set([]byte(indexPrefix+domain), must.Must(proto.Marshal(&index))()),
		txn.Commit(),
	))("failed to commit write block")
}

func ReadMetadata(txn *badger.Txn, domain string) (*v1.Metadata, error) {
	it, err := txn.Get([]byte(domain))
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return nil, err
		}
		return &v1.Metadata{}, nil
	}
	meta := &v1.Metadata{}
	err = it.Value(func(val []byte) error {
		return proto.Unmarshal(val, meta)
	})
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// WriteBlock saves record r in a parquet file with key. If the block exists a
// new file is created that adds record r to it.
func WriteBlock(ctx context.Context, txn *badger.Txn, b *bytes.Buffer, key []byte, r arrow.Record) (err error) {
	it, err := txn.Get(key)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		w, err := pqarrow.NewFileWriter(entry.Schema, b,
			parquet.NewWriterProperties(
				parquet.WithAllocator(entry.Pool),
				parquet.WithCompression(compress.Codecs.Zstd),
				parquet.WithCompressionLevel(10),
			),
			pqarrow.NewArrowWriterProperties(
				pqarrow.WithAllocator(entry.Pool),
			))

		if err != nil {
			return err
		}
		err = w.Write(r)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return txn.Set(key, b.Bytes())
	}
	err = it.Value(func(val []byte) error {
		pr, err := pqarrow.ReadTable(ctx, bytes.NewReader(val), &parquet.ReaderProperties{}, pqarrow.ArrowReadProperties{
			Parallel: true,
		}, entry.Pool)
		if err != nil {
			return err
		}
		defer pr.Release()
		w, err := pqarrow.NewFileWriter(entry.Schema, b,
			parquet.NewWriterProperties(
				parquet.WithAllocator(entry.Pool),
				parquet.WithCompression(compress.Codecs.Zstd),
				parquet.WithCompressionLevel(10),
			),
			pqarrow.NewArrowWriterProperties(
				pqarrow.WithAllocator(entry.Pool),
			))
		if err != nil {
			return err
		}
		err = w.WriteTable(pr, 1<<20)
		if err != nil {
			return err
		}
		err = w.Write(r)
		if err != nil {
			return err
		}
		return w.Close()
	})
	if err != nil {
		return err
	}
	return txn.Set(key, b.Bytes())
}

func get() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func put(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

var bufferPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// ReadBlock reads records constructed by combining fields from block with key.
// For each record cb is called with it. If cb returns false reading is halter.
func ReadBlock(ctx context.Context, db *badger.DB, key []byte, a Analysis) error {
	return db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			r, err := ReadRecord(ctx, bytes.NewReader(val), a.ColumnIndices(), nil)
			if err != nil {
				return err
			}
			a.Analyze(ctx, r)
			return nil
		})
	})
}

func ReadRecord(ctx context.Context, rd parquet.ReaderAtSeeker, cols, groups []int) (arrow.Record, error) {
	f, err := file.NewParquetReader(rd, file.WithReadProps(parquet.NewReaderProperties(
		entry.Pool,
	)))
	if err != nil {
		return nil, err
	}
	r, err := pqarrow.NewFileReader(f, pqarrow.ArrowReadProperties{
		Parallel: true,
	}, entry.Pool)
	if err != nil {
		return nil, err
	}
	if groups == nil {
		groups = make([]int, f.NumRowGroups())
		for idx := range groups {
			groups[idx] = idx
		}
	}
	if cols == nil {
		cols = make([]int, f.MetaData().Schema.NumColumns())
		for idx := range cols {
			cols[idx] = idx
		}
	}
	readers, schema, err := r.GetFieldReaders(ctx, cols, groups)
	if err != nil {
		return nil, err
	}
	count := int64(0)
	for _, rg := range groups {
		count += f.MetaData().RowGroup(rg).NumRows()
	}
	result := make([]arrow.Array, len(readers))
	errs := make([]error, len(readers))
	var wg sync.WaitGroup
	for i := 0; i < len(readers); i++ {
		wg.Add(1)
		go read(ctx, &wg, i, count, result, errs, readers[i])
	}
	wg.Wait()
	err = errors.Join(errs...)
	if err != nil {
		for i := range result {
			if result[i] != nil {
				result[i].Release()
				result[i] = nil
			}
		}
		return nil, err
	}
	return array.NewRecord(schema, result, count), nil
}

func read(ctx context.Context, wg *sync.WaitGroup, idx int, count int64, o []arrow.Array, errs []error, r *pqarrow.ColumnReader) {
	defer wg.Done()
	chunk, err := r.NextBatch(int64(count))
	if err != nil {
		errs[idx] = err
		return
	}
	data, err := chunksToSingle(chunk)
	if err != nil {
		errs[idx] = err
		chunk.Release()
		return
	}
	o[idx] = array.MakeFromData(data)
	data.Release()
	chunk.Release()
}

func chunksToSingle(chunked *arrow.Chunked) (arrow.ArrayData, error) {
	switch len(chunked.Chunks()) {
	case 0:
		return array.NewData(chunked.DataType(), 0, []*memory.Buffer{nil, nil}, nil, 0, 0), nil
	case 1:
		data := chunked.Chunk(0).Data()
		data.Retain() // we pass control to the caller
		return data, nil
	default: // if an item reader yields a chunked array, this is not yet implemented
		return nil, arrow.ErrNotImplemented
	}
}
