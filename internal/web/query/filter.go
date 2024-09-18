package query

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/vinceanalytics/vince/internal/ro2"
)

type Filter struct {
	Op    string
	Key   string
	Value []string
}

func (c *Filter) UnmarshalJSON(data []byte) error {
	var ls []any
	err := json.Unmarshal(data, &ls)
	if err != nil {
		return err
	}
	if len(ls) != 3 {
		return nil
	}
	// avoid panic for invalid casts
	var ok bool
	c.Op, ok = ls[0].(string)
	if !ok {
		return nil
	}
	c.Key, ok = ls[1].(string)
	if !ok {
		return nil
	}
	v, ok := ls[2].([]any)
	if !ok {
		return nil
	}
	c.Value = make([]string, len(v))
	for i := range v {
		c.Value[i], ok = v[i].(string)
		if !ok {
			return nil
		}
	}
	return nil
}

func (c *Filter) To(db *ro2.Store) ro2.Filter {
	if len(c.Value) == 0 {
		return ro2.Reject{}
	}
	if strings.HasPrefix(c.Key, "event:props:") {
		return ro2.Reject{}
	}
	if strings.HasPrefix(c.Key, "event:") {
		key := strings.TrimPrefix(c.Key, "event:")
		switch key {
		case "name", "page", "hostname":
			return build(db, c.Op, key, c.Value)
		default:
			return ro2.Reject{}
		}
	}
	if strings.HasPrefix(c.Key, "visit:") {
		key := strings.TrimPrefix(c.Key, "visit:")
		switch key {
		case "source",
			"referrer",
			"utm_medium",
			"utm_source",
			"utm_campaign",
			"utm_content",
			"utm_term",
			"screen",
			"device",
			"browser",
			"browser_version",
			"os",
			"os_version",
			"country",
			"region",
			"entry_page",
			"exit_page",
			"entry_page_hostname",
			"exit_page_hostname":
			return build(db, c.Op, key, c.Value)
		case "city":
			code, err := strconv.Atoi(c.Value[0])
			if err != nil {
				return ro2.Reject{}
			}
			return &ro2.EqInt{
				Field: uint64(db.Number(key)),
				Value: int64(code),
			}
		default:
			return ro2.Reject{}
		}
	}
	return ro2.Reject{}
}

func build(db *ro2.Store, op string, field string, value []string) ro2.Filter {
	f := uint64(db.Number(field))
	switch op {
	case "is":
		return ro2.NewEq(f, value[0])
	case "is_not":
		return ro2.Noop{}
	case "matches":
		return ro2.NewRe(f, value[0])
	case "does_not_match", "does_not_contain":
		return ro2.Noop{}
	case "contains":
		return ro2.NewRe(f, strings.Join(value, "|"))
	default:
		return ro2.Reject{}
	}
}
