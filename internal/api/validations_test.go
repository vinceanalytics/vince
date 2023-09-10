package api

import (
	"strings"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"google.golang.org/protobuf/proto"
)

func TestCreateSiteRequest(t *testing.T) {
	(CaseList{
		{
			Name:     "domain is required",
			Message:  &sitesv1.CreateSiteRequest{},
			Contains: "domain: value is required ",
		},
		{
			Name:     "reject invalid hostname",
			Message:  &sitesv1.CreateSiteRequest{Domain: "https://vinceanalytics.github.com"},
			Contains: "domain: value must be a valid hostname ",
		},
		{
			Name:    "accept valid hostname",
			Pass:    true,
			Message: &sitesv1.CreateSiteRequest{Domain: "vinceanalytics.github.com"},
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
