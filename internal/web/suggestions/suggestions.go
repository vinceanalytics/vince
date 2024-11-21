package suggestions

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func Suggest(db *db.Config, w http.ResponseWriter, r *http.Request) {
	property := models.Field_value[adjustProps(r.PathValue("filter_name"))]
	result := []Result{}
	switch property {
	case models.Field_page,
		models.Field_entry_page,
		models.Field_source,
		models.Field_os,
		models.Field_os_version,
		models.Field_device,
		models.Field_exit_page,
		models.Field_utm_source,
		models.Field_utm_medium,
		models.Field_utm_campaign,
		models.Field_utm_content,
		models.Field_utm_term,
		models.Field_referrer,
		models.Field_browser,
		models.Field_browser_version,
		models.Field_host:
		result = base(db, property, "")
	case models.Field_country:
		result = country(db, "")
	case models.Field_subdivision1_code:
		result = region(db, "")
	}
	db.JSON(w, result)
}

func country(db *db.Config, prefix string) (o []Result) {
	lo := db.Location()
	db.TimeSeries().SearchKeys(models.Field_country, []byte(prefix), func(key []byte) error {
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
	db.TimeSeries().SearchKeys(models.Field_subdivision1_code, []byte(prefix), func(key []byte) error {
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
