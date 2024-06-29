package migrate

import (
	"bytes"
	"log"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// Migrate reads raftpb and calls f with logged data.
func Migrate(path string, f func(data *v1.Data) error) error {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dbLogs := []byte("logs")
	curs := tx.Bucket(dbLogs).Cursor()
	var g raft.Log
	for k, value := curs.Seek(nil); k != nil; k, value = curs.Next() {
		err = decodeMsgPack(value, &g)
		if err != nil {
			log.Fatal(err)
		}
		if g.Type == raft.LogCommand {
			var data v1.Data
			err = proto.Unmarshal(g.Data, &data)
			if err != nil {
				return err
			}
			err := f(&data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}
