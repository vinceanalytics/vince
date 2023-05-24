package control

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/gernest/vince/pkg/apis/vince/v1alpha1"
	vince_listers "github.com/gernest/vince/pkg/gen/client/vince/listers/vince/v1alpha1"
	"github.com/gernest/vince/pkg/k8s"
	"github.com/gernest/vince/pkg/secrets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	app_listers "k8s.io/client-go/listers/apps/v1"
	listers "k8s.io/client-go/listers/core/v1"
)

type Topology struct {
	clients           k8s.Client
	vinceLister       vince_listers.VinceLister
	siteLister        vince_listers.SiteLister
	statefulSetLister app_listers.StatefulSetLister
	serviceLister     listers.ServiceLister
	secretsLister     listers.SecretLister
	retry             *k8s.Retry
}

func NewTopology(
	clients k8s.Client,
	vinceLister vince_listers.VinceLister,
	siteLister vince_listers.SiteLister,
	statefulSetLister app_listers.StatefulSetLister,
	serviceLister listers.ServiceLister,
	secretsLister listers.SecretLister,
) *Topology {
	return &Topology{
		clients:           clients,
		vinceLister:       vinceLister,
		siteLister:        siteLister,
		statefulSetLister: statefulSetLister,
		serviceLister:     serviceLister,
		secretsLister:     secretsLister,
		retry:             k8s.NewRetry(),
	}
}

func (t *Topology) Build(ctx context.Context, filter *k8s.ResourceFilter, defaultImage string) error {
	r, err := t.loadResources(filter)
	if err != nil {
		return err
	}
	return r.Resolve(ctx, t.retry, defaultImage, t.clients)
}

func (t *Topology) loadResources(filter *k8s.ResourceFilter) (*Resources, error) {
	r := &Resources{
		Secrets:     make(map[string]*corev1.Secret),
		Service:     make(map[string]*corev1.Service),
		Vinces:      make(map[string]*v1alpha1.Vince),
		Sites:       make(map[string][]*v1alpha1.Site),
		StatefulSet: make(map[string]*appsv1.StatefulSet),
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

	secrets, err := t.secretsLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets %v", err)
	}

	svcs, err := t.serviceLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list services %v", err)
	}

	sets, err := t.statefulSetLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list sets %v", err)
	}

	site, err := t.siteLister.List(base)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites %v", err)
	}

	for _, o := range secrets {
		if filter.IsIgnored(o) {
			continue
		}
		r.Secrets[key(o)] = o
	}

	for _, o := range svcs {
		if filter.IsIgnored(o) {
			continue
		}
		r.Service[key(o)] = o
	}

	for _, o := range sets {
		if filter.IsIgnored(o) {
			continue
		}
		r.StatefulSet[key(o)] = o
	}

	for _, o := range site {
		if filter.IsIgnored(o) || o.Spec.Target == nil {
			continue
		}
		k := types.NamespacedName{
			Namespace: o.Spec.Target.Namespace,
			Name:      o.Spec.Target.Name,
		}.String()
		r.Sites[k] = append(r.Sites[k], o)
	}
	return r, nil
}

type Resources struct {
	Secrets     map[string]*corev1.Secret
	Service     map[string]*corev1.Service
	StatefulSet map[string]*appsv1.StatefulSet
	Vinces      map[string]*v1alpha1.Vince
	Sites       map[string][]*v1alpha1.Site
}

