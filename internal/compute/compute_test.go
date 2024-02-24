package compute

import (
	"testing"

	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
)

func TestBounceRate(t *testing.T) {
	b := array.NewBooleanBuilder(memory.NewGoAllocator())
	type Case struct {
		args []int
		want int
	}
	cases := []Case{
		{args: []int{1, -1, 1, -1, 0, 0}, want: 0},
		{args: []int{1, -1, -1, -1, 0, 0}, want: 0},
		{args: []int{1, 1, 1, -1, 0, 0}, want: 2},
	}
	for i, k := range cases {
		for _, v := range k.args {
			switch v {
			case 1:
				b.Append(true)
			case 0:
				b.Append(false)
			case -1:
				b.AppendNull()
			}
		}

		require.Equal(t, k.want, CalcBounce(b.NewBooleanArray()), "case", i)
	}
}
