package k8s

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	vince "github.com/gernest/vince/pkg/apis/vince/v1alpha1"
	"github.com/gernest/vince/pkg/gen/client/vince/clientset/versioned"
	fake_vince_client "github.com/gernest/vince/pkg/gen/client/vince/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	fake_kube_client "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

type Mock struct {
	k8s   *fake_kube_client.Clientset
	vince *fake_vince_client.Clientset
}

func NewMock(path string) (*Mock, error) {
	yamlContent, err := os.ReadFile(filepath.FromSlash("./testdata/" + path))
	if err != nil {
		return nil, err
	}
	k0, so := mustParseYaml(yamlContent)
	return &Mock{
		k8s:   fake_kube_client.NewSimpleClientset(k0...),
		vince: fake_vince_client.NewSimpleClientset(so...),
	}, nil
}

func init() {
	err := vince.AddToScheme(scheme.Scheme)
	if err != nil {
		panic("failed to add site to scheme " + err.Error())
	}
}

var _ Client = (*Mock)(nil)

func (m *Mock) Kube() kubernetes.Interface {
	return m.k8s
}

func (m *Mock) Vince() versioned.Interface {
	return m.vince
}

func (m *Mock) Site() SiteAPI {
	return nil
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
		case "Site", "Vince":
			site = append(site, obj)
		case "Deployment", "Endpoints", "Service", "Ingress", "Secret", "Namespace", "Pod", "ConfigMap":
			core = append(core, obj)
		default:
			panic(fmt.Sprintf("The custom-roles configMap contained K8s object types which are not supported! Skipping object with type: %s", groupVersionKind.Kind))
		}
	}
	return
}

var serializer = json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

func Encode(w io.Writer, objects ...runtime.Object) error {
	for _, o := range objects {
		err := serializer.Encode(o, w)
		if err != nil {
			return err
		}
	}
	return nil
}
