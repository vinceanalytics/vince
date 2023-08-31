package ha

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/protobuf/proto"
)

const ApplyTTL = time.Second

type provider struct {
	raft *raft.Raft
	base db.Provider
}

var _ db.Provider = (*provider)(nil)

// NewProvider returns a new db.Provider that uses raft to distribute writes.
// Reads are always local to the base provider.
func NewProvider(r *raft.Raft, base db.Provider) db.Provider {
	return &provider{
		raft: r,
		base: base,
	}
}

func (p *provider) With(f func(db *badger.DB) error) error {
	return p.base.With(f)
}

func (p *provider) NewTransaction(update bool) db.Txn {
	return &txn{raft: p.raft, Txn: p.base.NewTransaction(update)}
}

func (p *provider) Txn(update bool, f func(txn db.Txn) error) error {
	x := p.NewTransaction(update)
	err := f(x)
	x.Close()
	return err
}

type txn struct {
	raft *raft.Raft
	db.Txn
}

var _ db.Txn = (*txn)(nil)

func (x *txn) Set(key, value []byte) error {
	return x.SetTTL(key, value, 0)
}

func (x *txn) SetTTL(key, value []byte, ttl time.Duration) error {
	e := must.Must(
		proto.Marshal(px.Raft_EntryFrom(key, value, ttl)),
	)("failed encoding raft entry")
	f := x.raft.Apply(e, ApplyTTL)
	if err := f.Error(); err != nil {
		return err
	}
	return nil
}
