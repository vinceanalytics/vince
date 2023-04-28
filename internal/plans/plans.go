package plans

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/pkg/log"
)

const enterprizeUsage = 10_000_000

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

var Enterprize = Plan{
	MonthlyProductID: 2023,
	YearlyProductID:  2023,
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

func (p Plan) IsEnterprize() bool {
	return p.MonthlyProductID == Enterprize.MonthlyProductID
}

func SuggestPlan(ctx context.Context, u *models.User, usage uint64) Plan {
	switch {
	case usage > enterprizeUsage:
		return Enterprize
	case u.IsEnterprize(ctx):
		return Enterprize
	default:
		for _, p := range All {
			if usage < p.Limit {
				return p
			}
		}
		return Plan{}
	}
}

func SubscriptionInterval(ctx context.Context, sub *models.Subscription) string {
	o := GetPlan(sub.PlanID)
	if o.MonthlyProductID == 0 {
		e := sub.GetEnterPrise(ctx)
		if e != nil {
			return e.BillingInterval
		}
		return ""
	}
	for _, p := range All {
		if p.MonthlyProductID == sub.PlanID {
			return "monthly"
		}
		if p.YearlyProductID == sub.PlanID {
			return "yearly"
		}
	}
	return ""
}

func Allowance(ctx context.Context, sub *models.Subscription) uint64 {
	o := GetPlan(sub.PlanID)
	if o.MonthlyProductID != 0 {
		return o.Limit
	}
	if e := sub.GetEnterPrise(ctx); e != nil {
		return e.MonthlyPageViewLimit
	}
	log.Get(ctx).Debug().Uint64("plan_id", sub.PlanID).Msg("Unknown allowance")
	return 0
}
