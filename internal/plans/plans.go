package plans

import (
	_ "embed"
	"encoding/json"
	"sync"
)

type Plan struct {
	Limit            uint64
	MonthlyProductID uint64
	YearlyProductID  uint64
	MonthlyCost      float64
	YearlyCost       float64
}

//go:embed v1.json
var planV1 []byte

var plan = &sync.Map{}

func init() {
	var o []Plan
	err := json.Unmarshal(planV1, &o)
	if err != nil {
		panic("failed to decode plans " + err.Error())
	}
	for _, p := range o {
		plan.Store(p.MonthlyProductID, p)
		plan.Store(p.YearlyProductID, p)
	}
}

func GetPlan(id uint64) (Plan, bool) {
	p, ok := plan.Load(id)
	if !ok {
		return Plan{}, ok
	}
	return p.(Plan), true
}
