package k8s

import (
	"context"

	"github.com/cenkalti/backoff/v4"
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
