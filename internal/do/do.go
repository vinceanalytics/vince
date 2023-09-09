package do

import (
	"context"
	"strings"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	configv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	queryv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/tokens"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Vince struct {
	v1.VinceClient
	*grpc.ClientConn
}

type Sites struct {
	sitesv1.SitesClient
	*grpc.ClientConn
}
type QueryClient struct {
	queryv1.QueryClient
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

func DoSite(ctx context.Context,
	addr string, token credentials.PerRPCCredentials,
	f func(context.Context, sitesv1.SitesClient) error) error {
	g, err := DialSites(addr, token)
	if err != nil {
		return err
	}
	defer g.Close()
	return f(ctx, g)
}

func DoQuery(ctx context.Context,
	addr string, token credentials.PerRPCCredentials,
	f func(context.Context, queryv1.QueryClient) error) error {
	g, err := DialQuery(addr, token)
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
	in *queryv1.QueryRequest,
) (o *queryv1.QueryResponse, err error) {
	err = DoQuery(ctx, addr, tokens.Source(token), func(ctx context.Context, vc queryv1.QueryClient) error {
		o, err = vc.Query(ctx, in)
		return err
	})
	return
}

func CreateSite(ctx context.Context,
	addr, token string,
	in *sitesv1.CreateSiteRequest,
) (o *sitesv1.CreateSiteResponse, err error) {
	err = DoSite(ctx, addr, tokens.Source(token), func(ctx context.Context, vc sitesv1.SitesClient) error {
		o, err = vc.CreateSite(ctx, in)
		return err
	})
	return
}

func DeleteSite(ctx context.Context,
	addr, token string,
	in *sitesv1.DeleteSiteRequest,
) (o *sitesv1.DeleteSiteResponse, err error) {
	err = DoSite(ctx, addr, tokens.Source(token), func(ctx context.Context, vc sitesv1.SitesClient) error {
		o, err = vc.DeleteSite(ctx, in)
		return err
	})
	return
}

func ListSites(ctx context.Context,
	addr, token string,
	in *sitesv1.ListSitesRequest,
) (o *sitesv1.ListSitesResponse, err error) {
	err = DoSite(ctx, addr, tokens.Source(token), func(ctx context.Context, vc sitesv1.SitesClient) error {
		o, err = vc.ListSites(ctx, in)
		return err
	})
	return
}

func GetSite(ctx context.Context,
	addr, token string,
	in *sitesv1.GetSiteRequest,
) (o *sitesv1.Site, err error) {
	err = DoSite(ctx, addr, tokens.Source(token), func(ctx context.Context, vc sitesv1.SitesClient) error {
		o, err = vc.GetSite(ctx, in)
		return err
	})
	return
}

func Build(ctx context.Context,
	addr, token string,
) (o *configv1.Build, err error) {
	err = Do(ctx, addr, tokens.Source(token), func(ctx context.Context, vc v1.VinceClient) error {
		o, err = vc.Version(ctx, &emptypb.Empty{})
		return err
	})
	return
}

func Dial(addr string, token credentials.PerRPCCredentials) (*Vince, error) {
	conn, err := dial(addr, token)
	if err != nil {
		return nil, err
	}
	return &Vince{ClientConn: conn, VinceClient: v1.NewVinceClient(conn)}, nil
}
func DialSites(addr string, token credentials.PerRPCCredentials) (*Sites, error) {
	conn, err := dial(addr, token)
	if err != nil {
		return nil, err
	}
	return &Sites{ClientConn: conn, SitesClient: sitesv1.NewSitesClient(conn)}, nil
}
func DialQuery(addr string, token credentials.PerRPCCredentials) (*QueryClient, error) {
	conn, err := dial(addr, token)
	if err != nil {
		return nil, err
	}
	return &QueryClient{ClientConn: conn, QueryClient: queryv1.NewQueryClient(conn)}, nil
}

func dial(addr string, token credentials.PerRPCCredentials) (*grpc.ClientConn, error) {
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	return grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(token),
	)
}
