package control

import (
	"fmt"

	"github.com/gernest/vince/pkg/apis/vince/v1alpha1"
	vince_listers "github.com/gernest/vince/pkg/gen/client/vince/listers/vince/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/gernest/vince/pkg/secrets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	stateful_listers "k8s.io/client-go/listers/apps/v1"
	listers "k8s.io/client-go/listers/core/v1"
)

type Topology struct {
	serviceLister     listers.ServiceLister
	vinceLister       vince_listers.VinceLister
	siteLister        vince_listers.SiteLister
	podLister         listers.PodLister
	secretsLister     listers.SecretLister
	configLister      listers.ConfigMapLister
	statefulSetLister stateful_listers.StatefulSetLister
}

func (t *Topology) Build(filter *k8s.ResourceFilter) error {
	r, err := t.loadResources(filter)
	if err != nil {
		return err
	}
	r.Resolve()
	return nil
}

func (t *Topology) loadResources(filter *k8s.ResourceFilter) (*Resources, error) {
	r := &Resources{
		Services: make(map[string]*corev1.Service),
		Secrets:  make(map[string]*corev1.Secret),
		Configs:  make(map[string]*corev1.ConfigMap),
		Pods:     make(map[string]*corev1.Pod),
		Vinces:   make(map[string]*v1alpha1.Vince),
		Sites:    make(map[string]*v1alpha1.Site),
	}

	// First we load all vince resources. These are root resources, we derive any
	// new managed resources from them.
	vince, err := t.vinceLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list vinces maps %v", err)
	}
	for _, o := range vince {
		if filter.IsIgnored(o) {
			continue
		}
		r.Vinces[key(o)] = o
	}
	base := baseSelector()
	svc, err := t.serviceLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list services %v", err)
	}
	secrets, err := t.secretsLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets %v", err)
	}
	config, err := t.configLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list config maps %v", err)
	}
	pods, err := t.podLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods maps %v", err)
	}
	site, err := t.siteLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list vinces maps %v", err)
	}
	stateful, err := t.statefulSetLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list vinces maps %v", err)
	}
	for _, o := range svc {
		if filter.IsIgnored(o) {
			continue
		}
		r.Services[key(o)] = o
	}
	for _, o := range secrets {
		if filter.IsIgnored(o) {
			continue
		}
		r.Secrets[key(o)] = o
	}
	for _, o := range config {
		if filter.IsIgnored(o) {
			continue
		}
		r.Configs[key(o)] = o
	}
	for _, o := range pods {
		if filter.IsIgnored(o) {
			continue
		}
		r.Pods[key(o)] = o
	}

	for _, o := range site {
		if filter.IsIgnored(o) {
			continue
		}
		r.Sites[key(o)] = o
	}
	for _, o := range stateful {
		if filter.IsIgnored(o) {
			continue
		}
		r.StatefulSets[key(o)] = o
	}
	return r, nil
}

type Resources struct {
	Services     map[string]*corev1.Service
	Secrets      map[string]*corev1.Secret
	Configs      map[string]*corev1.ConfigMap
	Pods         map[string]*corev1.Pod
	Vinces       map[string]*v1alpha1.Vince
	Sites        map[string]*v1alpha1.Site
	StatefulSets map[string]*appsv1.StatefulSet
}

func (r *Resources) Resolve() (c ChangeSet) {
	for k, v := range r.Vinces {
		status := v.Status
		switch v.Status.Secret {
		case "":
			continue
		case "Created":
			if _, ok := r.Secrets[k]; !ok {
				// Until we see the secret , no further action is done for this
				// resource
				// TODO: queue resource ?
				continue
			}
			status.Secret = "Resolved"
		}
	}
	return
}

type ChangeSet struct {
	Secrets      []*corev1.Secret
	Configs      []*corev1.ConfigMap
	Services     []*corev1.Service
	VinceStatus  []*v1alpha1.VinceStatus
	StatefulSets []*appsv1.StatefulSet
}

func key(o metav1.Object) string {
	ts := types.NamespacedName{
		Namespace: o.GetNamespace(),
		Name:      o.GetName(),
	}
	return ts.String()
}

func createSecret(o *v1alpha1.Vince) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name,
			Namespace: o.Namespace,
			Labels:    baseLabels(),
		},
		Data: map[string][]byte{
			secrets.API_KEY:    secrets.APIKey(),
			secrets.AGE_KEY:    secrets.AGE(),
			secrets.SECRET_KEY: secrets.ED25519(),
		},
	}
}
