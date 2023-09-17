package api

import (
	"context"

	"github.com/oklog/ulid/v2"
	snippetsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/snippets/v1"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
	"github.com/vinceanalytics/vince/internal/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ snippetsv1.SnippetsServer = (*API)(nil)

var snippets404 = status.Error(codes.NotFound, "snippet does not exists")

func SnippetsE404() error {
	return snippets404
}

func (a *API) CreateSnippet(ctx context.Context, req *snippetsv1.CreateSnippetRequest) (*snippetsv1.Snippet, error) {
	me := tokens.GetAccount(ctx)
	now := core.Now(ctx)
	o := snippetsv1.Snippet{
		Id:        ulid.Make().String(),
		Name:      req.Name,
		Query:     req.Query,
		Params:    req.Params,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Snippet(me.Name, o.Id)
		return txn.Set(key, px.Encode(&o))
	})
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (a *API) UpdateSnippet(ctx context.Context, req *snippetsv1.UpdateSnippetRequest) (*snippetsv1.Snippet, error) {
	me := tokens.GetAccount(ctx)
	var o snippetsv1.Snippet
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Snippet(me.Name, req.Id)
		err := txn.Get(key, px.Decode(&o), SnippetsE404)
		if err != nil {
			return err
		}
		proto.Merge(&o, &snippetsv1.Snippet{
			Name:   req.Name,
			Query:  req.Query,
			Params: req.Params,
		})
		return txn.Set(key, px.Encode(&o))
	})
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (a *API) ListSnippets(ctx context.Context, req *snippetsv1.ListSnippetsRequest) (*snippetsv1.ListSnippetsResponse, error) {
	me := tokens.GetAccount(ctx)
	var result snippetsv1.ListSnippetsResponse
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Snippet(me.Name, "")
		it := txn.Iter(db.IterOpts{
			Prefix:         key,
			PrefetchValues: true,
			PrefetchSize:   5,
		})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			var m snippetsv1.Snippet
			err := it.Value(px.Decode(&m))
			if err != nil {
				return err
			}
			result.Snippets = append(result.Snippets, &m)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *API) DeleteSnippet(ctx context.Context, req *snippetsv1.DeleteSnippetRequest) (*emptypb.Empty, error) {
	me := tokens.GetAccount(ctx)
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Snippet(me.Name, req.Id)
		return txn.Delete(key, SnippetsE404)
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
