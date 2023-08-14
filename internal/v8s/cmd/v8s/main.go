package v8s

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/log"
	"github.com/vinceanalytics/vince/internal/v8s/control"
	"github.com/vinceanalytics/vince/internal/v8s/k8s"
	"github.com/vinceanalytics/vince/internal/version"
	"golang.org/x/sync/errgroup"
)

func App() *cli.Command {
	o := &control.Options{}
	return &cli.Command{
		Name:  "v8s",
		Usage: "kubernetes controller for vince - The Cloud Native Web Analytics Platform.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "master-url",
				Usage:       "The address of the Kubernetes API server. Overrides any value in kubeconfig.",
				Destination: &o.MasterURL,
				EnvVars:     []string{"V8S_MASTER_URL"},
			},
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "Path to a kubeconfig. Only required if out-of-cluster.",
				Destination: &o.KubeConfig,
				EnvVars:     []string{"KUBECONFIG"},
			},
			&cli.StringFlag{
				Name:        "default-image",
				Usage:       "Default image of vince to use",
				Value:       fmt.Sprintf("ghcr.io/vinceanalytics/vince:%s", string(version.BuildVersion)),
				Destination: &o.Image,
				EnvVars:     []string{"V8S_VINCE_IMAGE"},
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "controller api port",
				Value:       9000,
				Destination: &o.Port,
				EnvVars:     []string{"V8S_API_PORT"},
			},
			&cli.StringFlag{
				Name:        "namespace",
				Usage:       "default namespace where resource managed by v8s will be deployed",
				Destination: &o.Namespace,
				EnvVars:     []string{"V8S_NAMESPACE"},
			},
			&cli.StringSliceFlag{
				Name:        "watch-namespaces",
				Usage:       "namespaces to watch for Vince and Site custom resources",
				Destination: &o.WatchNamespaces,
				EnvVars:     []string{"V8S_WATCH_NAMESPACES"},
			},
			&cli.StringSliceFlag{
				Name:        "ignore-namespaces",
				Usage:       "namespaces to ignore for Vince and Site custom resources",
				Destination: &o.IgnoreNamespaces,
				EnvVars:     []string{"V8S_IGNORE_NAMESPACES"},
			},
		},
		Action: func(ctx *cli.Context) error {
			return run(o)
		},
	}

}

func run(o *control.Options) error {
	xlg := log.Get()
	xlg.Debug().
		Int("port", o.Port).
		Msg("Starting controller")
	xk8 := k8s.New(xlg, o.MasterURL, o.KubeConfig)
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
