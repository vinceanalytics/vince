package k8s

import (
	"context"
	"os"
	"time"

	"github.com/gernest/vince/pkg/gen/client/site/clientset/versioned"
	"github.com/gernest/vince/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const ResyncPeriod = 5 * time.Minute

type Client interface {
	Kube() kubernetes.Interface
	Site() versioned.Interface
}

func New(ctx context.Context, masterURL, kubeConfig string) Client {
	return build(ctx, masterURL, kubeConfig)
}

func config(ctx context.Context, masterURL, kubeConfig string) *rest.Config {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		// If these env vars are set, we can build an in-cluster config.
		log.Get(ctx).Debug().Msg("creating in-cluster client")
		c, err := rest.InClusterConfig()
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to create k8s client")
			return nil
		}
		return c
	}
	if masterURL != "" || kubeConfig != "" {
		log.Get(ctx).Debug().
			Str("master_url", masterURL).
			Str("kube_config", kubeConfig).
			Msg("creating external k8s client ")
		c, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
		if err != nil {
			log.Get(ctx).Err(err).
				Str("master_url", masterURL).
				Str("kube_config", kubeConfig).
				Msg("failed creating external k8s client ")
			return nil
		}
		return c
	}
	log.Get(ctx).Error().Msg("missing masterURL or kubeConfig")
	return nil
}

var _ Client = (*wrap)(nil)

type wrap struct {
	k8s  *kubernetes.Clientset
	site *versioned.Clientset
}

func (w *wrap) Kube() kubernetes.Interface {
	return w.k8s
}

func (w *wrap) Site() versioned.Interface {
	return w.site
}

func build(ctx context.Context, masterURL, kubeConfig string) *wrap {
	r := config(ctx, masterURL, kubeConfig)
	if r == nil {
		return nil
	}
	k8s, err := kubernetes.NewForConfig(r)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to build k8s client")
		return nil
	}
	site, err := versioned.NewForConfig(r)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to build site client")
		return nil
	}
	return &wrap{k8s: k8s, site: site}
}
