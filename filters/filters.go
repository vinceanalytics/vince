package filters

import (
	"github.com/blevesearch/vellum/regexp"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
)

var PropToProjection = map[v1.Property]v1.Filters_Projection{
	v1.Property_event:           v1.Filters_Event,
	v1.Property_page:            v1.Filters_Path,
	v1.Property_entry_page:      v1.Filters_EntryPage,
	v1.Property_exit_page:       v1.Filters_EntryPage,
	v1.Property_source:          v1.Filters_ReferrerSource,
	v1.Property_referrer:        v1.Filters_Referrer,
	v1.Property_utm_source:      v1.Filters_UtmSource,
	v1.Property_utm_medium:      v1.Filters_UtmMedium,
	v1.Property_utm_campaign:    v1.Filters_UtmCampaign,
	v1.Property_utm_content:     v1.Filters_UtmContent,
	v1.Property_utm_term:        v1.Filters_UtmTerm,
	v1.Property_device:          v1.Filters_Screen,
	v1.Property_browser:         v1.Filters_Browser,
	v1.Property_browser_version: v1.Filters_BrowserVersion,
	v1.Property_os:              v1.Filters_Os,
	v1.Property_os_version:      v1.Filters_OsVersion,
	v1.Property_country:         v1.Filters_Country,
	v1.Property_region:          v1.Filters_Region,
	v1.Property_domain:          v1.Filters_Domain,
	v1.Property_city:            v1.Filters_City,
}

type CompiledFilter struct {
	Column string
	Op     v1.Filter_OP
	Value  []byte
	Re     *regexp.Regexp
}

func CompileFilters(f *v1.Filters) ([]*CompiledFilter, error) {
	o := make([]*CompiledFilter, 0, len(f.List))
	for _, v := range f.List {
		x, err := compileFilter(v)
		if err != nil {
			return nil, err
		}
		o = append(o, x)
	}
	return o, nil
}

func compileFilter(f *v1.Filter) (*CompiledFilter, error) {
	o := &CompiledFilter{
		Column: PropToProjection[f.Property].String(),
		Op:     f.Op,
	}
	o.Value = []byte(f.Value)
	switch f.Op {
	case v1.Filter_re_equal, v1.Filter_re_not_equal:
		re, err := regexp.New(f.Value)
		if err != nil {
			return nil, err
		}
		o.Re = re
	}
	return o, nil
}
