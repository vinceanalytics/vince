package tenant

import (
	"context"
	"net/url"

	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
)

const Default = "staples"

func Config(o *v1.Config, domains []string) *v1.Config {
	t := &v1.Tenant{
		Id: Default,
	}
	for _, d := range domains {
		t.Domains = append(t.Domains, &v1.Domain{
			Name: d,
		})
	}
	o.Tenants = append(o.Tenants, t)
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

func (t *Tenants) Load(ctx context.Context, q url.Values) context.Context {
	v := q.Get("tenant_id")
	if v == "" {
		site := q.Get("site_id")
		if site != "" {
			s := t.Get(site)
			if s != nil {
				v = s.Id
			}
		}
	}
	return With(ctx, v)
}

func (t *Tenants) All() []*v1.Tenant {
	return t.all
}

type tenantId struct{}

func With(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, tenantId{}, id)
}

func Get(ctx context.Context) string {
	return ctx.Value(tenantId{}).(string)
}
