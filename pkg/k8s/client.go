package k8s

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gernest/vince/pkg/apis/vince/v1alpha1"
	"github.com/gernest/vince/pkg/gen/client/vince/clientset/versioned"
	"github.com/gernest/vince/pkg/secrets"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const ResyncPeriod = 5 * time.Minute

type Client interface {
	Kube() kubernetes.Interface
	Vince() versioned.Interface
	Site() SiteAPI
}

type SiteAPI interface {
	Create(context.Context, *v1.Secret, *v1alpha1.Site) error
	Delete(context.Context, *v1.Secret, *v1alpha1.Site) error
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

func (w *wrap) Site() SiteAPI {
	return siteAPI{}
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

var httpClient = &http.Client{}

type siteAPI struct{}

func (siteAPI) request(secret *v1.Secret, site *v1alpha1.Site, method, path string, body io.Reader) (*http.Request, error) {
	uri := fmt.Sprintf("http://%s.%s.svc.cluster.local:80/api/v1/sites%s", site.Spec.Target.Name, site.Spec.Target.Namespace, path)
	r, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	bearer := base64.StdEncoding.EncodeToString(secret.Data[secrets.API_KEY])
	r.Header.Set("Authorization", "Bearer "+bearer)
	return r, nil
}

func (siteAPI) Create(ctx context.Context, secret *v1.Secret, site *v1alpha1.Site) error {
	domain, _ := json.Marshal(map[string]any{
		"domain": site.Spec.Domain,
	})
	r, err := siteAPI{}.request(secret, site, http.MethodPost, "", bytes.NewReader(domain))
	if err != nil {
		return err
	}
	return siteAPI{}.do(r)
}

func (siteAPI) Delete(ctx context.Context, secret *v1.Secret, site *v1alpha1.Site) error {
	domain := url.PathEscape(site.Spec.Domain)
	r, err := siteAPI{}.request(secret, site, http.MethodDelete, "/"+domain, nil)
	if err != nil {
		return err
	}
	return siteAPI{}.do(r)
}

func (siteAPI) do(r *http.Request) error {
	res, err := httpClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(http.StatusText(res.StatusCode))
	}
	return nil
}
