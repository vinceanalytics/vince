package klient

import (
	"context"
	"strings"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/tokens"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type G struct {
	v1.VinceClient
	*grpc.ClientConn
}

func Do(ctx context.Context,
	addr string, token credentials.PerRPCCredentials,
	f func(context.Context, v1.VinceClient) error) error {
	g, err := Dial(addr, token)
	if err != nil {
		return err
	}
	defer g.Close()
	return f(ctx, g)
}

func Login(ctx context.Context,
	addr, username, password string,
	in *v1.LoginRequest,
) (o *v1.LoginResponse, err error) {
	err = Do(ctx, addr, tokens.Basic{
		Username: username,
		Password: password,
	}, func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.Login(ctx, in)
		return err
	})
	return
}

func Query(ctx context.Context,
	addr, token string,
	in *v1.QueryRequest,
) (o *v1.QueryResponse, err error) {
	err = Do(ctx, addr, tokens.Source(token), func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.Query(ctx, in)
		return err
	})
	return
}

func CreateSite(ctx context.Context,
	addr, token string,
	in *v1.CreateSiteRequest,
) (o *v1.CreateSiteResponse, err error) {
	err = Do(ctx, addr, tokens.Source(token), func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.CreateSite(ctx, in)
		return err
	})
	return
}

func DeleteSite(ctx context.Context,
	addr, token string,
	in *v1.DeleteSiteRequest,
) (o *v1.DeleteSiteResponse, err error) {
	err = Do(ctx, addr, tokens.Source(token), func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.DeleteSite(ctx, in)
		return err
	})
	return
}

func ListSites(ctx context.Context,
	addr, token string,
	in *v1.ListSitesRequest,
) (o *v1.ListSitesResponse, err error) {
	err = Do(ctx, addr, tokens.Source(token), func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.ListSites(ctx, in)
		return err
	})
	return
}

func GetSite(ctx context.Context,
	addr, token string,
	in *v1.GetSiteRequest,
) (o *v1.GetSiteResponse, err error) {
	err = Do(ctx, addr, tokens.Source(token), func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.GetSite(ctx, in)
		return err
	})
	return
}

func Dial(addr string, token credentials.PerRPCCredentials) (*G, error) {
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(token),
	)
	if err != nil {
		return nil, err
	}
	return &G{ClientConn: conn, VinceClient: v1.NewVinceClient(conn)}, nil
}
