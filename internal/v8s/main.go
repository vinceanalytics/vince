package v8s

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/v8s/control"
	"github.com/vinceanalytics/vince/internal/v8s/k8s"
	"golang.org/x/sync/errgroup"
)

func CMD() *cli.Command {
	o := &control.Options{}
	return &cli.Command{
		Name:  "k8s",
		Usage: "kubernetes controller for vince - The Cloud Native Web Analytics Platform.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "master-url",
				Usage:       "The address of the Kubernetes API server. Overrides any value in kubeconfig.",
				Destination: &o.MasterURL,
				Sources:     cli.EnvVars("V8S_MASTER_URL"),
			},
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "Path to a kubeconfig. Only required if out-of-cluster.",
				Destination: &o.KubeConfig,
				Sources:     cli.EnvVars("KUBECONFIG"),
			},
			&cli.StringFlag{
				Name:        "default-image",
				Usage:       "Default image of vince to use",
				Value:       "ghcr.io/vinceanalytics/vince:latest",
				Destination: &o.Image,
				Sources:     cli.EnvVars("VINCE_IMAGE"),
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "controller api port",
				Value:       9000,
				Destination: &o.Port,
				Sources:     cli.EnvVars("V8S_API_PORT"),
			},
			&cli.StringFlag{
				Name:        "namespace",
				Usage:       "default namespace where resource managed by v8s will be deployed",
				Destination: &o.Namespace,
				Sources:     cli.EnvVars("V8S_NAMESPACE"),
			},
			&cli.StringSliceFlag{
				Name:        "watch-namespaces",
				Usage:       "namespaces to watch for Vince and Site custom resources",
				Destination: &o.WatchNamespaces,
				Sources:     cli.EnvVars("V8S_WATCH_NAMESPACES"),
			},
			&cli.StringSliceFlag{
				Name:        "ignore-namespaces",
				Usage:       "namespaces to ignore for Vince and Site custom resources",
				Destination: &o.IgnoreNamespaces,
				Sources:     cli.EnvVars("V8S_IGNORE_NAMESPACES"),
			},
		},
		Action: func(ctx *cli.Context) error {
			return run(o)
		},
	}

}

func run(o *control.Options) error {
	slog.SetDefault(config.Logger("debug"))
	slog.Debug("starting controller", slog.Int64("port", o.Port))

	xk8 := k8s.New(o.MasterURL, o.KubeConfig)
	a := &api{}
	xctr := control.New(xk8, *o, a.Ready)
	base, cancel := context.WithCancel(context.Background())
	var g errgroup.Group
	svr := &http.Server{
		Handler: a,
		Addr:    fmt.Sprintf(":%d", o.Port),
	}
	g.Go(func() error {
		defer cancel()
		return svr.ListenAndServe()
	})
	g.Go(func() error {
		defer svr.Close()
		return xctr.Run(base)
	})
	return g.Wait()
}

type api struct {
	ready atomic.Bool
}

func (a *api) Ready() {
	a.ready.Store(true)
}

func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/status/readiness":
		if !a.ready.Load() {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
