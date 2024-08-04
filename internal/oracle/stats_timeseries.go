package oracle

type Timeseries struct {
	Results []map[string]any `json:"results"`
}

func (o *Oracle) Timeseries(start, end int64, domain string, filter Filter, metrics []string) (*Timeseries, error) {
	b, err := o.Breakdown(start, end, domain, filter, metrics, "date")
	if err != nil {
		return nil, err
	}
	return &Timeseries{Results: b.Results}, nil
}
