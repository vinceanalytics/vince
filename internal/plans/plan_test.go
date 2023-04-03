package plans

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
)

func TestGeneratePlans(t *testing.T) {
	t.Skip("use this only to generate new plans")
	o := make([]Plan, 8)
	for i := range o {
		o[i].MonthlyProductID = rand.Uint64()
		o[i].YearlyProductID = rand.Uint64()
	}
	{
		o[0].Limit = 10000
		o[0].MonthlyCost = 9
		o[0].YearlyCost = 90
	}
	{
		o[1].Limit = 100000
		o[1].MonthlyCost = 19
		o[1].YearlyCost = 190
	}
	{
		o[2].Limit = 200000
		o[2].MonthlyCost = 29
		o[2].YearlyCost = 290
	}
	{
		o[3].Limit = 500000
		o[3].MonthlyCost = 49
		o[3].YearlyCost = 490
	}
	{
		o[4].Limit = 1000000
		o[4].MonthlyCost = 69
		o[4].YearlyCost = 690
	}
	{
		o[5].Limit = 2000000
		o[5].MonthlyCost = 89
		o[5].YearlyCost = 890
	}
	{
		o[6].Limit = 5000000
		o[6].MonthlyCost = 129
		o[6].YearlyCost = 1290
	}
	{
		o[7].Limit = 10000000
		o[7].MonthlyCost = 169
		o[7].YearlyCost = 1690
	}
	b, _ := json.Marshal(o)
	os.WriteFile("v1.json", b, 0600)
}
