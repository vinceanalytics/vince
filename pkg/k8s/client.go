package k8s

import (
	"os"
	"time"

	"github.com/gernest/vince/pkg/gen/client/vince/clientset/versioned"
	"github.com/rs/zerolog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const ResyncPeriod = 5 * time.Minute

type Client interface {
	Kube() kubernetes.Interface
	Vince() versioned.Interface
}

func New(log *zerolog.Logger, masterURL, kubeConfig string) Client {
	return build(log, masterURL, kubeConfig)
}

func config(log *zerolog.Logger, masterURL, kubeConfig string) *rest.Config {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		// If these env vars are set, we can build an in-cluster config.
		log.Debug().Msg("creating in-cluster client")
		c, err := rest.InClusterConfig()
		if err != nil {
			log.Fatal().Msg("failed to create k8s client")
		}
		return c
	}
	if masterURL != "" || kubeConfig != "" {
		log.Debug().
			Str("master_url", masterURL).
			Str("kube_config", kubeConfig).
			Msg("creating external k8s client ")
		c, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
		if err != nil {
			log.Fatal().
				Str("master_url", masterURL).
				Str("kube_config", kubeConfig).
				Msg("failed creating external k8s client ")
			return nil
		}
		return c
	}
	log.Fatal().Msg("missing masterURL or kubeConfig")
	return nil
}

var _ Client = (*wrap)(nil)

type wrap struct {
	k8s   *kubernetes.Clientset
	vince *versioned.Clientset
}

func (w *wrap) Kube() kubernetes.Interface {
	return w.k8s
}

func (w *wrap) Vince() versioned.Interface {
	return w.vince
}

func build(log *zerolog.Logger, masterURL, kubeConfig string) *wrap {
	r := config(log, masterURL, kubeConfig)
	if r == nil {
		return nil
	}
	k8s, err := kubernetes.NewForConfig(r)
	if err != nil {
		log.Fatal().Msg("failed to build k8s client")
	}
	site, err := versioned.NewForConfig(r)
	if err != nil {
		log.Fatal().Msg("failed to build site client")
	}
	return &wrap{k8s: k8s, vince: site}
}
