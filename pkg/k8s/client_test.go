package k8s

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"

	"github.com/vinceanalytics/vince/pkg/secrets"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSiteAPI(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		r := cloneTransport{}
		a := siteAPI{
			client: http.Client{
				Transport: &r,
			},
		}
		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vince",
				Namespace: "default",
			},
			Data: map[string][]byte{
				secrets.API_KEY: []byte("secret"),
			},
		}
		err := a.Create(context.TODO(), &secret, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		expect, err := os.ReadFile("./testdata/create_api")
		if err != nil {
			t.Fatal(err)
		}
		got, err := httputil.DumpRequestOut(r.r, true)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expect, got) {
			t.Errorf(" ==>\n EXPECTED \n%s\n GOT\n%s", string(expect), string(got))
		}
	})
	t.Run("Delete", func(t *testing.T) {
		r := cloneTransport{}
		a := siteAPI{
			client: http.Client{
				Transport: &r,
			},
		}
		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vince",
				Namespace: "default",
			},
			Data: map[string][]byte{
				secrets.API_KEY: []byte("secret"),
			},
		}
		err := a.Delete(context.TODO(), &secret, "example.com")
		if err != nil {
			t.Fatal(err)
		}
		expect, err := os.ReadFile("./testdata/delete_api")
		if err != nil {
			t.Fatal(err)
		}
		got, err := httputil.DumpRequestOut(r.r, true)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expect, got) {
			t.Errorf(" ==>\n EXPECTED \n%s\n GOT\n%s", string(expect), string(got))
		}
	})

}

var _ http.RoundTripper = (*cloneTransport)(nil)

type cloneTransport struct {
	r *http.Request
}

func (c *cloneTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.r = r.Clone(context.TODO())
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}, nil
}
