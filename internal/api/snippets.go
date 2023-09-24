package api

import (
	"context"

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

// Snippets are named piece of sql query that can be executed. A snippet
// consists of a name, a query string and parameters
//
// Parameter allows named parameter substitution inside the query. Snippets
// names are globally unique, you can't have more than one snippet with the same
// name.

var _ snippetsv1.SnippetsServer = (*API)(nil)

var ErrSnippetNotFound = status.Error(codes.NotFound, "snippet does not exists")

func (a *API) CreateSnippet(ctx context.Context, req *snippetsv1.CreateSnippetRequest) (*snippetsv1.CreateSnippetResponse, error) {
	me := tokens.GetAccount(ctx)
	now := core.Now(ctx)
	key := keys.Snippet(req.Name)
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		if txn.Has(key) {
			return status.Error(codes.AlreadyExists, "snippet already exists")
		}
		return txn.Set(key, px.Encode(&snippetsv1.Snippet{
			Name:      req.Name,
			Query:     req.Query,
			Params:    req.Params,
			CreatedBy: me.Name,
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
		}))
	})
	if err != nil {
		return nil, err
	}
	return &snippetsv1.CreateSnippetResponse{}, nil
}

func (a *API) UpdateSnippet(ctx context.Context, req *snippetsv1.UpdateSnippetRequest) (*snippetsv1.UpdateSnippetResposnes, error) {
	key := keys.Snippet(req.Name)
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		var o snippetsv1.Snippet
		err := txn.Get(key, px.Decode(&o), snippetsE404)
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
	return &snippetsv1.UpdateSnippetResposnes{}, nil
}

func (a *API) ListSnippets(ctx context.Context, req *snippetsv1.ListSnippetsRequest) (*snippetsv1.ListSnippetsResponse, error) {
	var result snippetsv1.ListSnippetsResponse
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Snippet("")
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
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Snippet(req.Name)
		return txn.Delete(key, snippetsE404)
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func snippetsE404() error {
	return ErrSnippetNotFound
}
