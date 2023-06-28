package timeseries

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSlice(t *testing.T) {
	s := newSlice()
	t.Run("correct capacity when taken from pool", func(t *testing.T) {
		if got, want := cap(s.d), 1<<10; want != got {
			t.Errorf("expected %d got %d", want, got)
		}
	})

	t.Run("zero position when taken from pool", func(t *testing.T) {
		if got, want := s.pos, 0; want != got {
			t.Errorf("expected %d got %d", want, got)
		}
	})

	t.Run("writes u16", func(t *testing.T) {
		x := uint16(2)
		if got, want := s.u16(x), binary.BigEndian.AppendUint16(make([]byte, 0, 2), x); !bytes.Equal(got, want) {
			t.Errorf("expected %x got %x", want, got)
		}
		if got, want := s.pos, 2; want != got {
			t.Errorf("expected %d got %d", want, got)
		}
	})
}
