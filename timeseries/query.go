package timeseries

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/gernest/vince/log"
	"github.com/gernest/vince/timex"
	"github.com/segmentio/parquet-go"
)

type Query struct {
	start           time.Time
	end             time.Time
	period          string
	interval        string
	selected        []string
	filters         filterHandList
	sampleThreshold int
}

func (q Query) Select(fields ...string) Query {
	q.selected = append(q.selected, fields...)
	return q
}

func (q Query) Filter(field string, h MatchFunc) Query {
	q.filters = append(q.filters, &filterHand{
		field: field,
		h:     h,
	})
	return q
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
		date := timex.Today()
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
		startDate := timex.BeginningOfMonth(date)
		endDate := timex.EndOfMonth(date)
		return Query{
			period:          "month",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "6mo":
		endDate := timex.EndOfMonth(parseSingleDate(params.Get("date")))
		startDate := timex.BeginningOfMonth(endDate.AddDate(0, -5, 0))
		return Query{
			period:          "6mo",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "12mo":
		endDate := timex.EndOfMonth(parseSingleDate(params.Get("date")))
		startDate := timex.BeginningOfMonth(endDate.AddDate(0, -11, 0))
		return Query{
			period:          "12mo",
			interval:        interval,
			start:           startDate,
			end:             endDate,
			filters:         parseFilters(params.Get("filters")).build(),
			sampleThreshold: sampleThreshold,
		}
	case "year":
		endDate := timex.EndOfYear(parseSingleDate(params.Get("date")))
		startDate := timex.BeginningOfYear(endDate)
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

const ISO8601 = "2006-01-02"

func parseSingleDate(date string) time.Time {
	if date == "today" || date == "" {
		return timex.Today()
	}
	ts, err := time.Parse(ISO8601, date)
	if err != nil {
		return timex.Today()
	}
	return timex.Date(ts)
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
	h     MatchFunc
	field string
}

type MatchFunc func(ctx context.Context, values []parquet.Value, b []bool, page parquet.Page) bool

func Eq(value string) MatchFunc {
	x := parquet.ValueOf(value)
	return func(ctx context.Context, values []parquet.Value, b []bool, page parquet.Page) bool {
		dict := page.Dictionary()
		var ok bool
		for i := 0; i < dict.Len(); i += 1 {
			if ok = parquet.Equal(dict.Index(int32(i)), x); ok {
				break
			}
		}
		if !ok {
			return false
		}
		_, err := page.Values().ReadValues(values)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Get(ctx).Err(err).Msg("failed to get pages values")
			return false
		}
		for i := range values {
			b[i] = parquet.Equal(x, values[i])
		}
		return true
	}
}
