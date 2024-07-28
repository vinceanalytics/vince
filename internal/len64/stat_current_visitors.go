package len64

import (
	"time"
)

func (db *Store) CurrentVisitor(domain string, duration time.Duration) (uint64, error) {
	if duration == 0 {
		duration = 5 * time.Minute
	}
	end := time.Now().UTC()
	start := end.
		Add(-duration).
		Truncate(time.Second)
	match, err := db.Select(start, end, domain, NoopFilter{}, []string{"uid"})
	if err != nil {
		return 0, err
	}
	return match.Visitors(), nil
}
