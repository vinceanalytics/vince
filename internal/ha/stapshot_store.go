package ha

import (
	"bytes"
	"io"
	"sort"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"google.golang.org/protobuf/proto"
)

var _ raft.SnapshotStore = (*DB)(nil)

func (db *DB) Create(version raft.SnapshotVersion, index, term uint64, configuration raft.Configuration,
	configurationIndex uint64, trans raft.Transport) (raft.SnapshotSink, error) {
	meta := &v1.Raft_Snapshot{
		Version: v1.Raft_Snapshot_Version(version),
		Index:   index,
		Term:    term,
		Id:      ulid.Make().String(),
		Config: &v1.Raft_Config{
			Servers: make([]*v1.Raft_Config_Server, 0, len(configuration.Servers)),
		},
		ConfigIndex: configurationIndex,
	}
	for _, c := range configuration.Servers {
		meta.Config.Servers = append(meta.Config.Servers, &v1.Raft_Config_Server{
			Suffrage: v1.Raft_Config_Server_Suffrage(c.Suffrage),
			Id:       string(c.ID),
			Address:  string(c.Address),
		})
	}
	return &sink{db: db.db, meta: meta}, nil
}

func (db *DB) List() (ls []*raft.SnapshotMeta, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		prefix := keys.RaftSnapshotMeta("")
		defer prefix.Release()
		o := badger.DefaultIteratorOptions
		o.Prefix = prefix.Bytes()
		it := txn.NewIterator(o)
		defer it.Close()
		var m v1.Raft_Snapshot
		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				err = proto.Unmarshal(val, &m)
				if err != nil {
					return err
				}
				ls = append(ls, toMeta(&m))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	sort.Slice(ls, func(i, j int) bool {
		return ls[i].Index > ls[j].Index
	})
	return
}

func (db *DB) Open(id string) (meta *raft.SnapshotMeta, r io.ReadCloser, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		mk := keys.RaftSnapshotMeta(id)
		defer mk.Release()
		xm, err := txn.Get(mk.Bytes())
		if err != nil {
			return err
		}
		var m v1.Raft_Snapshot
		err = xm.Value(func(val []byte) error {
			return proto.Unmarshal(val, &m)
		})
		if err != nil {
			return err
		}
		meta = toMeta(&m)
		dk := keys.RaftSnapshotData(id)
		defer dk.Release()
		xd, err := txn.Get(dk.Bytes())
		if err != nil {
			return err
		}
		v, err := xd.ValueCopy(nil)
		if err != nil {
			return err
		}
		r = io.NopCloser(bytes.NewReader(v))
		return nil
	})
	return
}

func toMeta(m *v1.Raft_Snapshot) (o *raft.SnapshotMeta) {
	o = &raft.SnapshotMeta{
		Version: raft.SnapshotVersion(m.Version),
		Index:   m.Index,
		Term:    m.Term,
		ID:      m.Id,
		Configuration: raft.Configuration{
			Servers: make([]raft.Server, 0, len(m.Config.Servers)),
		},
		ConfigurationIndex: m.ConfigIndex,
	}
	for _, c := range m.Config.Servers {
		o.Configuration.Servers = append(o.Configuration.Servers, raft.Server{
			Suffrage: raft.ServerSuffrage(c.Suffrage),
			ID:       raft.ServerID(c.Id),
			Address:  raft.ServerAddress(c.Address),
		})
	}
	return
}

type sink struct {
	bytes.Buffer
	db   *badger.DB
	meta *v1.Raft_Snapshot
}

var _ raft.SnapshotSink = (*sink)(nil)

func (s *sink) ID() string    { return s.meta.Id }
func (s *sink) Cancel() error { return nil }
func (s *sink) Close() error {
	return s.db.Update(func(txn *badger.Txn) error {
		s.meta.Size = int64(s.Len())
		meta := must.Must(proto.Marshal(s.meta))("failed serializing snapshot metadata")
		a := keys.RaftSnapshotMeta(s.meta.Id)
		defer a.Release()
		err := txn.Set(a.Bytes(), meta)
		if err != nil {
			return err
		}
		b := keys.RaftSnapshotData(s.meta.Id)
		defer b.Release()
		return txn.Set(b.Bytes(), s.Bytes())
	})
}
