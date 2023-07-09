package spec

import (
	"encoding/json"
	"fmt"

	lo "github.com/polarsignals/frostdb/query/logicalplan"
)

type Op lo.Op

var op_names, op_keys = func() (map[string]Op, map[Op]string) {
	m := make(map[string]Op)
	s := make(map[Op]string)
	for i := lo.OpEq; i <= lo.OpOr; i++ {
		m[i.String()] = Op(i)
		s[Op(i)] = i.String()
	}
	return m, s
}()

func (o Op) MarshalJSON() ([]byte, error) {
	return json.Marshal(op_keys[o])
}

func (o *Op) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	n, ok := op_names[s]
	if !ok {
		return fmt.Errorf("unknown op %q", s)
	}
	*o = n
	return nil
}

var ValidSegment = map[string]struct{}{
	"browser":         {},
	"browser_version": {},
	"city":            {},
	"country_code":    {},
	"entry_page":      {},
	"exit_page":       {},
	"host":            {},
	"name":            {},
	"os":              {},
	"os_version":      {},
	"path":            {},
	"referrer":        {},
	"referrer_source": {},
	"region":          {},
	"screen":          {},
	"utm_campaign":    {},
	"utm_content":     {},
	"utm_medium":      {},
	"utm_source":      {},
	"utm_term":        {},
}

type Segment string

func (o Segment) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(o))
}

func (o *Segment) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	_, ok := ValidSegment[s]
	if !ok {
		return fmt.Errorf("unknown op %q", s)
	}
	*o = Segment(s)
	return nil
}

type Match struct {
	Segment Segment `json:"segment"`
	Op      Op      `json:"op"`
	Value   string  `json:"value"`
}

func (m *Match) Expr() *lo.BinaryExpr {
	return &lo.BinaryExpr{
		Left:  lo.Col(string(m.Segment)),
		Op:    lo.Op(m.Op),
		Right: lo.Literal(m.Value),
	}
}
