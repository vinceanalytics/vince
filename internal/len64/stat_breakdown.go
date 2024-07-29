package len64

import (
	"encoding/binary"
	"time"
)

type Breakdown struct {
	Results []map[string]any `json:"results"`
}

func (db *Store) Breakdown(domain string, start,
	end time.Time, filter Filter,
	metrics []string, property string) (*Breakdown, error) {
	match, err := db.Select(start, end, domain, filter,
		append(metricsToProject(metrics), property),
	)
	if err != nil {
		return nil, err
	}
	groups := match.GroupBy(property)
	a := &Breakdown{
		Results: make([]map[string]any, 0, len(groups)),
	}
	key := make([]byte, 2+4)
	copy(key, trIDPrefix)

	for _, group := range groups {
		binary.BigEndian.PutUint32(key[2:], uint32(group.Value))
		var str string
		if value := db.cache.Get(nil, key[2:]); len(value) > 0 {
			str = string(value)
		} else {
			value, done, err := db.db.Get(key)
			if err == nil {
				done.Close()
				str = string(value)
			}
		}
		m := make(map[string]any)
		m[property] = str
		for _, k := range metrics {
			m[k] = group.Projection.Compute(k)
		}
		a.Results = append(a.Results, m)
	}
	return a, nil
}
