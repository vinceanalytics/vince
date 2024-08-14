package ro

import "time"

func DateRange(from, to time.Time) []uint64 {
	from = toDate(from)
	to = toDate(to)
	if from.After(to) {
		return []uint64{}
	}
	var o []uint64
	for i := from; !i.After(to); i = i.AddDate(0, 0, 1) {
		o = append(o, uint64(i.UnixMilli()))
	}
	return o
}

func Today() time.Time {
	return toDate(time.Now())
}

func toDate(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}
