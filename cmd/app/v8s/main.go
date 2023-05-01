package v8s

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gernest/vince/cmd/version"
	"github.com/gernest/vince/pkg/control"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func App() *cli.App {
	return &cli.App{
		Name:  "v8s",
		Usage: "kubernetes controller for vince - The Cloud Native Web Analytics Platform.",
		Commands: []*cli.Command{
			version.Version(),
		},
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "master-url",
				Usage:   "The address of the Kubernetes API server. Overrides any value in kubeconfig.",
				EnvVars: []string{"V8S_MASTER_URL"},
			},
			&cli.StringFlag{
				Name:    "kubeconfig",
				Usage:   "Path to a kubeconfig. Only required if out-of-cluster.",
				EnvVars: []string{"KUBECONFIG"},
			},
			&cli.IntFlag{
				Name:    "port",
				Usage:   "controller api port",
				EnvVars: []string{"V8S_API_PORT"},
				Value:   9000,
			},
		},
		Action: run,
	}

}

func run(ctx *cli.Context) error {
	xlg := zerolog.New(os.Stderr)
	master := ctx.String("master-url")
	kubeconfig := ctx.String("kubeconfig")
	port := ctx.Int("port")
	xlg.Debug().
		Str("master-url", master).
		Str("kubeconfig", kubeconfig).
		Int("port", port).
		Msg("Starting controller")
	xk8 := k8s.New(&xlg, master, kubeconfig)
	a := &api{}
	xctr := control.New(&xlg, xk8, control.Options{}, a.Ready)
	base, cancel := context.WithCancel(context.Background())
	var g errgroup.Group
	svr := &http.Server{
		Handler: a,
		Addr:    fmt.Sprintf(":%d", port),
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
