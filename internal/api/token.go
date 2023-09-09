package api

import (
	"context"
	"crypto/ed25519"
	"time"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	apiv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	configv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/secrets"
	"github.com/vinceanalytics/vince/internal/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (*API) Login(ctx context.Context, tr *apiv1.LoginRequest) (*apiv1.LoginResponse, error) {
	a := tokens.GetAccount(ctx)
	if tr.Ttl == nil {
		tr.Ttl = durationpb.New(30 * 24 * time.Hour)
	}
	priv := secrets.Get(ctx)
	pub := priv.Public()
	if tr.Generate {
		tr.Token, _ = tokens.Generate(ctx, priv, apiv1.Token_SERVER,
			a.Name, core.Now(ctx).Add(tr.Ttl.AsDuration()))
	} else {
		pub = ed25519.PublicKey(tr.PublicKey)
		tok, err := jwt.Parse(tr.Token, func(t *jwt.Token) (interface{}, error) {
			return pub, nil
		})
		if err != nil || !tok.Valid {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}
	}
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Token(tr.Token)
		defer key.Release()
		tok := must.Must(
			proto.Marshal(&apiv1.Token{
				PubKey: pub.(ed25519.PublicKey),
			}),
		)("failed encoding token")
		return txn.SetTTL(key.Bytes(), tok, tr.Ttl.AsDuration())
	})
	if err != nil {
		slog.Error("failed saving token", "err", err)
		return nil, e500
	}
	o := config.Get(ctx)
	return &apiv1.LoginResponse{
		Auth: &configv1.Client_Auth{
			Name:        a.Name,
			AccessToken: tr.Token,
			ServerId:    o.ServerId,
		},
	}, nil
}
