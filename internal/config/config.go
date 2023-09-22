package config

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/secrets"
	"google.golang.org/protobuf/proto"
)

type configKey struct{}

func Get(ctx context.Context) *Options {
	return ctx.Value(configKey{}).(*Options)
}

func Load(base *Options, x *cli.Context) (context.Context, error) {
	// parse env
	base.Env = v1.Config_dev
	if env := x.String("env"); env != "" {
		env = strings.ToLower(env)
		base.Env = v1.Config_Env(v1.Config_Env_value[env])
	}
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
	var secretKey string
	switch e := base.SecretKey.Value.(type) {
	case *v1.Config_SecretKey_Env:
		env := os.Getenv(e.Env)
		must.Assert(env == "")("failed getting secret from env %s", e.Env)
		es := must.Must(base64.StdEncoding.DecodeString(env))(
			"failed decoding key from env %s", e.Env,
		)
		secretKey = string(es)
	case *v1.Config_SecretKey_File:
		e.File = resolve(root, e.File)
		b := must.Must(os.ReadFile(e.File))(
			"failed reading  secret file %s", e.File,
		)
		fs := must.Must(base64.StdEncoding.DecodeString(string(b)))(
			"failed decoding key from file %s", e.File,
		)
		secretKey = string(fs)
	case *v1.Config_SecretKey_Raw:
		rs := must.Must(base64.StdEncoding.DecodeString(string(e.Raw)))(
			"failed decoding raw key",
		)
		secretKey = string(rs)
	}
	baseCtx := secrets.Open(context.Background(), slog.Default(), secretKey)
	base.DbPath = resolve(root, base.DbPath)
	base.RaftPath = resolve(root, base.RaftPath)
	if e, ok := base.BlocksStore.Provider.(*v1.BlockStore_Fs); ok {
		e.Fs.Directory = resolve(root, e.Fs.Directory)
	}
	base.BlocksStore.CacheDir = resolve(root, base.BlocksStore.CacheDir)
	baseCtx = context.WithValue(baseCtx, configKey{}, base)
	return baseCtx, nil
}

func resolve(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.Clean(path))
}

// SocketFile returns path to mysql socket file. Internally we use this socket
// for query endpoint.
func SocketFile(o *Options) string {
	return filepath.Join(o.DbPath, MYSQL_SOCKET)
}
