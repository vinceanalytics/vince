package query

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/vinceanalytics/vince/internal/oracle"
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

func (c *Filter) To() oracle.Filter {
	if len(c.Value) == 0 {
		return oracle.Reject()
	}
	if strings.HasPrefix(c.Key, "event:props:") {
		key := strings.TrimPrefix(c.Key, "event:props:")
		return build(c.Op, "props."+key, c.Value)
	}
	if strings.HasPrefix(c.Key, "event:") {
		key := strings.TrimPrefix(c.Key, "event:")
		switch key {
		case "name", "page", "hostname":
			return build(c.Op, key, c.Value)
		default:
			return oracle.Reject()
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
			return build(c.Op, key, c.Value)
		case "city":
			code, err := strconv.Atoi(c.Value[0])
			if err != nil {
				return oracle.Reject()
			}
			return oracle.NewEqInt(key, int64(code))
		default:
			return oracle.Reject()
		}
	}
	return oracle.Reject()
}

func build(op string, field string, value []string) oracle.Filter {
	switch op {
	case "is":
		return oracle.NewEq(field, value[0])
	case "is_not":
		return oracle.NewNeq(field, value[0])
	case "matches":
		return oracle.NewRe(field, value[0])
	case "does_not_match":
		return oracle.NewNre(field, value[0])
	case "contains":
		return oracle.NewRe(field, strings.Join(value, "|"))
	case "does_not_contain":
		return oracle.NewNre(field, strings.Join(value, "|"))
	default:
		return oracle.Reject()
	}
}
