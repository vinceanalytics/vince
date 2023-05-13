package alerts

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pkg/log"
)

//go:embed vince.ts
var vince []byte

type Alerts struct {
	Path    string
	Scripts map[string][]byte
}

type alertsKey struct{}

func Setup(ctx context.Context) context.Context {
	o := config.Get(ctx)
	path := filepath.Join(o.DataPath, "alerts")
	os.MkdirAll(path, 0755)
	vinceTS := filepath.Join(path, "vince.ts")
	err := os.WriteFile(vinceTS, vince, 0600)
	if err != nil {
		log.Get(ctx).Fatal().Err(err).
			Str("path", vinceTS).
			Msg("failed to write vince.ts in the alerts path")
	}
	scripts, err := Compile(path)
	if err != nil {
		if err != nil {
			log.Get(ctx).Fatal().Err(err).
				Str("path", vinceTS).
				Msg("failed to compile alerts")
		}
	}
	a := &Alerts{
		Path:    path,
		Scripts: scripts,
	}
	return context.WithValue(ctx, alertsKey{}, a)
}

func Get(ctx context.Context) *Alerts {
	return ctx.Value(alertsKey{}).(*Alerts)
}
