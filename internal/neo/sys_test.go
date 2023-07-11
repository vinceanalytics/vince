package neo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/segmentio/parquet-go"
)

func TestSys(t *testing.T) {
	var b bytes.Buffer
	w := Writer[Sys](&b)
	ts := time.Now()
	start := ts.Add(-24 * time.Hour)

	var a []Sys

	for i := 0; i < 5; i++ {
		a = append(a, Sys{
			Timestamp: start.Add(time.Duration(i) * time.Hour),
			Labels: map[string]string{
				"code": strconv.Itoa(i),
			},
			Name:  "request",
			Value: float64(i),
		})
	}
	a[0].Labels["say"] = "hello"
	_, err := w.Write(a)
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	f, err := parquet.OpenFile(bytes.NewReader(b.Bytes()), int64(b.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range f.RowGroups() {
		scheme := g.Schema()
		cols := g.ColumnChunks()
		keys, _ := scheme.Lookup("labels", "key_value", "key")
		values, _ := scheme.Lookup("labels", "key_value", "value")
		x := read(cols[keys.ColumnIndex])
		y := read(cols[values.ColumnIndex])
		a := Collect(x, y)
		fmt.Println(a)
	}
	t.Error()
}

func read(col parquet.ColumnChunk) (o []parquet.Value) {
	println(col.BloomFilter() == nil)
	x := col.Pages()
	defer x.Close()
	for {
		p, err := x.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			panic(err)
		}
		v := make([]parquet.Value, p.NumValues())
		p.Values().ReadValues(v)
		o = append(o, v...)
	}
}

func Collect(a, b []parquet.Value) (o []map[string]string) {
	m := make(map[string]string)
	var idx int
	for i := range a {
		x := a[i]
		y := b[i]
		if x.RepetitionLevel() == 0 {
			if i != 0 {
				o = append(o, m)
				m = make(map[string]string)
				idx++
			}
		}
		m[x.String()] = y.String()
	}
	o = append(o, m)
	println(idx, len(o))
	return
}
