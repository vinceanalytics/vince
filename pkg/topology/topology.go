package topology

import (
	"fmt"

	"github.com/gernest/vince/pkg/apis/vince/v1alpha1"
	vince_listers "github.com/gernest/vince/pkg/gen/client/vince/listers/vince/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	apppsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	listers "k8s.io/client-go/listers/core/v1"
)

type Topology struct {
	serviceLister listers.ServiceLister
	vinceLister   vince_listers.VinceLister
	siteLister    vince_listers.SiteLister
	podLister     listers.PodLister
	secretsLister listers.SecretLister
	configLister  listers.ConfigMapLister
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
		Services: make(map[types.NamespacedName]*corev1.Service),
		Secrets:  make(map[types.NamespacedName]*corev1.Secret),
		Configs:  make(map[types.NamespacedName]*corev1.ConfigMap),
		Pods:     make(map[types.NamespacedName]*corev1.Pod),
		Vinces:   make(map[types.NamespacedName]*v1alpha1.Vince),
		Sites:    make(map[types.NamespacedName]*v1alpha1.Site),
	}
	svc, err := t.serviceLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list services %v", err)
	}
	secrets, err := t.secretsLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets %v", err)
	}
	config, err := t.configLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list config maps %v", err)
	}
	pods, err := t.podLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list pods maps %v", err)
	}
	vince, err := t.vinceLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list vinces maps %v", err)
	}
	site, err := t.siteLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list vinces maps %v", err)
	}
	for _, o := range svc {
		if filter.IsIgnored(o) {
			continue
		}
		r.Services[Key{Namespace: o.Namespace, Name: o.Name}] = o
	}
	for _, o := range secrets {
		if filter.IsIgnored(o) {
			continue
		}
		r.Secrets[Key{Namespace: o.Namespace, Name: o.Name}] = o
	}
	for _, o := range config {
		if filter.IsIgnored(o) {
			continue
		}
		r.Configs[Key{Namespace: o.Namespace, Name: o.Name}] = o
	}
	for _, o := range pods {
		if filter.IsIgnored(o) {
			continue
		}
		r.Pods[Key{Namespace: o.Namespace, Name: o.Name}] = o
	}
	for _, o := range vince {
		if filter.IsIgnored(o) {
			continue
		}
		r.Vinces[Key{Namespace: o.Namespace, Name: o.Name}] = o
	}
	for _, o := range site {
		if filter.IsIgnored(o) {
			continue
		}
		r.Sites[Key{Namespace: o.Namespace, Name: o.Name}] = o
	}
	return r, nil
}

type Key = types.NamespacedName

type Resources struct {
	Services map[Key]*corev1.Service
	Secrets  map[Key]*corev1.Secret
	Configs  map[Key]*corev1.ConfigMap
	Pods     map[Key]*corev1.Pod
	Vinces   map[Key]*v1alpha1.Vince
	Sites    map[Key]*v1alpha1.Site
}

func (r *Resources) Resolve() *ChangeSet {
	return nil
}

type ChangeSet struct {
	Secrets     []*corev1.Secret
	Configs     []*corev1.ConfigMap
	Services    []*corev1.Service
	VinceStatus []v1alpha1.VinceStatus
	Deployments []*apppsv1.Deployment
}
