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

	b.Write(&v1.List{
		Items: []*v1.Data{
			{Bounce: nil},
			{Bounce: True},
			{Bounce: False},
			{Page: "/"},
		},
	})
	r := b.NewRecord()
	got, _ := r.MarshalJSON()
	// os.WriteFile("testdata/record.json", got, 0600)
	want, _ := os.ReadFile("testdata/record.json")
	require.Equal(t, string(want), string(got))
}
