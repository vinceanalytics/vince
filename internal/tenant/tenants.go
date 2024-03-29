package tenant

import (
	"context"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

const Default = "staples"

type Loader interface {
	TenantBySiteID(ctx context.Context, siteId string) (tenantId string)
}

func Config(o *v1.Config, domains []string) *v1.Config {
	if len(domains) == 0 {
		return o
	}
	t := &v1.Tenant{
		Id: Default,
	}
	for _, d := range domains {
		t.Domains = append(t.Domains, &v1.Domain{
			Name: d,
		})
	}
	o.Tenants = append(o.Tenants, t)
	if o.Credentials == nil {
		o.Credentials = &v1.Credential_List{}
	}
	// The tenant we configure is the super user. We assign all permissions to
	// them.
	o.Credentials.Items = append(o.Credentials.Items, &v1.Credential{
		Username: t.Id,
		Password: o.AuthToken,
		Perms:    []v1.Credential_Permission{v1.Credential_ALL},
	})
	return o
}

type Tenants struct {
	domains map[string]*v1.Tenant
	id      map[string]*v1.Tenant
	all     []*v1.Tenant
}

func NewTenants(o *v1.Config) *Tenants {
	t := &Tenants{
		domains: make(map[string]*v1.Tenant),
		id:      make(map[string]*v1.Tenant),
		all:     o.Tenants,
	}
	for _, v := range o.Tenants {
		t.id[v.Id] = v
		for _, d := range v.Domains {
			t.domains[d.Name] = v
		}
	}
	return t
}

func (t *Tenants) Get(domain string) *v1.Tenant {
	return t.domains[domain]
}

func (t *Tenants) GetByID(id string) *v1.Tenant {
	return t.id[id]
}

func (t *Tenants) TenantBySiteID(ctx context.Context, siteId string) (tenantId string) {
	return t.domains[siteId].GetId()
}

func (t *Tenants) Domains(id string) []*v1.Domain {
	n, ok := t.id[id]
	if ok {
		return n.Domains
	}
	return nil
}

func (t *Tenants) AllDomains() (o []*v1.Domain) {
	for _, v := range t.id {
		o = append(o, v.Domains...)
	}
	return
}

func (t *Tenants) All() []*v1.Tenant {
	return t.all
}
