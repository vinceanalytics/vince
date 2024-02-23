package events

import (
	"os"
	"testing"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/events/v1"
)

func TestBuild(t *testing.T) {
	b := New(memory.NewGoAllocator())
	defer b.Release()

	r := b.Write(&v1.List{
		Items: []*v1.Data{
			{Bounce: nil},
			{Bounce: True},
			{Bounce: False},
			{Page: "/"},
		},
	})
	got, _ := r.MarshalJSON()
	// os.WriteFile("testdata/record.json", got, 0600)
	want, _ := os.ReadFile("testdata/record.json")
	require.JSONEq(t, string(want), string(got))
	gotSchema := r.Schema().String()
	// os.WriteFile("testdata/schema.txt", []byte(gotSchema), 0600)
	wantSchema, _ := os.ReadFile("testdata/schema.txt")
	require.Equal(t, string(wantSchema), gotSchema)
}
