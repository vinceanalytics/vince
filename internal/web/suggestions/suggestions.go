package suggestions

import (
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func Suggest(db *db.Config, w http.ResponseWriter, r *http.Request) {
	property := v1.Field(v1.Field_value[adjustProps(r.PathValue("filter_name"))])
	result := []Result{}
	switch property {
	case v1.Field_page,
		v1.Field_entry_page,
		v1.Field_source,
		v1.Field_os,
		v1.Field_os_version,
		v1.Field_device,
		v1.Field_exit_page,
		v1.Field_utm_source,
		v1.Field_utm_medium,
		v1.Field_utm_campaign,
		v1.Field_utm_content,
		v1.Field_utm_term,
		v1.Field_referrer,
		v1.Field_browser,
		v1.Field_browser_version,
		v1.Field_host:
		result = base(db, property, "")
	case v1.Field_country:
		result = country(db, "")
	case v1.Field_subdivision1_code:
		result = region(db, "")
	}
	db.JSON(w, result)
}

func country(db *db.Config, prefix string) (o []Result) {
	lo := db.Location()
	db.TimeSeries().SearchKeys(v1.Field_country, []byte(prefix), func(key []byte) error {
		code := string(key)
		name := lo.GetCountryName(code)
		o = append(o, Result{
			Label: name,
			Value: code,
		})
		return nil
	})
	return
}

func region(db *db.Config, prefix string) (o []Result) {
	lo := db.Location()
	db.TimeSeries().SearchKeys(v1.Field_subdivision1_code, []byte(prefix), func(key []byte) error {
		code := string(key)
		name := lo.GetRegionName(key)
		o = append(o, Result{
			Label: name,
			Value: code,
		})
		return nil
	})
	return
}

func adjustProps(name string) string {
	switch name {
	case "region":
		return "subdivision1_code"
	case "screen", "screen_size":
		return "device"
	case "operating_system":
		return "os"
	case "operating_system_version":
		return "os_version"
	case "hostname":
		return "host"
	default:
		return name
	}
}

type Result struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func base(db *db.Config, field models.Field, prefix string) (o []Result) {
	o = make([]Result, 0, 16)
	db.TimeSeries().SearchKeys(field, []byte(prefix), func(key []byte) error {
		o = append(o, Result{
			Label: string(key),
			Value: string(key),
		})
		return nil
	})
	return
}
