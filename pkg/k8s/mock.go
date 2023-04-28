package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	site "github.com/gernest/vince/pkg/apis/site/v1alpha1"
	"github.com/gernest/vince/pkg/gen/client/site/clientset/versioned"
	fake_site_client "github.com/gernest/vince/pkg/gen/client/site/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	fake_kube_client "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

type Mock struct {
	k8s  *fake_kube_client.Clientset
	site *fake_site_client.Clientset
}

func NewMock(path string) *Mock {
	yamlContent, err := os.ReadFile(filepath.FromSlash("./testdata/" + path))
	if err != nil {
		panic(err)
	}
	k0, so := mustParseYaml(yamlContent)
	return &Mock{
		k8s:  fake_kube_client.NewSimpleClientset(k0...),
		site: fake_site_client.NewSimpleClientset(so...),
	}
}

func init() {
	err := site.AddToScheme(scheme.Scheme)
	if err != nil {
		panic("failed to add site to scheme " + err.Error())
	}
}

var _ Client = (*Mock)(nil)

func (m *Mock) Kube() kubernetes.Interface {
	return m.k8s
}

func (m *Mock) Site() versioned.Interface {
	return m.site
}

func mustParseYaml(content []byte) (core, site []runtime.Object) {
	files := strings.Split(string(content), "---")
	for _, file := range files {
		if file == "\n" || file == "" {
			continue
		}
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(file), nil, nil)
		if err != nil {
			panic(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
		}
		switch groupVersionKind.Kind {
		case "Site":
			site = append(site, obj)
		case "Deployment", "Endpoints", "Service", "Ingress", "Secret", "Namespace", "Pod", "ConfigMap":
			core = append(core, obj)
		default:
			panic(fmt.Sprintf("The custom-roles configMap contained K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind))
		}
	}
	return
}
