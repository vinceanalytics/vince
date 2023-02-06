package timeseries

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Filters map[string]any

func parseFilters(f string) (o Filters) {
	o = make(Filters)
	if err := json.Unmarshal([]byte(f), &o); err != nil {
		parts := strings.Split(f, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			var kv []string
			var isNegated bool
			if strings.Contains(p, "==") {
				kv = strings.Split(p, "==")
			} else if strings.Contains(p, "!=") {
				kv = strings.Split(p, "!=")
				isNegated = true
			} else {
				continue
			}
			kv[0] = strings.TrimSpace(kv[0])
			kv[1] = strings.TrimSpace(kv[1])
			ls, isList := parseMemberList(kv[1])
			isWildcard := strings.Contains(kv[1], "*")
			finalValue := strings.ReplaceAll(kv[1], "\\|", "|")
			switch {
			case kv[0] == "event:goal":
				o[kv[0]] = parseGoal(finalValue)
			case isWildcard && isNegated:
				o[kv[0]] = &filterExpr{
					op:    filterWildNeq,
					value: kv[1],
				}
			case isWildcard:
				o[kv[0]] = &filterExpr{
					op:    filterWildEq,
					value: kv[1],
				}
			case isList:
				for i := 0; i < len(ls); i += 1 {
					ls[i] = strings.ReplaceAll(ls[i], "\\|", "|")
				}
				o[kv[0]] = ls
			case isNegated:
				o[kv[0]] = &filterExpr{
					op:    filterNeq,
					value: finalValue,
				}
			default:
				o[kv[0]] = &filterExpr{
					op:    filterEq,
					value: finalValue,
				}
			}
		}
	}
	return
}

func parseGoal(s string) *filterGoal {
	if strings.HasPrefix(s, "Visit ") {
		page := strings.TrimPrefix(s, "Visit ")
		return &filterGoal{
			key:   "page",
			value: page,
		}
	}
	return &filterGoal{
		key:   "event",
		value: s,
	}
}

func parseMemberList(s string) (ls []string, ok bool) {
	if !strings.Contains(s, "|") {
		return
	}
	var begin int
	for i, v := range s {
		if v == '|' && i > 0 && s[i-1] != '\\' {
			ls = append(ls, s[begin:i])
			begin = i + 1
		}
	}
	if len(ls) > 0 {
		ls = append(ls, s[begin:])
	}
	ok = len(ls) > 0
	return
}

func (f Filters) String() string {
	var s strings.Builder
	var keys []string
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i != 0 {
			s.WriteByte('\n')
		}
		s.WriteString(k)
		switch e := f[k].(type) {
		case *filterGoal:
			s.WriteString(e.String())
		case *filterExpr:
			switch e.op {
			case filterEq:
				s.WriteString(" is ")
			case filterNeq:
				s.WriteString(" is_not ")
			case filterWildEq:
				s.WriteString(" matches ")
			case filterWildNeq:
				s.WriteString(" matches_not ")
			}
			s.WriteString(e.value)
		case []string:
			s.WriteString(" [ ")
			for x, v := range e {
				if x != 0 {
					s.WriteString(", ")
				}
				s.WriteString(v)
			}
			s.WriteString(" ]")
		}
	}
	return s.String()
}

type filterHandList []*filterHand

func (f Filters) build() (ls filterHandList) {
	for key, v := range f {
		switch e := v.(type) {
		case *filterGoal:
			var name string
			switch e.key {
			case "event":
				name = "name"
			case "page":
				name = "path"
			default:
				return nil
			}
			op := filterEq
			if strings.Contains(e.value, "*") {
				op = filterWildEq
			}
			ls = append(ls, &filterHand{
				field: name,
				h: basicDictFilterMatch(
					matchDictField(op, e.value),
				),
			})
		case *filterExpr:
			ls = append(ls, &filterHand{
				field: key,
				h: basicDictFilterMatch(
					matchDictField(e.op, e.value),
				),
			})
		case []string:
			ls = append(ls, matchDictBasicMembers(key, e))
		}
	}
	for _, h := range ls {
		switch h.field {
		case "screen":
			h.field = "screen_size"
		case "os":
			h.field = "operating_system"
		case "country":
			h.field = "country_code"
		}
	}
	sort.Slice(ls, func(i, j int) bool {
		return ls[i].field < ls[j].field
	})
	return
}

type filterGoal struct {
	key   string
	value string
}

func (f *filterGoal) String() string {
	return fmt.Sprintf("[:is :%s %q]", f.key, f.value)
}

type filterExpr struct {
	op    filterOp
	value string
}

type filterOp uint

const (
	filterEq filterOp = iota
	filterNeq
	filterWildEq
	filterWildNeq
)
