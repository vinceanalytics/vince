package query

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Filters []*Filter

func (f Filters) Translate() Filters {
	o := make(Filters, 0, len(f))
	for i := range f {
		n := f[i].To()
		if n != nil {
			o = append(o, n)
		}
	}
	return o
}

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
	c.Value = make([]string, 0, len(v))
	for i := range v {
		switch e := v[i].(type) {
		case string:
			c.Value = append(c.Value, e)
		case float64:
			c.Value = append(c.Value, strconv.Itoa(int(e)))
		default:
			return nil
		}
	}
	return nil
}

func (c *Filter) To() *Filter {
	if len(c.Value) == 0 {
		return nil
	}
	if strings.HasPrefix(c.Key, "event:props:") {
		return nil
	}
	if strings.HasPrefix(c.Key, "event:") {
		key := strings.TrimPrefix(c.Key, "event:")
		switch key {
		case "name", "page", "hostname":
			return &Filter{Op: c.Op, Key: key, Value: c.Value}
		case "goal":
			if strings.HasPrefix(c.Value[0], "Visit ") {
				return &Filter{Op: c.Op, Key: "page", Value: []string{
					strings.TrimPrefix(c.Value[0], "Visit "),
				}}
			}
			return &Filter{Op: c.Op, Key: "name", Value: c.Value}
		default:
			return nil
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
			"exit_page_hostname", "city":
			return &Filter{Op: c.Op, Key: key, Value: c.Value}
		default:
			return nil
		}
	}
	return nil
}
