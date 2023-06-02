package k8s

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/vinceanalytics/vince/pkg/apis/vince/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Retry struct {
	opts backoff.BackOff
}

func NewRetry() *Retry {
	return &Retry{
		opts: backoff.NewExponentialBackOff(),
	}
}

func (r *Retry) CreateSecret(ctx context.Context, clients Client, secret *v1.Secret) error {
	x := clients.Kube().CoreV1().Secrets(secret.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		_, err := x.Create(ctx, secret.DeepCopy(), metav1.CreateOptions{})
		return err
	}, r.opts)
}

func (r *Retry) DeleteSecret(ctx context.Context, clients Client, secret *v1.Secret) error {
	x := clients.Kube().CoreV1().Secrets(secret.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		return x.Delete(ctx, secret.Name, metav1.DeleteOptions{})
	}, r.opts)
}

func (r *Retry) CreateService(ctx context.Context, clients Client, service *v1.Service) error {
	x := clients.Kube().CoreV1().Services(service.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		_, err := x.Create(ctx, service.DeepCopy(), metav1.CreateOptions{})
		return err
	}, r.opts)
}

func (r *Retry) UpdateService(ctx context.Context, clients Client, service *v1.Service) error {
	x := clients.Kube().CoreV1().Services(service.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		_, err := x.Update(ctx, service.DeepCopy(), metav1.UpdateOptions{})
		return err
	}, r.opts)
}

func (r *Retry) DeleteService(ctx context.Context, clients Client, service *v1.Service) error {
	x := clients.Kube().CoreV1().Services(service.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		return x.Delete(ctx, service.Name, metav1.DeleteOptions{})
	}, r.opts)
}

func (r *Retry) CreateStatefulSet(ctx context.Context, clients Client, set *appsv1.StatefulSet) error {
	x := clients.Kube().AppsV1().StatefulSets(set.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		_, err := x.Create(ctx, set.DeepCopy(), metav1.CreateOptions{})
		return err
	}, r.opts)
}

func (r *Retry) UpdateStatefulSet(ctx context.Context, clients Client, set *appsv1.StatefulSet) error {
	x := clients.Kube().AppsV1().StatefulSets(set.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		_, err := x.Update(ctx, set.DeepCopy(), metav1.UpdateOptions{})
		return err
	}, r.opts)
}

func (r *Retry) DeleteStatefulSet(ctx context.Context, clients Client, set *appsv1.StatefulSet) error {
	x := clients.Kube().AppsV1().StatefulSets(set.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		return x.Delete(ctx, set.Name, metav1.DeleteOptions{})
	}, r.opts)
}

func (r *Retry) UpdateVinceStatus(ctx context.Context, clients Client, set *v1alpha1.Vince) error {
	x := clients.Vince().StaplesV1alpha1().Vinces(set.Namespace)
	r.opts.Reset()
	return backoff.Retry(func() error {
		_, err := x.UpdateStatus(ctx, set.DeepCopy(), metav1.UpdateOptions{})
		return err
	}, r.opts)
}
