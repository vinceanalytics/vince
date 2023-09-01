package api

import (
	"strings"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"google.golang.org/protobuf/proto"
)

func TestCreateSiteRequest(t *testing.T) {
	(CaseList{
		{
			Name:     "domain is required",
			Message:  &v1.CreateSiteRequest{},
			Contains: "domain: value is required ",
		},
		{
			Name:     "reject invalid hostname",
			Message:  &v1.CreateSiteRequest{Domain: "https://vinceanalytics.github.com"},
			Contains: "domain: value must be a valid hostname ",
		},
		{
			Name:    "accept valid hostname",
			Pass:    true,
			Message: &v1.CreateSiteRequest{Domain: "vinceanalytics.github.com"},
		},
	}).Apply(t)
}

func TestCreateTokenRequest(t *testing.T) {
	(CaseList{
		{
			Name:     "name is required",
			Message:  &v1.CreateTokenRequest{Password: "xxx", Generate: true},
			Contains: "name: value is required",
		},
		{
			Name:     "password  is required",
			Message:  &v1.CreateTokenRequest{Name: "test-token", Generate: true},
			Contains: "password: value is required",
		},
		{
			Name:     "client side token missing token",
			Message:  &v1.CreateTokenRequest{Name: "test-token", Password: "xxx", PublicKey: []byte("xxx")},
			Contains: "token  is required ",
		},
		{
			Name:     "client side token missing public key",
			Message:  &v1.CreateTokenRequest{Name: "test-token", Password: "xxx", Token: "xxx"},
			Contains: "public_key  is required",
		},
	}).Apply(t)
}

type Case struct {
	Name     string
	Pass     bool
	Message  proto.Message
	Contains string
}

type CaseList []Case

func (ls CaseList) Apply(t *testing.T) {
	v, err := protovalidate.New()
	if err != nil {
		t.Fatal()
	}
	for i := range ls {
		if ls[i].Pass {
			validateOk(t, v, ls[i].Name, ls[i].Message)
		} else {
			validateFail(t, v, ls[i].Name, ls[i].Message, ls[i].Contains)
		}
	}
}

func validateFail(t *testing.T, v *protovalidate.Validator, name string, m proto.Message, contains string) {
	t.Helper()
	err := v.Validate(m)
	if err != nil {
		if !strings.Contains(err.Error(), contains) {
			t.Errorf("%s: expected %q got %q", name, contains, err.Error())
		}
		return
	}
	t.Errorf("%s: expected %q", name, contains)
}

func validateOk(t *testing.T, v *protovalidate.Validator, name string, m proto.Message) {
	t.Helper()
	err := v.Validate(m)
	if err != nil {
		t.Errorf("%s: %v", name, err)
	}
}
