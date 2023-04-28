package plans

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

type Plan struct {
	Limit            uint64
	MonthlyProductID uint64
	YearlyProductID  uint64
	MonthlyCost      float64
	YearlyCost       float64
}

func (p Plan) Before() float64 {
	return p.MonthlyCost * 12
}

func (p Plan) Format() string {
	n := p.Limit
	switch {
	case n >= 1_000 && n < 1_000_000:
		thousands := (n / 100) / 10
		return fmt.Sprintf("%dK", thousands)
	case n >= 1_000_000 && n < 1_000_000_000:
		millions := (n / 100_000) / 10
		return fmt.Sprintf("%dM", millions)
	case n >= 1_000_000_000 && n < 1_000_000_000_000:
		billions := (n / 100_000_000) / 10
		return fmt.Sprintf("%dB", billions)
	default:
		return strconv.FormatUint(n, 10)
	}
}

//go:embed v1.json
var planV1 []byte

var plan = &sync.Map{}

var All []Plan

func init() {
	err := json.Unmarshal(planV1, &All)
	if err != nil {
		panic("failed to decode plans " + err.Error())
	}
	for _, p := range All {
		plan.Store(p.MonthlyProductID, p)
		plan.Store(p.YearlyProductID, p)
	}
}

func GetPlan(id uint64) Plan {
	p, ok := plan.Load(id)
	if !ok {
		return Plan{}
	}
	return p.(Plan)
}

func SubscriptionInterval(ctx context.Context, pid uint64) string {
	for _, p := range All {
		if p.MonthlyProductID == pid {
			return "monthly"
		}
		if p.YearlyProductID == pid {
			return "yearly"
		}
	}
	return ""
}
