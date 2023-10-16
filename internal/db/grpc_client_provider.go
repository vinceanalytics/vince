package db

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"time"

	v1 "github.com/vinceanalytics/proto/gen/go/vince/store/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

const defaultBatch = 16

type gProvider struct {
	ctx    context.Context
	client v1.StorageClient
}

var _ Provider = (*gProvider)(nil)

func NewGRPCProvider(ctx context.Context, client v1.StorageClient) Provider {
	return &gProvider{ctx: ctx, client: client}
}

func (g *gProvider) NewTransaction(update bool) Transaction {
	tn, err := g.client.NewTransaction(g.ctx, &v1.NewTransactionRequest{
		Update: update,
	})
	if err != nil {
		slog.Error("failed creating new transaction", "err", err)
		os.Exit(1)
	}
	return &gTxn{
		ctx:    g.ctx,
		txn:    tn,
		client: g.client,
	}

}

type gTxn struct {
	ctx    context.Context
	client v1.StorageClient
	txn    *v1.Transaction
}

var _ Transaction = (*gTxn)(nil)

func (g *gTxn) Set(key, value []byte, ttl time.Duration) error {
	_, err := g.client.Set(g.ctx, &v1.SetRequest{
		Txn:   g.txn,
		Key:   key,
		Value: value,
		Ttl:   durationpb.New(ttl),
	})
	return err
}

func (g *gTxn) Get(key []byte, value func(val []byte) error, notFound ...func() error) error {
	res, err := g.client.Get(g.ctx, &v1.GetRequest{
		Txn: g.txn,
		Key: key,
	})
	if err != nil {
		if len(notFound) == 0 {
			return err
		}
		if isNotFound(err) {
			return notFound[0]()
		}
		return err
	}
	return value(res.Value)
}

func (g *gTxn) Close() error {
	_, err := g.client.Close(g.ctx, g.txn)
	return err
}

func (g *gTxn) Iter(o IterOpts) Iter {
	it, err := g.client.Iter(g.ctx)
	if err != nil {
		slog.Error("failed creating iter stream", "err", err)
		os.Exit(1)
	}

	i := &gIter{
		client: g.client,
		iter:   it,
		txn:    g.txn,
		batch:  defaultBatch,
	}
	err = i.init(o)
	if err != nil {
		slog.Error("failed to initialize stream iterator", "err", err)
		os.Exit(1)
	}
	return i
}

func (g *gTxn) Delete(key []byte, notFound ...func() error) error {
	_, err := g.client.Delete(g.ctx, &v1.DeleteRequest{
		Txn: g.txn,
		Key: key,
	})
	if err != nil {
		if len(notFound) > 0 && isNotFound(err) {
			return notFound[0]()
		}
		return err
	}
	return nil
}

func isNotFound(err error) bool {
	return isCode(err, codes.NotFound)
}

func isCode(err error, code codes.Code) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == code
}

type gIter struct {
	client v1.StorageClient
	iter   v1.Storage_IterClient
	opts   IterOpts
	txn    *v1.Transaction
	items  v1.IterStep
	batch  int
	pos    int
	err    error
}

var _ Iter = (*gIter)(nil)

func (g *gIter) init(o IterOpts) error {
	err := g.iter.SendMsg(&v1.IterRequest{
		Command: &v1.IterRequest_Setup{
			Setup: &v1.IterOption{
				Txn:     g.txn,
				Prefix:  o.Prefix,
				Reverse: o.Reverse,
			},
		},
	})
	if err != nil {
		return err
	}
	err = g.iter.RecvMsg(&g.items)
	if err != nil {
		return err
	}
	g.opts = o
	return g.iter.RecvMsg(&g.items)
}

func (g *gIter) Close() {
	g.iter.CloseSend()
}

func (g *gIter) Valid() bool {
	if kv := g.peek(); kv != nil {
		return bytes.HasPrefix(kv.Key, g.opts.Prefix)
	}
	return false
}

func (g *gIter) Key() []byte {
	return g.peek().Key
}

func (g *gIter) Value(f func([]byte) error) error {
	return f(g.peek().Value)
}

func (g *gIter) Next() {
	g.next()
}

func (g *gIter) peek() *v1.KeyValue {
	if g.err != nil {
		return nil
	}
	if len(g.items.Items) < g.pos {
		return g.items.Items[g.pos]
	}
	return nil
}

func (g *gIter) next() *v1.KeyValue {
	if g.err != nil {
		return nil
	}
	if len(g.items.Items) < g.pos {
		kv := g.items.Items[g.pos]
		g.pos++
		return kv
	}
	err := g.fetch()
	if err != nil {
		g.err = err
		return nil
	}
	g.pos = 0
	return g.next()
}

func (g *gIter) fetch() error {
	err := g.iter.SendMsg(&v1.IterRequest{
		Command: &v1.IterRequest_Batch{
			Batch: uint32(g.batch),
		},
	})
	if err != nil {
		return err
	}
	return g.iter.RecvMsg(&g.items)
}
