package bpb

import (
	"github.com/dgraph-io/badger/v4/pb"
)

func To(ls *KVList) (o *pb.KVList) {
	o = &pb.KVList{
		Kv:       make([]*pb.KV, 0, len(ls.Kv)),
		AllocRef: ls.AllocRef,
	}
	for _, k := range ls.Kv {
		o.Kv = append(o.Kv, &pb.KV{
			Key:        k.Key,
			Value:      k.Value,
			UserMeta:   k.UserMeta,
			Version:    k.Version,
			ExpiresAt:  k.ExpiresAt,
			Meta:       k.Meta,
			StreamId:   k.StreamId,
			StreamDone: k.StreamDone,
		})
	}
	return
}

func From(ls *pb.KVList) (o *KVList) {
	o = &KVList{
		Kv:       make([]*KV, 0, len(ls.Kv)),
		AllocRef: ls.AllocRef,
	}
	for _, k := range ls.Kv {
		o.Kv = append(o.Kv, &KV{
			Key:        k.Key,
			Value:      k.Value,
			UserMeta:   k.UserMeta,
			Version:    k.Version,
			ExpiresAt:  k.ExpiresAt,
			Meta:       k.Meta,
			StreamId:   k.StreamId,
			StreamDone: k.StreamDone,
		})
	}
	return
}
