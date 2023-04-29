package main

import (
	"os"

	"github.com/gernest/vince/pkg/version"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

func main() {
	a := cli.NewApp()
	a.Name = "v8s"
	a.Usage = "The open source single file, self hosted web analytics platform."
	a.Commands = []*cli.Command{
		version.Version(),
	}
	a.Flags = []cli.Flag{
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
	}
	a.Action = make
	err := a.Run(os.Args)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func make(ctx *cli.Context) error {
	xlg := zerolog.New(os.Stderr)
	master := ctx.String("master-url")
	kubeconfig := ctx.String("kubeconfig")
	port := ctx.Int("port")
	xlg.Debug().
		Str("master-url", master).
		Str("kubeconfig", kubeconfig).
		Int("port", port).
		Msg("Starting controller")
	return nil
}
