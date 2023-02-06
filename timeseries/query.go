package timeseries

import (
	"errors"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/parquet-go"
)

type Query struct {
	start           time.Time
	end             time.Time
	period          string
	interval        string
	sampleThreshold int
	filters         filterHandList
}

func QueryFrom(params url.Values) Query {
	interval := params.Get("interval")
	if interval == "" {
		interval = defaultIntervalForPeriod(params.Get("period"))
	}
	sampleThreshold := 20_000_000
	if x := params.Get("sample_threshold"); x != "" {
		n, _ := strconv.Atoi(x)
		if n != 0 {
			sampleThreshold = n
		}
	}
	switch params.Get("period") {
	case "realtime":
		date := today()
		return Query{
			period:          "realtime",
			interval:        interval,
			start:           date,
			end:             date,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "day":
		date := parseSingleDate(params.Get("date"))
		return Query{
			period:          "day",
			interval:        interval,
			start:           date,
			end:             date,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "7d":
		endDate := parseSingleDate(params.Get("date"))
		startDate := endDate.Truncate(24 * 6 * time.Hour)
		return Query{
			period:          "7d",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "30d":
		endDate := parseSingleDate(params.Get("date"))
		startDate := endDate.Truncate(24 * 30 * time.Hour)
		return Query{
			period:          "30d",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "month":
		date := parseSingleDate(params.Get("date"))
		startDate := beginningOfMonth(date)
		endDate := endOfMonth(date)
		return Query{
			period:          "month",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "6mo":
		endDate := endOfMonth(parseSingleDate(params.Get("date")))
		startDate := beginningOfMonth(endDate.AddDate(0, -5, 0))
		return Query{
			period:          "6mo",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "12mo":
		endDate := endOfMonth(parseSingleDate(params.Get("date")))
		startDate := beginningOfMonth(endDate.AddDate(0, -11, 0))
		return Query{
			period:          "12mo",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "year":
		endDate := endOfYear(parseSingleDate(params.Get("date")))
		startDate := beginningOfYear(endDate)
		return Query{
			period:          "year",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "custom":
		endDate := parseSingleDate(params.Get("to"))
		startDate := parseSingleDate(params.Get("from"))
		return Query{
			period:          "custom",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	}
	return Query{}
}

func beginningOfMonth(ts time.Time) time.Time {
	y, m, _ := ts.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, ts.Location())
}
func beginningOfYear(ts time.Time) time.Time {
	y, _, _ := ts.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, ts.Location())
}

func endOfMonth(ts time.Time) time.Time {
	return beginningOfMonth(ts).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func endOfYear(ts time.Time) time.Time {
	return beginningOfYear(ts).AddDate(1, 0, 0).Add(-time.Nanosecond)
}

func today() time.Time {
	return toDate(time.Now().UTC())
}

func parseSingleDate(date string) time.Time {
	if date == "today" || date == "" {
		return today()
	}
	ts, err := time.Parse(ISO8601, date)
	if err != nil {
		return today()
	}
	return toDate(ts)
}

func toDate(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(
		y, m, d, 0, 0, 0, 0, ts.Location(),
	)
}

func defaultIntervalForPeriod(period string) string {
	switch period {
	case "realtime":
		return "minute"
	case "day":
		return "hour"
	case "custom", "7d", "30d", "month":
		return "date"
	case "6mo", "12mo", "year":
		return "month"
	default:
		return ""
	}
}

type filterHand struct {
	field string
	h     matchFunc
}

type matchFunc func(o []bool, index int, rowGroup parquet.RowGroup, page parquet.Page) bool

func matchDictBasicMembers(name string, values []string) *filterHand {
	str := strings.Join(values, "")
	f := func(v parquet.Value) bool {
		return strings.Contains(str, v.String())
	}
	return &filterHand{
		field: name,
		h:     basicDictFilterMatch(f),
	}
}

func matchDictField(op filterOp, value string) func(parquet.Value) bool {
	x := parquet.ValueOf(value)
	var f func(parquet.Value) bool
	switch op {
	case filterEq:
		f = func(v parquet.Value) bool {
			return parquet.Equal(x, v)
		}
	case filterNeq:
		f = func(v parquet.Value) bool {
			return !parquet.Equal(x, v)
		}
	case filterWildEq:
		f = func(v parquet.Value) bool {
			ok, _ := path.Match(value, v.String())
			return ok
		}
	case filterWildNeq:
		f = func(v parquet.Value) bool {
			ok, _ := path.Match(value, v.String())
			return !ok
		}
	}
	return func(v parquet.Value) bool {
		if f != nil {
			return f(v)
		}
		return false
	}
}
func basicDictFilterMatch(match func(parquet.Value) bool) matchFunc {
	return func(o []bool, index int, rowGroup parquet.RowGroup, page parquet.Page) bool {
		dict := page.Dictionary()
		ok := false
		for i := 0; i < dict.Len(); i += 1 {
			if match(dict.Index(int32(i))) {
				ok = true
				break
			}
		}
		if !ok {
			// skip this page
			return false
		}
		values := make([]parquet.Value, page.NumValues())
		if _, err := page.Values().ReadValues(values); err != nil && !errors.Is(err, io.EOF) {
			panic("unexpected error while reading values " + err.Error())
		}
		for i := 0; i < int(page.NumValues()); i += 1 {
			o[i] = match(values[i])
		}
		return true
	}
}

func HasString(field, fieldValue string) *filterHand {
	return &filterHand{
		field: field,
		h: func(o []bool, index int, rowGroup parquet.RowGroup, page parquet.Page) bool {
			dict := page.Dictionary()
			value := parquet.ValueOf(fieldValue)
			ok := false
			for i := 0; i < dict.Len(); i += 1 {
				if parquet.Equal(value, dict.Index(int32(i))) {
					ok = true
					break
				}
			}
			if !ok {
				// skip this page
				return false
			}
			values := make([]parquet.Value, page.NumValues())
			if _, err := page.Values().ReadValues(values); err != nil && !errors.Is(err, io.EOF) {
				panic("unexpected error while reading values " + err.Error())
			}
			for i := 0; i < int(page.NumValues()); i += 1 {
				o[i] = parquet.Equal(values[i], value)
			}
			return true
		},
	}
}
