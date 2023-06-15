package alerts

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/js"
)

type Alerts struct {
	path      string
	file      *js.File
	scheduler *js.Scheduler
}

func (a *Alerts) Close() error {
	return a.scheduler.Close()
}

type alertsKey struct{}

func Setup(ctx context.Context, o *config.Options) (*Alerts, error) {
	path := o.Alerts.Source
	if path == "" {
		path := filepath.Join(o.DataPath, "alerts")
		os.MkdirAll(path, 0755)
	}
	scripts, err := js.Compile(ctx, path)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to compile alerts :%v", err)
		}
	}
	a := &Alerts{
		path:      path,
		file:      scripts,
		scheduler: js.NewScheduler(),
	}
	for _, u := range a.file.Units() {
		a.scheduler.Add(u)
	}
	return a, nil
}

func Set(ctx context.Context, a *Alerts) context.Context {
	return context.WithValue(ctx, alertsKey{}, a)
}

func Get(ctx context.Context) *Alerts {
	return ctx.Value(alertsKey{}).(*Alerts)
}

func (a *Alerts) Run(ctx context.Context) {
	a.scheduler.Run(ctx)
}

func (a *Alerts) Work(ctx context.Context) func() error {
	return func() error {
		a.Run(ctx)
		return nil
	}
}
