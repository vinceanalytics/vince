package config

import (
	"context"
	"os"
	"path/filepath"

	"github.com/bufbuild/protovalidate-go"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"google.golang.org/protobuf/proto"
)

type configKey struct{}

func Get(ctx context.Context) *Options {
	return ctx.Value(configKey{}).(*Options)
}

func Load(base *Options, x *cli.Context) (context.Context, error) {
	root, err := filepath.Abs(x.Args().First())
	if err != nil {
		return nil, err
	}
	file := filepath.Join(root, FILE)
	b := must.Must(os.ReadFile(file))(
		"called vince on non vince project, call vince init and try again",
	)
	var f Options
	must.One(pj.Unmarshal(b, &f))("invalid configuration file")
	proto.Merge(base, &f)

	valid := must.Must(protovalidate.New())("failed initializing protovalidate")
	if err := valid.Validate(base); err != nil {
		println(err.Error())
		os.Exit(1)
	}
	base.DbPath = resolve(root, base.DbPath)
	base.RaftPath = resolve(root, base.RaftPath)
	if e, ok := base.BlocksStore.Provider.(*v1.BlockStore_Fs); ok {
		e.Fs.Directory = resolve(root, e.Fs.Directory)
	}
	base.BlocksStore.CacheDir = resolve(root, base.BlocksStore.CacheDir)
	baseCtx := context.WithValue(context.Background(), configKey{}, base)
	return baseCtx, nil
}

func resolve(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.Clean(path))
}
