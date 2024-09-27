package ro2

import (
	"errors"

	"filippo.io/age"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/model"
	"github.com/vinceanalytics/vince/internal/rbf"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Store struct {
	*DB
}

func Open(path string) (*Store, error) {
	return open(path)
}

func open(path string) (*Store, error) {
	db, err := newDB(path)
	if err != nil {
		return nil, err
	}
	o := &Store{
		DB: db,
	}
	return o, nil
}

var (
	fields  = new(v1.Model).ProtoReflect().Descriptor().Fields()
	tsField = fields.ByNumber(protowire.Number(alicia.TIMESTAMP))
)

func (o *Store) Web() (secret *age.X25519Identity, err error) {
	err = o.Update(func(tx *Tx) error {
		key := tx.get().WebSession()
		it, err := tx.tx.Get(key)
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
			secret, err = age.GenerateX25519Identity()
			if err != nil {
				return err
			}
			return tx.tx.Set(key, []byte(secret.String()))
		}
		return it.Value(func(val []byte) error {
			secret, err = age.ParseX25519Identity(string(val))
			return err
		})
	})
	return
}

func (o *Store) Name(number uint32) string {
	f := fields.ByNumber(protowire.Number(number))
	return string(f.Name())
}

func (o *Store) Number(name string) uint32 {
	f := fields.ByName(protoreflect.Name(name))
	return uint32(f.Number())
}

func (o *Store) ApplyBatch(b model.Batch) error {
	if len(b) == 0 {
		return nil
	}
	for shard, views := range b {
		err := o.shards.Update(shard, func(rtx *rbf.Tx) error {
			for k, v := range views {
				_, err := rtx.AddRoaring(k, v)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Store) Seq() (id uint64) {
	o.View(func(tx *Tx) error {
		id = tx.Seq()
		return nil
	})
	return
}

func (o *Store) Shards(tx *Tx) (a []uint64) {
	q := tx.Seq()
	if q == 0 {
		return
	}
	e := q / shardwidth.ShardWidth

	for i := uint64(0); i <= e; i++ {
		a = append(a, i)
	}
	return
}
