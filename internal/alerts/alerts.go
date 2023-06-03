package alerts

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/pkg/log"
)

//go:embed vince.ts
var vince []byte

type Alerts struct {
	Path    string
	Scripts map[string]*Script
}

type alertsKey struct{}

func Setup(ctx context.Context, o *config.Options) context.Context {
	path := o.Alerts.Source
	if path == "" {
		path := filepath.Join(o.DataPath, "alerts")
		os.MkdirAll(path, 0755)
	}
	vinceTS := filepath.Join(path, "vince.ts")
	err := os.WriteFile(vinceTS, vince, 0600)
	if err != nil {
		log.Get().Fatal().Err(err).
			Str("path", vinceTS).
			Msg("failed to write vince.ts in the alerts path")
	}
	scripts, err := Compile(path)
	if err != nil {
		if err != nil {
			log.Get().Fatal().Err(err).
				Str("path", vinceTS).
				Msg("failed to compile alerts")
		}
	}
	a := &Alerts{
		Path:    path,
		Scripts: make(map[string]*Script),
	}
	for k, v := range scripts {
		g := goja.New()
		m := &Script{m: make(map[string][]goja.Value)}
		m.runtime = g
		g.Set("VINCE", m)
		_, err = g.RunString(string(v))
		if err != nil {
			log.Get().Fatal().Err(err).
				Str("script", k).
				Msg("failed to evaluate alert script")
		}
		if len(m.m) == 0 {
			// No registration happened
			continue
		}
		a.Scripts[k] = m
	}

	return context.WithValue(ctx, alertsKey{}, a)
}

func Get(ctx context.Context) *Alerts {
	return ctx.Value(alertsKey{}).(*Alerts)
}
