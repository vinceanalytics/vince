package cluster

import (
	"fmt"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "cluster",
		Usage: "Manage vince clusters",
		Commands: []*cli.Command{
			setup(),
			add(),
		},
	}
}

func setup() *cli.Command {
	return &cli.Command{
		Name:  "new",
		Usage: "Initialize a new empty cluster",
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			name := ctx.Args().First()
			if name == "" {
				w.Err("missing cluster name")
				w.Suggest(
					"vince cluster create foo",
				).Exit()
			}
			client, path := auth.LoadClient()
			if client.Clusters != nil && client.Clusters[name] != nil {
				return w.Ok("cluster %q already exists", name).Complete(nil)
			}
			if client.Clusters == nil {
				client.Clusters = make(map[string]*v1.Cluster_Config)
			}
			client.Clusters[name] = &v1.Cluster_Config{
				Name: name,
			}
			auth.Save(w, client, path)
			return w.Ok(name).Complete(nil)
		},
	}
}

const addDesc = `Add authenticated vince instance to a cluster. This does not apply the changes.
example:
	vince cluster add foo east007 root
	# vince cluster add [NAME_OF_CLUSTER] [NAME_INSTANCE] [NAME_OF_ACCOUNT]`

func add() *cli.Command {
	return &cli.Command{
		Name:        "add",
		Usage:       "Adds authenticated vince instance to a cluster ",
		Description: addDesc,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "bootstrap,b",
				Usage: "use this node to bootstrap the cluster",
			},
		},
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			a := ctx.Args()
			name := a.First()
			suggest := "vince cluster add" + " " + name
			if name == "" {
				suggest += " foo"
				w.Err("missing cluster name")
				w.Suggest(
					suggest,
				).Exit()
			}
			instance := a.Get(1)
			suggest += " " + instance
			if instance == "" {
				suggest += "http://localhost:8080"
				w.Err("missing  instance")
				w.Suggest(
					suggest,
				).Exit()
			}
			account := a.Get(2)
			if account == "" {
				account = "root"
				w.KV(">", "missing account defaulting to root").Flush()
			}
			client, path := auth.LoadClient()
			if client.Clusters == nil || client.Clusters[name] == nil {
				return w.Complete(fmt.Errorf("cluster %q does not exist", name))
			}
			// adjust instance when we use server id
			if client.ServerId != nil && client.ServerId[instance] != "" {
				instance = client.ServerId[instance]
			}
			if client.Instance == nil || client.Instance[instance] == nil {
				return w.Complete(fmt.Errorf("instance %q does not exist", instance))
			}
			if client.Instance[instance].Accounts == nil || client.Instance[instance].Accounts[account] == nil {
				return w.Complete(fmt.Errorf("account %q does not exist", account))
			}
			ax := client.Instance[instance].Accounts[account]
			if client.Clusters[name].Nodes == nil {
				client.Clusters[name].Nodes = make(map[string]*v1.Cluster_Config_Node)
			}
			client.Clusters[name].Nodes[ax.ServerId] = &v1.Cluster_Config_Node{
				Address:   instance,
				Account:   ax,
				Bootstrap: ctx.Bool("bootstrap"),
			}
			auth.Save(w, client, path)
			return w.Ok(name).Complete(nil)
		},
	}
}
