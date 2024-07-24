package query

import "encoding/json"

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
