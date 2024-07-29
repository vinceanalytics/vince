package len64

import "time"

type Timeseries struct {
	Results []map[string]any `json:"results"`
}

func (db *Store) Timeseries(domain string, start, end time.Time, filter Filter, metrics []string) (*Timeseries, error) {
	match, err := db.Select(start, end, domain, filter,
		append(metricsToProject(metrics), "date"),
	)
	if err != nil {
		return nil, err
	}
	groups := match.GroupBy("date")
	a := &Timeseries{
		Results: make([]map[string]any, 0, len(groups)),
	}
	for _, group := range groups {
		m := make(map[string]any)
		m["date"] = time.UnixMilli(group.Value).UTC().Format("2006-01-02")
		for _, k := range metrics {
			m[k] = group.Projection.Compute(k)
		}
		a.Results = append(a.Results, m)
	}
	return a, nil
}
