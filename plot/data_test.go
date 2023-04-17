package plot

import (
	"bytes"
	"os"
	"testing"
)

func TestData(t *testing.T) {
	var b bytes.Buffer
	t.Run("zero data", func(t *testing.T) {
		var d Data
		err := d.Render(&b)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("calculate sum and percent", func(t *testing.T) {
		views := []float64{1, 2, 3, 5, 4}
		var d Data
		d.Set(Event, "pageviews", AggrValues{
			Views: views,
		}, AggregateOptions{
			NoSum:     true,
			NoPercent: true,
		})
		b.Reset()
		err := d.Render(&b)
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile("data.json", b.Bytes(), 0600)
	})
}
