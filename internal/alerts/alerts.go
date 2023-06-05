package alerts

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/vinceanalytics/vince/internal/config"
)

type Alerts struct {
	Path    string
	Scripts map[string]*Script
}

func (a *Alerts) Close() error {
	e := make([]error, 0, len(a.Scripts))
	for _, v := range a.Scripts {
		e = append(e, v.Close())
	}
	return errors.Join(e...)
}

type alertsKey struct{}

func Setup(o *config.Options) (*Alerts, error) {
	path := o.Alerts.Source
	if path == "" {
		path := filepath.Join(o.DataPath, "alerts")
		os.MkdirAll(path, 0755)
	}
	scripts, err := Compile(path)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to compile alerts :%v", err)
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
			return nil, fmt.Errorf("script:%q failed to evaluate %v", k, err)
		}
		if len(m.m) == 0 {
			// No registration happened
			continue
		}
		a.Scripts[k] = m
	}
	return a, nil
}

func Set(ctx context.Context, a *Alerts) context.Context {
	return context.WithValue(ctx, alertsKey{}, a)
}
func Get(ctx context.Context) *Alerts {
	return ctx.Value(alertsKey{}).(*Alerts)
}