func (r *Resources) Resolve(ctx context.Context, retry *k8s.Retry, defaultImage string, clients k8s.Client) error {
	for k, v := range r.Vinces {
		{
			_, ok := r.Secrets[k]
			if !ok {
				secret := createSecret(v)
				err := retry.CreateSecret(ctx, clients, secret)
				if err != nil {
					return err
				}
			}
			// Once created, secret for Vince instance is never updated.
		}
		svc, set := createStatefulSet(v, defaultImage)
		{
			o, ok := r.Service[k]
			if !ok {
				err := retry.CreateService(ctx, clients, svc)
				if err != nil {
					return err
				}
			} else {
				if !reflect.DeepEqual(o.Spec, svc.Spec) {
					err := retry.UpdateService(ctx, clients, svc)
					if err != nil {
						return err
					}
				}
			}
		}
		{
			o, ok := r.StatefulSet[k]
			if !ok {
				err := retry.CreateStatefulSet(ctx, clients, set)
				if err != nil {
					return err
				}
			} else {
				if !reflect.DeepEqual(o.Spec, svc.Spec) {
					err := retry.UpdateStatefulSet(ctx, clients, set)
					if err != nil {
						return err
					}
				}
			}
		}
		// resolve sites
		{
			_, ok := r.StatefulSet[k]
			if !ok {
				// Resources for Vince haven't't been created  yet.
				continue
			}
			sites := r.Sites[k]
			if len(sites) > 0 {
				secret := r.Secrets[k]
				attached := make(map[string]struct{})
				for _, x := range v.Status.Sites {
					attached[x] = struct{}{}
				}
				listed := make(map[string]struct{})
				for _, x := range sites {
					listed[x.Spec.Domain] = struct{}{}
					if _, ok := attached[x.Spec.Domain]; ok {
						// This site has already been created.
						continue
					}
					err := clients.Site().Create(ctx, secret, x.Spec.Domain)
					if err != nil {
						return err
					}
				}
				for x := range attached {
					if _, ok := listed[x]; !ok {
						err := clients.Site().Delete(ctx, secret, x)
						if err != nil {
							return err
						}
					}
				}
				statusChanged := len(attached) != len(listed)
				if !statusChanged && len(listed) > 0 {
					// len(listed) == len(attached) . Make sure all elements are
					// equal as well.
					for x := range listed {
						if _, ok := attached[x]; !ok {
							statusChanged = true
							break
						}
					}
				}
				if statusChanged {
					clone := v.DeepCopy()
					clone.Status.Sites = make([]string, 0, len(listed))
					for x := range listed {
						clone.Status.Sites = append(v.Status.Sites, x)
					}
					sort.Strings(clone.Status.Sites)
					retry.UpdateVinceStatus(ctx, clients, clone)
				}
			}
		}
	}

	for k, v := range r.StatefulSet {
		if !v.DeletionTimestamp.IsZero() {
			continue
		}
		_, ok := r.Vinces[k]
		if !ok {
			retry.DeleteStatefulSet(ctx, clients, v)
		}
	}
	for k, v := range r.Service {
		if !v.DeletionTimestamp.IsZero() {
			continue
		}
		_, ok := r.Vinces[k]
		if !ok {
			retry.DeleteService(ctx, clients, v)
		}
	}
	for k, v := range r.Secrets {
		if !v.DeletionTimestamp.IsZero() {
			continue
		}
		_, ok := r.Vinces[k]
		if !ok {
			retry.DeleteSecret(ctx, clients, v)
		}
	}
	return nil
}

func key(o metav1.Object) string {
	ts := types.NamespacedName{
		Namespace: o.GetNamespace(),
		Name:      o.GetName(),
	}
	return ts.String()
}

func createSecret(o *v1alpha1.Vince) *corev1.Secret {
	var ok bool
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name,
			Namespace: o.Namespace,
			Labels:    baseLabels(),
		},
		Immutable: &ok,
		Data: map[string][]byte{
			secrets.API_KEY:    secrets.APIKey(),
			secrets.AGE_KEY:    secrets.AGE(),
			secrets.SECRET_KEY: secrets.ED25519(),
		},
	}
}

func createStatefulSet(o *v1alpha1.Vince, defaultImage string) (*corev1.Service, *appsv1.StatefulSet) {
	volume := v1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOncePod,
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: o.Spec.Size,
			},
		},
	}
	if o.Spec.StorageClass != "" {
		volume.StorageClassName = &o.Spec.StorageClass
	}
	if o.Spec.Selector != nil {
		volume.Selector = o.Spec.Selector
	}
	container := v1.Container{
		Name:  "vince",
		Image: o.Spec.Image,
	}
	if container.Image == "" {
		container.Image = defaultImage
	}
	container.Name = "vince"
	if container.Image == "" {
		container.Image = defaultImage
	}
	container.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      o.Name,
			MountPath: "/data",
		},
	}
	if o.Spec.SubPath != "" {
		container.VolumeMounts[0].SubPath = o.Spec.SubPath
	}
	container.Env = append(o.Spec.Env,
		corev1.EnvVar{
			Name:  "VINCE_DATA",
			Value: "/data",
		},
		corev1.EnvVar{
			Name: secrets.API_KEY_ENV,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.Name,
					},
					Key: secrets.API_KEY,
				},
			},
		},
		corev1.EnvVar{
			Name: secrets.AGE_KEY_ENV,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.Name,
					},
					Key: secrets.AGE_KEY,
				},
			},
		},
		corev1.EnvVar{
			Name: secrets.SECRET_KEY_ENV,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.Name,
					},
					Key: secrets.SECRET_KEY,
				},
			},
		},
		corev1.EnvVar{
			Name:  "VINCE_LISTEN",
			Value: ":80",
		},
		corev1.EnvVar{
			Name:  "VINCE_TLS_LISTEN",
			Value: ":443",
		},
	)
	container.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: 80,
		},
		{
			Name:          "https",
			ContainerPort: 443,
		},
	}

	return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      o.Name,
				Namespace: o.Namespace,
				Labels:    baseLabels(),
			},
			Spec: corev1.ServiceSpec{
				Selector: baseLabels(),
				Ports: []corev1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
					{
						Name: "https",
						Port: 443,
					},
				},
			},
		}, &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      o.Name,
				Namespace: o.Namespace,
				Labels:    baseLabels(),
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: baseLabels(),
				},
				ServiceName: o.Name,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: baseLabels(),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							container,
						},
					},
				},
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: o.Name,
						},
						Spec: volume,
					},
				},
			},
		}
}
