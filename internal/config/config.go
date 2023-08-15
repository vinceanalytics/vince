package config

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type configKey struct{}

func Get(ctx context.Context) *Options {
	return ctx.Value(configKey{}).(*Options)
}

func Load(base *Options, x *cli.Context) (context.Context, error) {
	base.SyncInterval = durationpb.New(x.Duration("sync-interval"))
	b := must.Must(os.ReadFile(FILE))(
		"called vince on non vince project, call vince init and try again",
	)
	var f Options
	must.One(pj.Unmarshal(b, &f))("invalid configuration file")
	proto.Merge(base, &f)
	baseCtx := context.WithValue(context.Background(), configKey{}, base)
	return baseCtx, nil
}
