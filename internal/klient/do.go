package klient

import (
	"bytes"
	"context"
	"net/http"

	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

var client = &http.Client{}

type Input interface {
	*v1.Token_Create_Request |
		*v1.Site_Get_Request |
		*v1.Site_Create_Request |
		*v1.Site_List_Request |
		*v1.Site_Delete_Request |
		*v1.Query_Request
}

type Output interface {
	*v1.Token_Create_Response |
		*v1.Site_Create_Response |
		*v1.Site_Get_Response |
		*v1.Site_List_Response |
		*v1.Site_Delete_Response |
		*v1.Query_Response
}

func POST[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) *v1.Error {
	return Do(ctx, http.MethodPost, uri, in, out, token...)
}

func DELETE[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) *v1.Error {
	return Do(ctx, http.MethodDelete, uri, in, out, token...)
}

func GET[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) *v1.Error {
	return Do(ctx, http.MethodGet, uri, in, out, token...)
}

func Do[I Input, O Output](ctx context.Context, method, uri string, in I, out O, token ...string) *v1.Error {
	data := must.Must(pj.Marshal(any(in).(proto.Message)))(
		"failed encoding api request object",
	)
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

func CLI_POST[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) {
	cli(POST(ctx, uri, in, out, token...))
}

func CLI_GET[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) {
	cli(GET(ctx, uri, in, out, token...))
}

func CLI_DELETE[I Input, O Output](ctx context.Context, uri string, in I, out O, token ...string) {
	cli(DELETE(ctx, uri, in, out, token...))
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
