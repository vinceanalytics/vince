package timex

import (
	"testing"
	"time"
)

func TestY(t *testing.T) {
	now := time.Now()
	r := Range{
		From: now,
		To:   now.Add(24 * time.Hour),
	}
	ls := r.Timestamps()
	t.Error(len(ls))
	t.Error(ls[len(ls)-1], r.To.Truncate(time.Hour).Unix())
}
