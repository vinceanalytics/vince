package query

import (
	"io"

	"github.com/gernest/rows"
)

type RowIterator interface {
	io.Seeker
	Next() (*rows.Row, uint64, *int64, bool, error)
}

// IDs is a query return type for just uint64 row ids.
// It should only be used internally (since RowIdentifiers
// is the external return type), but it is exported because
// the proto package needs access to it.
type IDs []uint64

func (r IDs) Merge(other IDs, limit int) IDs {
	i, j := 0, 0
	result := make(IDs, 0)
	for i < len(r) && j < len(other) && len(result) < limit {
		av, bv := r[i], other[j]
		if av < bv {
			result = append(result, av)
			i++
		} else if av > bv {
			result = append(result, bv)
			j++
		} else {
			result = append(result, bv)
			i++
			j++
		}
	}
	for i < len(r) && len(result) < limit {
		result = append(result, r[i])
		i++
	}
	for j < len(other) && len(result) < limit {
		result = append(result, other[j])
		j++
	}
	return result
}
