package db

import (
	"maps"

	"github.com/cespare/xxhash/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	v2 "github.com/vinceanalytics/vince/gen/go/vince/v2"
)

type Batch struct {
	ts       []int64
	uid      []int64
	bounce   []bool
	session  []bool
	view     []bool
	duration []int64
	attr     []map[string]string
	hash     xxhash.Digest
	labels   map[string]string
}

func newBatch() *Batch {
	return &Batch{
		labels: make(map[string]string),
	}
}

func (b *Batch) Reset() {
	b.ts = b.ts[:0]
	b.uid = b.uid[:0]
	b.bounce = b.bounce[:0]
	b.session = b.session[:0]
	b.view = b.view[:0]
	b.duration = b.duration[:0]
	b.attr = b.attr[:0]
	clear(b.labels)
}

func (b *Batch) Append(e *v2.Data) {
	b.ts = append(b.ts, e.Timestamp)
	b.hash.Reset()
	b.hash.Write(e.Id)
	b.uid = append(b.uid, int64(b.hash.Sum64()))
	b.bounce = append(b.bounce, e.GetBounce())
	b.session = append(b.session, e.GetSession())
	b.view = append(b.view, e.GetView())
	b.duration = append(b.duration, e.Duration)

	clear(b.labels)
	b.prop(v1.Property_browser, e.Browser)
	b.prop(v1.Property_browser_version, e.BrowserVersion)
	b.prop(v1.Property_city, e.City)
	b.prop(v1.Property_country, e.Country)
	b.prop(v1.Property_device, e.Device)
	b.prop(v1.Property_domain, e.Domain)
	b.prop(v1.Property_entry_page, e.EntryPage)
	b.prop(v1.Property_event, e.Event)
	b.prop(v1.Property_exit_page, e.ExitPage)
	b.prop(v1.Property_browser, e.Host)
	b.prop(v1.Property_os, e.Os)
	b.prop(v1.Property_os_version, e.OsVersion)
	b.prop(v1.Property_page, e.Page)
	b.prop(v1.Property_referrer, e.Referrer)
	b.prop(v1.Property_region, e.Region)
	b.prop(v1.Property_source, e.Source)
	b.prop(v1.Property_utm_campaign, e.UtmCampaign)
	b.prop(v1.Property_utm_content, e.UtmContent)
	b.prop(v1.Property_utm_medium, e.UtmMedium)
	b.prop(v1.Property_utm_source, e.UtmSource)
	b.prop(v1.Property_utm_term, e.UtmTerm)
	b.prop(v1.Property_tenant_id, e.TenantId)
	b.attr = append(b.attr, maps.Clone(b.labels))
}

func (b *Batch) prop(k v1.Property, v string) {
	if v != "" {
		b.labels[k.String()] = v
	}
}
