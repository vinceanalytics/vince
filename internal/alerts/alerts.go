package alerts

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/config"
)

type Alerts struct {
	path      string
	files     map[string]*File
	scheduler *Scheduler
}

func (a *Alerts) Close() error {
	return a.scheduler.Close()
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
		path:      path,
		files:     make(map[string]*File),
		scheduler: newScheduler(),
	}
	for k, v := range scripts {
		s, err := Create(string(v))
		if err != nil {
			return nil, fmt.Errorf("script:%q failed to create File instance  %v", k, err)
		}
		if len(s.calls) == 0 {
			continue
		}
		a.files[k] = s
		a.scheduler.add(s.calls)
	}
	return a, nil
}

func Set(ctx context.Context, a *Alerts) context.Context {
	return context.WithValue(ctx, alertsKey{}, a)
}

func Get(ctx context.Context) *Alerts {
	return ctx.Value(alertsKey{}).(*Alerts)
}
