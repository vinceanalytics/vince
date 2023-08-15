package config

import (
	"context"
	"os"
	"path/filepath"

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
	root := x.Args().First()
	base.SyncInterval = durationpb.New(x.Duration("sync-interval"))
	file := FILE
	if root != "" {
		file = filepath.Join(root, file)
	}
	b := must.Must(os.ReadFile(file))(
		"called vince on non vince project, call vince init and try again",
	)
	var f Options
	must.One(pj.Unmarshal(b, &f))("invalid configuration file")
	proto.Merge(base, &f)

	// resolve paths to their absolute values when file path is passed as first
	// argument to serve.
	if root := x.Args().First(); root != "" {
		if !filepath.IsAbs(root) {
			root = must.Must(filepath.Abs(root))(
				"failed resolving absolute path",
			)
		}
		if !filepath.IsAbs(base.MetaPath) {
			base.MetaPath = filepath.Join(root, base.MetaPath)
		}
		if !filepath.IsAbs(base.BlocksPath) {
			base.BlocksPath = filepath.Join(root, base.BlocksPath)
		}
	} else {
		if !filepath.IsAbs(base.MetaPath) {
			base.MetaPath = must.Must(filepath.Abs(base.MetaPath))(
				"failed to to resolve absolute path for meta",
			)
		}
		if !filepath.IsAbs(base.BlocksPath) {
			base.BlocksPath = must.Must(filepath.Abs(base.BlocksPath))(
				"failed to to resolve absolute path for blocks",
			)
		}
	}
	baseCtx := context.WithValue(context.Background(), configKey{}, base)
	return baseCtx, nil
}
