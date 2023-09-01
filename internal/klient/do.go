package klient

import (
	"bytes"
	"context"
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"google.golang.org/protobuf/proto"
)

var client = &http.Client{}

type Input interface {
	*v1.CreateTokenRequest |
		*v1.GetSiteRequest |
		*v1.CreateSiteRequest |
		*v1.ListSitesRequest |
		*v1.DeleteSiteRequest |
		*v1.QueryRequest |
		*v1.ApplyClusterRequest |
		*v1.GetClusterRequest
}

type Output interface {
	*v1.CreateTokenResponse |
		*v1.CreateSiteResponse |
		*v1.GetSiteResponse |
		*v1.ListSitesResponse |
		*v1.DeleteSiteResponse |
		*v1.QueryResponse |
		*v1.ApplyClusterResponse |
		*v1.GetClusterResponse
}

func Do[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) *v1.Error {
	data := must.Must(pj.Marshal(any(in).(proto.Message)))(
		"failed encoding api request object",
	)
	var method, path string
	switch any(in).(type) {
	case *v1.CreateTokenRequest:
		path = "/tokens"
		method = http.MethodPost
	case *v1.GetSiteRequest:
		path = "/sites"
		method = http.MethodGet
	case *v1.CreateSiteRequest:
		path = "/sites"
		method = http.MethodPost

	case *v1.ListSitesRequest:
		path = "/sites"
		method = http.MethodGet

	case *v1.DeleteSiteRequest:
		path = "/sites"
		method = http.MethodDelete

	case *v1.QueryRequest:
		path = "/query"
		method = http.MethodPost

	case *v1.ApplyClusterRequest:
		path = "/apply"
		method = http.MethodPost

	case *v1.GetClusterRequest:
		path = "/cluster"
		method = http.MethodGet
	}
	uri += path
	r := must.Must(http.NewRequestWithContext(ctx, method, uri, bytes.NewReader(data)))(
		"failed creating api request",
	)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("content-type", "application/json")
	if len(token) > 0 {
		r.Header.Set("authorization", "Bearer "+token[0])
	}
	res := must.Must(client.Do(r))(
		"failed sending api request", "uri", uri,
	)
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var out v1.Error
		must.One(pj.UnmarshalDefault(&out, res.Body))(
			"failed decoding api error",
		)
		return &out
	}
	must.One(pj.UnmarshalDefault(any(out).(proto.Message), res.Body))(
		"failed decoding api result",
	)
	return nil
}

func CLI[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) {
	cli(Do(ctx, uri, in, out, token...))
}

func cli(err *v1.Error) {
	if err != nil {
		w := ansi.New()
		w.Err(err.Error)
		if err.Code == http.StatusUnauthorized {
			w.Suggest(
				"vince login",
			)
		}
		w.Exit()
	}
}
