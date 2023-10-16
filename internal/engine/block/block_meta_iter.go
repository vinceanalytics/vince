package block

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/blocks/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
)

type MetaIter struct {
	meta   v1.BlockInfo
	domain string
	it     db.Iter
	txn    db.Transaction
}

func NewMetaIter(db db.Provider, domain string) *MetaIter {
	return &MetaIter{
		domain: domain,
		txn:    db.NewTransaction(false),
	}
}

func (m *MetaIter) Next() bool {
	if m.it == nil {
		m.it = m.txn.Iter(db.IterOpts{
			Prefix:         keys.BlockMetadata(m.domain, ""),
			PrefetchValues: true,
			PrefetchSize:   64,
		})
		m.it.Rewind()
	} else {
		m.it.Next()
	}
	return m.it.Valid()
}

func (m *MetaIter) Block() (*v1.BlockInfo, error) {
	return &m.meta, m.it.Value(px.Decode(&m.meta))
}

func (m *MetaIter) Close(_ *sql.Context) error {
	if m.it != nil {
		m.it.Close()
	}
	return m.txn.Close()
}
