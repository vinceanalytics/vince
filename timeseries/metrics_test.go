package timeseries

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/segmentio/parquet-go"
)

type Slice struct {
	Visitors uint64
	Visits   uint64
}

func TestMet(t *testing.T) {

	s := parquet.SchemaOf(&Metrics{})
	for _, path := range s.Columns() {
		leaf, _ := s.Lookup(path...)
		fmt.Printf("%d => %q\n", leaf.ColumnIndex, strings.Join(path, "."))
	}
	var b bytes.Buffer
	{
		key, values := lookup(PROPS_NAME, METRIC_TYPE_page_views)
		fmt.Printf("=> %d \n", key.ColumnIndex)
		for _, v := range values {
			fmt.Printf("  - %d \n", v.ColumnIndex)

		}

	}
	w := parquet.NewGenericWriter[*Metrics](&b)
	w.Write([]*Metrics{
		{
			Name: map[string]Total{
				"pageviews": {
					Visitors: 10,
				},
				"custom": {
					Visitors: 2,
				},
			},
		},
	})
	w.Close()

	f, err := parquet.OpenFile(bytes.NewReader(b.Bytes()), int64(b.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range f.RowGroups() {
		ch := g.ColumnChunks()
		{
			println("==========")
			pages := ch[7].Pages()
			p, _ := pages.ReadPage()
			v := make([]parquet.Value, p.NumValues())
			p.Values().ReadValues(v)
			for _, x := range v {
				println(x.String(), x.DefinitionLevel(), x.RepetitionLevel())
			}
			pages.Close()
		}
		{
			println("==========")
			pages := ch[8].Pages()
			p, _ := pages.ReadPage()
			v := make([]parquet.Value, p.NumValues())
			p.Values().ReadValues(v)
			for _, x := range v {
				println(x.String(), x.DefinitionLevel(), x.RepetitionLevel())
			}
			pages.Close()
		}
	}
	t.Error()
}

func TestS(t *testing.T) {
	type U struct {
		a *int
	}

	n := func(x **int) {
		if *x == nil {
			var v int
			*x = &v
		}
	}
	u := &U{}
	n(&u.a)

	t.Error(*u.a == 0)
}
